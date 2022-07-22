part of kcore_android;

enum AccepterState implements EnumLike<int> {
  started(0),
  starting(1),
  error(2),
  ended(3),
  ending(4);

  static final Map<int, AccepterState> _lookup = _enumLookupMap(values);

  const AccepterState(this.value);
  factory AccepterState.of(int value) => _lookup[value]!;

  @override
  final int value;

  bool get isStart => this == started || this == starting;
  bool get isEnd => this == ended || this == ending || this == error;
  bool get isTransient => this == ending || this == starting;
}

class Accepter {
  final TransportType _type;

  final _stateCon = StreamController<AccepterState>();
  Stream<AccepterState> get state => _stateCon.stream;

  Accepter._(this._type) {
    final ccb =
        _CB.put((int cstate) => _stateCon.add(AccepterState.of(cstate)));
    Kcore.lib.KCore_AccepterSetStateCallback(_type.value, ccb);
  }

  void stop() {
    Kcore.lib.KCore_AccepterStop(_type.value);
  }

  void start() {
    Kcore.lib.KCore_AccepterStart(_type.value);
  }

  static late final StreamController<Session> _sessionCon;
  static Stream<Session> get sessions => _sessionCon.stream;
  static StreamController<Session> initSessions() {
    final con = _sessionCon = StreamController<Session>();
    final ccb = _CB.put(
      (List args) {
        final sess = Session(
          args[0],
          TransportKey.fromTriple(args[1], args[2], args[3]),
          SessionExtra(
            Pointer.fromAddress(args[4]),
            Pointer.fromAddress(args[5]),
            Pointer.fromAddress(args[6]),
          ),
        );
        con.add(sess);
      },
    );
    Kcore.lib.KCore_AccepterAccept(ccb);
    return con;
  }
}

final accepters = <TransportType, Accepter>{
  TransportType.bluetooth: Accepter._(TransportType.bluetooth)
};
