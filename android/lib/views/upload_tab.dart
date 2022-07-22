// Dart imports:
import 'dart:async';
import 'dart:io';

// Flutter imports:
import 'package:flutter/material.dart';

// Package imports:
import 'package:async/async.dart';
import 'package:dart_zoned_task/dart_zoned_task.dart';
import 'package:provider/provider.dart';

// Project imports:
import 'package:kuroko/endpoints/out.dart';
import 'package:kuroko/models/bridges.dart';
import 'package:kuroko/states.dart';
import 'package:kuroko/util/async.dart';

enum FileState {
  waiting,
  uploading,
  finished,
  canceled,
  canceling,
  errored,
}

class UploadFile extends ChangeNotifier {
  UploadFile({required String filename, required this.group, Bridge? bridge})
      : _bridge = bridge,
        file = File(filename) {
    start();
  }

  File file;
  UploadFileGroup group;
  Bridge? _bridge;
  Bridge get bridge => _bridge ?? group.bridge;
  set bridge(Bridge b) => _bridge = b;

  final state = MemoizedStreamController(initial: FileState.waiting);
  final _perc = StreamController<double>.broadcast();
  late final stateOrPerc =
      StreamGroup.mergeBroadcast([_perc.stream, state.stream]);

  ZonedTask<double>? _task;

  void start() {
    state.add(FileState.waiting);
    _task = SendFile(bridge, file).task
      ..stream.asBroadcastStream().listen(
        (perc) {
          if (state.current != FileState.uploading) {
            state.add(FileState.uploading);
          }
          _perc.add(perc);
        },
        onDone: () {
          if (state.current == FileState.uploading) {
            state.add(FileState.finished);
          }
          _task = null;
        },
        onError: (e) {
          final newState =
              e is CanceledError ? FileState.canceled : FileState.errored;
          state.add(newState);
          _task = null;
        },
      );
  }

  void cancel() {
    state.add(FileState.canceling);
    _task?.cancel().whenComplete(() => state.add(FileState.canceled));
  }

  void onTap() {
    switch (state.current) {
      case FileState.waiting:
      case FileState.uploading:
        cancel();
        break;
      case FileState.finished:
      case FileState.canceling:
        break;
      case FileState.canceled:
      case FileState.errored:
        start();
        break;
    }
  }
}

class UploadFileGroup extends ChangeNotifier {
  UploadFileGroup({required this.bridge, required List<String> filenames})
      : assert(!bridge.isAnonymous) {
    files = filenames.map((f) => UploadFile(filename: f, group: this)).toList();
  }
  late final List<UploadFile> files;
  Bridge bridge;
}

class PendingUploadFiles extends ChangeNotifier {
  PendingUploadFiles._();
  static final i = PendingUploadFiles._();

  List<String>? _filenames;
  bool get isPending => _filenames != null;

  void replace(List<String>? fns) {
    if (fns?.isEmpty ?? true) fns = null;
    _filenames = fns;
    notifyListeners();
  }

  void resolve(Bridge bridge) {
    if (_filenames == null) return;
    UploadManager.i.addGroup(
      UploadFileGroup(bridge: bridge, filenames: _filenames!),
    );
    replace(null);
    TabIndex.i.toUploadTab();
  }
}

class UploadManager extends ChangeNotifier {
  UploadManager._();
  static final i = UploadManager._();

  final groups = MemoizedStreamController(initial: <UploadFileGroup>[]);

  void addGroup(UploadFileGroup group) {
    groups.current.add(group);
    groups.add(groups.current);
  }
}

class UploadTab extends StatelessWidget {
  const UploadTab({super.key});

  @override
  Widget build(BuildContext context) {
    return Consumer<List<UploadFileGroup>>(builder: (context, groups, _) {
      return ListView.builder(
        itemCount: groups.length,
        itemBuilder: (context, index) => UploadGroupCard(
          group: groups.elementAt(groups.length - 1 - index),
        ),
      );
    });
  }
}

class UploadGroupCard extends StatelessWidget {
  final UploadFileGroup group;
  const UploadGroupCard({super.key, required this.group});

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Column(
        children: group.files.map((file) {
          return UploadFileItem(file: file);
        }).toList(),
      ),
    );
  }
}

class UploadFileItem extends StatelessWidget {
  final UploadFile file;
  const UploadFileItem({super.key, required this.file});
  @override
  Widget build(BuildContext context) {
    return ListTile(
      leading: Text(file.file.uri.pathSegments.last),
      trailing: StreamBuilder(
        stream: file.stateOrPerc,
        initialData: file.state.current,
        builder: (ctx, snap) => Text(snap.data.toString()),
      ),
      onTap: () {
        file.onTap();
      },
    );
  }
}
