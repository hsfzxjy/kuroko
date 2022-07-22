// Dart imports:
import 'dart:async';
import 'dart:convert';
import 'dart:developer';
import 'dart:io';

// Package imports:
import 'package:dart_zoned_task/dart_zoned_task.dart';
import 'package:logging/logging.dart';

// Project imports:
import 'package:kuroko/core/bridge_session.dart';
import 'package:kuroko/endpoints/consts.dart';
import 'package:kuroko/models/bridges.dart';
import 'package:kuroko/util/info.dart';

abstract class BaseRemoteCall<T> {
  int getOpcode();
  bool getIdentityRequired() => false;

  Future<void> _writeHeader(BridgeSession sess) async {
    final magicBuffer = utf8.encode(MAGIC);
    await sess.writeIntList(magicBuffer);
    await sess.writeU32(getOpcode());
    if (getIdentityRequired()) {
      await sess.writeUint8List(await MachineInfo.instance.uid);
    }
  }
}

abstract class BaseStreamRemoteCall<T> extends BaseRemoteCall {
  Stream<T> get stream => task.stream;
  late final ZonedTask<T> task;

  BaseStreamRemoteCall(Bridge bridge) {
    task = ZonedTask.fromStream<T>(() async* {
      final sess = await BridgeSession.dial(bridge);
      try {
        await _writeHeader(sess);
        yield* _call(sess);
      } finally {
        await sess.close();
      }
    });
  }

  Stream<T> _call(BridgeSession sess);
}

abstract class BaseFutureRemoteCall<T> extends BaseRemoteCall {
  Future<T> get future => task.future;
  late final ZonedTask<T> task;

  BaseFutureRemoteCall(Bridge bridge) {
    task = ZonedTask.fromFuture<T>(() async {
      final sess = await BridgeSession.dial(bridge);
      try {
        await _writeHeader(sess);
        return await _call(sess);
      } finally {
        await sess.close();
      }
    });
  }

  Future<T> _call(BridgeSession sess);
}

class SendFile extends BaseStreamRemoteCall<double> {
  SendFile(Bridge bridge, this.file) : super(bridge);
  final File file;
  @override
  int getOpcode() => 0;

  @override
  bool getIdentityRequired() => true;

  @override
  Stream<double> _call(BridgeSession sess) async* {
    log('Sending ${file.path}', level: Level.INFO.value);
    yield 0;

    await sess.writeString(file.uri.pathSegments.last);
    final stat = await file.stat();
    final fileSize = stat.size;
    await sess.writeU64(fileSize);

    if (fileSize > 0) {
      double nbytesRead = 0;
      await for (var data in file.openRead()) {
        await sess.writeIntList(data);
        nbytesRead += data.length;
        yield nbytesRead / fileSize;
      }
    }

    yield 1;
  }
}

class ExchangeIdentity extends BaseFutureRemoteCall<void> {
  ExchangeIdentity(Bridge bridge) : super(bridge);

  @override
  int getOpcode() => 1;

  @override
  Future<void> _call(BridgeSession sess) async {
    final info = MachineInfo.instance;

    await sess.writeUint8List(await info.uid);
    await sess.writeString(await info.name);

    final peerId = await sess.readUint(6);
    final peerName = await sess.readString();

    await Bridge.ensureExist(
      peerid: peerId,
      name: peerName,
      type: sess.bridge.type,
      address: sess.bridge.address,
    );

    await sess.bridge.fill(peerId);
  }
}
