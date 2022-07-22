library kcore_android;

// Dart imports:
import 'dart:async';
import 'dart:collection';
import 'dart:convert';
import 'dart:ffi';
import 'dart:isolate';
import 'dart:math';
import 'dart:typed_data';

// Flutter imports:
import 'package:flutter/services.dart';

// Package imports:
import 'package:dart_zoned_task/dart_zoned_task.dart';
import 'package:permission_handler/permission_handler.dart';

// Project imports:
import 'package:kcore_android/libkcore_generated_binding.dart';

part './callback_manager.dart';
part './accepter.dart';
part './bluetooth.dart';
part './session.dart';
part './types.dart';
part './util.dart';

class Kcore {
  static const _channel = MethodChannel('kcore');
  static final dll = DynamicLibrary.open('libkcore.so');
  static final lib = LibKcore(dll);
  static late final ReceivePort _receivePort;

  static Future<void> init() async {
    _receivePort = ReceivePort();
    _CB.bindPort(_receivePort);
    lib.InitializeDartFFI(
      NativeApi.initializeApiDLData,
      _receivePort.sendPort.nativePort,
    );
    await _channel.invokeMethod('initDLL');
    Accepter.initSessions();
  }

  static Future<void> ensurePermission() async {
    final perms = [
      Permission.location,
      Permission.bluetooth,
      Permission.bluetoothScan,
      Permission.bluetoothConnect,
      Permission.bluetoothAdvertise,
    ];
    for (final p in perms) {
      final st = await p.status;
      if (st.isPermanentlyDenied) {
        if (!await openAppSettings()) throw 'cannot grant permission';
        final st = await p.status;
        if (!st.isGranted) throw 'cannot grant permission';
      } else if (!st.isGranted) {
        final st = await p.request();
        if (!st.isGranted) throw 'cannot grant permission';
      }
    }
  }
}
