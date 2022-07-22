// Dart imports:
import 'dart:async';

// Package imports:
import 'package:kcore_android/kcore.dart';

// Project imports:
import 'package:kuroko/models/bridges.dart';

class BridgeSession extends SessionProvider
    with SessionMixin
    implements BridgeLike {
  @override
  final Session session;

  final Bridge bridge;

  @override
  String get address => session.tk.address;
  @override
  TransportType get type => session.tk.type;

  BridgeSession(this.session, {BridgeLike? bridge})
      : bridge = bridge != null
            ? Bridge.fromBridgeLike(bridge)
            : Bridge(session.tk.type, session.tk.address);

  static Future<BridgeSession> dial(BridgeLike b) async {
    final tk = TransportKey.of(b.type, b.address);
    final sess = await Session.dial(tk).call();
    return BridgeSession(sess, bridge: b);
  }
}
