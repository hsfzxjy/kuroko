// Dart imports:
import 'dart:convert';
import 'dart:developer';

// Package imports:
import 'package:logging/logging.dart';

// Project imports:
import 'package:kuroko/core/bridge_session.dart';
import 'package:kuroko/endpoints/consts.dart';
import 'package:kuroko/models/bridges.dart';
import 'package:kuroko/models/peers.dart';
import 'package:kuroko/util/info.dart';

Future<void> routeEndpoint(BridgeSession sess) async {
  try {
    final magic = await sess.readUint8List(MAGIC.length);
    if (utf8.decode(magic) != MAGIC) {
      throw 'bad magic';
    }
    final opcode = await sess.readU32();
    final Future<void> Function(BridgeSession) op;
    switch (opcode) {
      case 0x0000:
        op = recieveFile;
        break;
      case 0x0001:
        op = exchangeIdentity;
        break;
      default:
        throw 'unknown opcode: $opcode';
    }
    await op(sess);
  } catch (e, st) {
    log('error in routeEndpoint(): $e', stackTrace: st);
  } finally {
    await sess.close();
  }
}

Future<void> recieveFile(BridgeSession sess) async {
  final peerId = await sess.readUint(6);
  if (!await Peer.exists(peerId)) {
    throw 'unknown peer';
  }
  final filename = await sess.readString();
  final ctlen = await sess.readU64();
  await sess.readUint8List(ctlen);
  log('Recieved: $filename, $ctlen bytes', level: Level.INFO.value);
}

Future<void> exchangeIdentity(BridgeSession sess) async {
  final info = MachineInfo.instance;

  final peerId = await sess.readUint(6);
  final peerName = await sess.readString();

  log('Paired with id: $peerId, name: $peerName', level: Level.INFO.value);

  await sess.writeUint8List(await info.uid);
  await sess.writeString(await info.name);

  await Bridge.ensureExist(
    peerid: peerId,
    name: peerName,
    type: sess.bridge.type,
    address: sess.bridge.address,
  );

  await sess.bridge.fill(peerId);
}
