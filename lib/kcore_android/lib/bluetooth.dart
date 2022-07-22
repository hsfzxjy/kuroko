part of kcore_android;

enum BluetoothDeviceType implements EnumLike<int> {
  unknown(0),
  classic(1),
  le(2),
  dual(3);

  static final _lookup = _enumLookupMap<int, BluetoothDeviceType>(values);

  const BluetoothDeviceType(this.value);
  factory BluetoothDeviceType.of(int value) => _lookup[value]!;

  @override
  final int value;
}

enum BondState implements EnumLike<int> {
  none(10),
  bonding(11),
  bonded(12);

  static final _lookup = _enumLookupMap<int, BondState>(values);

  const BondState(this.value);
  factory BondState.of(int value) => _lookup[value]!;

  @override
  final int value;
}

const _btMethodChannel = MethodChannel('kcore/bluetooth');

class BluetoothDevice {
  final BluetoothDeviceType type;
  final String address;
  BondState bondState;
  final String? name;

  BluetoothDevice._fromMap(Map m)
      : type = BluetoothDeviceType.of(m['type']),
        bondState = BondState.of(m['bondState']),
        address = m['address'],
        name = m['name'];

  Future<bool> startBond() =>
      _btMethodChannel.invokeMethodNonNull('startBond', {'address': address});

  Future<bool> cancelBond() =>
      _btMethodChannel.invokeMethodNonNull('stopBond', {'address': address});

  Future<BondState> updateBondState() async {
    final int result = await _btMethodChannel
        .invokeMethodNonNull('getBondState', {'address': address});
    bondState = BondState.of(result);
    return bondState;
  }
}

class BluetoothDiscovery {
  static const _channel = EventChannel('kcore/bluetooth/discovery');

  static Stream<BluetoothDevice> start() async* {
    await Kcore.ensurePermission();

    late StreamSubscription sub;
    final con = StreamController<BluetoothDevice>(
      onCancel: () {
        return sub.cancel();
      },
    );

    await _btMethodChannel.invokeMethod('startDiscovery');

    sub = _channel.receiveBroadcastStream().listen(
          (m) => con.add(BluetoothDevice._fromMap(m)),
          onError: con.addError,
          onDone: con.close,
        );

    yield* con.stream;
  }

  static Future<void> stop() => _btMethodChannel.invokeMethod('stopDiscovery');
}

class BluetoothAdapter {
  static Future<bool> ensureEnabled() =>
      _btMethodChannel.invokeMethodNonNull('ensureEnabled');
}
