// Dart imports:
import 'dart:async';

// Flutter imports:
import 'package:flutter/material.dart';

// Package imports:
import 'package:dart_zoned_task/dart_zoned_task.dart';
import 'package:kcore_android/kcore.dart';
import 'package:provider/provider.dart';

// Project imports:
import 'package:kuroko/endpoints/out.dart';
import 'package:kuroko/models/bridges.dart';
import 'package:kuroko/util/async.dart';
import 'package:kuroko/util/queue.dart';

enum _DeviceState {
  unbonded,
  bonding,
  bonded,
  verifying,
  verified,
  unverified,
  cancelling,
  canceled,
  waiting,
  retrying;

  factory _DeviceState.from(BondState state) {
    if (state == BondState.bonded) {
      return bonded;
    } else if (state == BondState.bonding) {
      return bonding;
    } else {
      return unbonded;
    }
  }

  Color color() {
    MaterialColor color;
    switch (this) {
      case unbonded:
        return Colors.white;
      case bonding:
      case waiting:
        color = Colors.grey;
        break;
      case bonded:
        color = Colors.cyan;
        break;
      case cancelling:
      case retrying:
      case verifying:
        color = Colors.amber;
        break;
      case verified:
        color = Colors.green;
        break;
      case unverified:
      case canceled:
        color = Colors.red;
        break;
    }
    return color.shade600;
  }

  String displayName() => this == unbonded ? '' : toString().split('.').last;
}

class _BluetoothDevice extends ChangeNotifier implements QueuedItem {
  _BluetoothDevice(this.device, this.manager)
      : bridge = Bridge.fromBluetoothDevice(device) {
    bridge.fill(null).then((success) {
      if (success) {
        state.add(_DeviceState.verified);
      }
    });
    _updateBondState();
  }

  final BluetoothManager manager;
  final BluetoothDevice device;
  final Bridge bridge;

  String get displayName => bridge.name ?? device.name ?? '<UNKNOWN>';

  final state = MemoizedStreamController(initial: _DeviceState.unbonded);

  Future<void> _updateBondState() async {
    final value = await device.updateBondState();
    state.add(_DeviceState.from(value));
  }

  ZonedTask<void>? _linkTask;
  void onTap() => manager.deviceQueue.toggle(this);

  @override
  void onQueuedItemEnqueued() {
    state.add(_DeviceState.waiting);
  }

  @override
  Future<void> onQueuedItemCanceled(bool started) {
    Future<void> _update() async {
      if (bridge.isAnonymous) {
        await _updateBondState();
      } else {
        state.add(_DeviceState.verified);
      }
    }

    if (started) {
      state.add(_DeviceState.cancelling);
      return _linkTask!.cancel().whenComplete(() {
        state.add(_DeviceState.canceled);
        _linkTask = null;
        _update();
      });
    } else {
      state.add(_DeviceState.canceled);
      _update();
      return Future.value();
    }
  }

  @override
  void onQueuedItemScheduled(void Function() onFinished) {
    assert(_linkTask == null);
    _linkTask = ZonedTask.fromFuture(() => _link());
    _linkTask!.future.whenComplete(() {
      _linkTask = null;
      onFinished();
    });
  }

  Future<bool> _bond() async {
    return Task.future<bool, void>((ctx) async {
      try {
        return await device.startBond();
      } catch (_) {
        ctx.checkCanceled();
        return false;
      }
    }, onCancel: (_) async {
      await device.cancelBond();
    }).call();
  }

  Future<void> _link() async {
    var nRetries = 5;
    await _updateBondState();

    if (state.current != _DeviceState.bonded) {
      await Task.future((ctx) async {
        while (nRetries >= 0) {
          ctx.checkCanceled();
          nRetries -= 1;
          state.add(_DeviceState.bonding);

          if (await _bond()) break;

          state.add(_DeviceState.retrying);
          await Task.sleep(const Duration(seconds: 1)).call();
        }
      }).call();

      await _updateBondState();
      return;
    }
    state.add(_DeviceState.verifying);
    try {
      await ExchangeIdentity(bridge).future;
      state.add(_DeviceState.verified);
    } catch (e) {
      state.add(_DeviceState.unverified);
    }
  }
}

typedef Devices = List<_BluetoothDevice>;

enum _DiscoveryState {
  notStarted,
  started,
  ending,
  ended,
  notOpened;

  bool get isStart => this == started;
  bool get isEnd => this == ending || this == ended;
}

class BluetoothManager extends ChangeNotifier {
  StreamSubscription<BluetoothDevice>? _discoverySubscription;

