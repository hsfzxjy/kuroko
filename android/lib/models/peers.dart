// Dart imports:
import 'dart:async';

// Flutter imports:
import 'package:flutter/foundation.dart';

// Project imports:
import 'package:kuroko/core/db.dart';
import 'package:kuroko/util/async.dart';
import './bridges.dart';

class Peer {
  Peer(this.peerid, this.name);

  final int peerid;
  final String name;
  List<Bridge>? bridges;

  static Future<List<Peer>> getList() async {
    final bridges = await Bridge.getList();
    Peer? peer;
    final results = <Peer>[];
    for (var bridge in bridges) {
      if (peer == null || peer.peerid != bridge.peerid) {
        peer = Peer(bridge.peerid!, bridge.name!);
        peer.bridges = [bridge];
        results.add(peer);
      } else {
        peer.bridges!.add(bridge);
      }
    }
    return results;
  }

  static Future<bool> exists(int peerId) async {
    final db = await openDb();
    final ret = await db.query('peers',
        columns: ['peerid'], where: 'peerid = ?', whereArgs: [peerId]);
    return ret.isNotEmpty;
  }
}

class PeerList extends ChangeNotifier {
  PeerList._();
  static final PeerList i = PeerList._();

  final Poller _poller = Poller();
  bool _disposed = false;

  Future<void> refetch() {
    _poller.poll();
    return _poller.done;
  }

  Stream<List<Peer>> get list async* {
    while (!_disposed) {
      yield await Peer.getList();
      await _poller.awake();
    }
  }

  @override
  void dispose() {
    _disposed = true;
    _poller.poll();
    super.dispose();
  }
}
