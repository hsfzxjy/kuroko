part of kcore_android;

const _TK_LEN = 24;

enum TransportType implements EnumLike<int> {
  bluetooth(1);

  static final _lookup = _enumLookupMap<int, TransportType>(values);

  const TransportType(this.value);
  factory TransportType.of(int value) => _lookup[value]!;

  @override
  final int value;
}

class TransportKey {
  final Uint8List _list;

  TransportType? _type;
  TransportType get type {
    _type ??= TransportType.of(_list[0]);
    return _type!;
  }

  String? _address;
  String get address {
    if (_address != null) return _address!;
    switch (type) {
      case TransportType.bluetooth:
        _address = _list
            .sublist(1, 7)
            .map((x) => x.toRadixString(16).padLeft(2, '0'))
            .join(':')
            .toUpperCase();
        break;
      default:
        throw 'unknown TransportType: $_type';
    }
    return _address!;
  }

  int get x1 => _list.buffer.asByteData().getUint64(0, Endian.big);
  int get x2 => _list.buffer.asByteData().getUint64(8, Endian.big);
  int get x3 => _list.buffer.asByteData().getUint64(16, Endian.big);

  TransportKey(Uint8List? list) : _list = list ?? Uint8List(_TK_LEN);

  TransportKey.of(TransportType type, String address)
      : _list = Uint8List(_TK_LEN) {
    _list[0] = type.value;
    switch (type) {
      case TransportType.bluetooth:
        _list.setRange(
          1,
          7,
          address.split(':').map((x) => int.parse(x, radix: 16)),
        );
        break;
    }
  }

  TransportKey.fromTriple(int x1, int x2, int x3) : _list = Uint8List(_TK_LEN) {
    _list.buffer.asByteData()
      ..setUint64(0, x1, Endian.big)
      ..setUint64(8, x2, Endian.big)
      ..setUint64(16, x3, Endian.big);
  }
}
