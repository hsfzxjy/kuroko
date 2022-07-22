// Dart imports:
import 'dart:convert';
import 'dart:typed_data';

// Package imports:
import 'package:crypto/crypto.dart';
import 'package:device_info_plus/device_info_plus.dart';

class MachineInfo {
  MachineInfo._();

  static final instance = MachineInfo._();

  String? _name;
  Uint8List? _uid;

  Future<Uint8List> get uid async {
    await _resolve();
    return _uid!;
  }

  Future<String> get name async {
    await _resolve();
    return _name!;
  }

  Future<void> _resolve() async {
    if (_name != null && _uid != null) return;
    final info = await DeviceInfoPlugin().androidInfo;
    _name = info.model ?? 'Android Device';
    final fingerprint = utf8.encode(info.fingerprint!);
    final uid = sha256.convert(fingerprint).bytes.getRange(0, 6);
    _uid = Uint8List.fromList(uid.toList());
  }
}