  final devices = MemoizedStreamController<Devices>(initial: []);
  final state = MemoizedStreamController(initial: _DiscoveryState.notStarted);
  final deviceQueue = ExclusiveQueue();

  Future<void> ensureOpened() async {
    if (await BluetoothAdapter.ensureEnabled()) {
      await stopDiscovery();
      devices.add([]);
    } else {
      state.add(_DiscoveryState.notOpened);
      throw '';
    }
  }

  void startDiscovery() {
    if (state.current == _DiscoveryState.started) return;
    devices.add([]);
    ensureOpened()
        .whenComplete(() => deviceQueue.clear())
        .whenComplete(() async {
      await stopDiscovery();

      state.add(_DiscoveryState.started);
      _discoverySubscription = BluetoothDiscovery.start().listen(
        (device) {
          devices.current.add(_BluetoothDevice(device, this));
          devices.add(devices.current);
        },
        onDone: () {
          state.add(_DiscoveryState.ended);
          _discoverySubscription = null;
        },
      );
    });
  }

  Future<void> stopDiscovery({bool notify = true}) async {
    if (!state.current.isStart) return;
    if (notify) {
      state.add(_DiscoveryState.ending);
    }
    await _discoverySubscription?.cancel();
    await BluetoothDiscovery.stop();
    if (notify) {
      state.add(_DiscoveryState.ended);
    }
  }

  @override
  void dispose() {
    deviceQueue.clear();
    stopDiscovery(notify: false);
    super.dispose();
  }
}

class FindBluetoothDeviceScreen extends StatefulWidget {
  const FindBluetoothDeviceScreen({super.key});

  @override
  createState() => FindBluetoothDeviceScreenState();
}

class FindBluetoothDeviceScreenState extends State<FindBluetoothDeviceScreen> {
  final manager = BluetoothManager();

  @override
  void initState() {
    manager.state.stream.listen((state) {
      if (state == _DiscoveryState.notOpened) {
        Navigator.pop(context);
      }
    });
    manager.startDiscovery();
    super.initState();
  }

  @override
  void deactivate() {
    manager.stopDiscovery();
    super.deactivate();
  }

  @override
  void dispose() {
    manager.dispose();
    super.dispose();
  }

  Widget _buildFab() {
    return StreamBuilder<_DiscoveryState>(
      stream: manager.state.stream,
      initialData: manager.state.current,
      builder: (context, snap) {
        final state = snap.data!;
        if (state == _DiscoveryState.started) {
          return FloatingActionButton.extended(
            onPressed: () => manager.stopDiscovery(),
            label: const Text('Stop'),
            icon: const Icon(Icons.stop),
            backgroundColor: Colors.red.shade600,
          );
        } else if (state == _DiscoveryState.ended) {
          return FloatingActionButton.extended(
            onPressed: () => manager.startDiscovery(),
            label: const Text('Rescan'),
            icon: const Icon(Icons.refresh_outlined),
            backgroundColor: Colors.green.shade600,
          );
        } else {
          return FloatingActionButton.extended(
            onPressed: () {},
            label: const Text('Waiting'),
            icon: const Icon(Icons.do_not_disturb_alt_sharp),
            backgroundColor: Colors.orange.shade600,
          );
        }
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    final Widget body = StreamBuilder<Devices>(
      stream: manager.devices.stream,
      initialData: manager.devices.current,
      builder: (context, snap) {
        final devices = snap.data!;
        final length = devices.length;
        return ListView.builder(
          itemCount: length + 1,
          itemBuilder: (context, i) => i < length
              ? ChangeNotifierProvider.value(
                  value: devices[i],
                  child: const BluetoothDeviceListTile(),
                )
              : const ListTile(),
        );
      },
    );

    return ChangeNotifierProvider.value(
      value: manager,
      child: Scaffold(
        appBar: AppBar(title: const Text('Choose')),
        body: body,
        floatingActionButton: _buildFab(),
      ),
    );
  }
}

class BluetoothDeviceListTile extends StatelessWidget {
  const BluetoothDeviceListTile({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<_BluetoothDevice>(builder: _build);
  }

  Widget _build(BuildContext context, _BluetoothDevice device, Widget? _) {
    return ListTile(
      title: Text(device.displayName),
      trailing: StreamBuilder<_DeviceState>(
        stream: device.state.stream,
        initialData: device.state.current,
        builder: (context, snap) {
          final state = snap.data!;
          return Text(
            state.displayName().toUpperCase(),
            style: TextStyle(
              color: state.color(),
              fontWeight: FontWeight.bold,
            ),
          );
        },
      ),
      subtitle: Text(device.device.address),
      onTap: () {
        device.onTap();
      },
    );
  }
}
