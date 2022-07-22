part of kcore_android;

const _SESSION_BUFSIZE = 1024;

final _FreeExtraNativeFunction =
    NativeFinalizer(Kcore.dll.lookup('KCore_SessionRelease'));

class SessionExtra implements Finalizable {
  late Pointer<Void> extra;
  late final Pointer<Uint8> read;
  late final Pointer<Uint8> write;
  late Uint8List readList;
  late Uint8List writeList;

  SessionExtra(this.extra, this.read, this.write)
      : readList = read.asTypedList(_SESSION_BUFSIZE),
        writeList = write.asTypedList(_SESSION_BUFSIZE) {
    _FreeExtraNativeFunction.attach(this, extra);
  }
}

class Session {
  final int sid;
  final SessionExtra extra;
  late TransportKey tk;
  bool _closed = false;

  Session(this.sid, this.tk, this.extra);

  static FutureTask<Session> dial(TransportKey tk) {
    final comp = Completer<List>();
    final cb = _CB.putCompleter(comp);

    return Task.future<Session, int>(
      (_) => comp.future.then(
        (arr) => Session(
          arr[0],
          tk,
          SessionExtra(
            Pointer<Void>.fromAddress(arr[1]),
            Pointer<Uint8>.fromAddress(arr[2]),
            Pointer<Uint8>.fromAddress(arr[3]),
          ),
        ),
      ),
      onCancel: (ctx) async => Kcore.lib.KCore_CancelDial(ctx.value!),
      context: Kcore.lib.KCore_SessionDial(tk.x1, tk.x2, tk.x3, cb),
    );
  }

  Future<Uint8List> _read(int n) {
    assert(n <= _SESSION_BUFSIZE);
    final comp = Completer<int>();
    final cb = _CB.putCompleter(comp);
    Kcore.lib.KCore_SessionRead(sid, extra.read, n, cb);
    return comp.future.then((_) => extra.readList.sublist(0, n));
  }

  Future<int> _write(Uint8List data) {
    final n = data.lengthInBytes;
    assert(n <= _SESSION_BUFSIZE);
    extra.writeList.setRange(0, n, data); // TODO: copy-free
    final comp = Completer<int>();
    final cb = _CB.putCompleter(comp);
    Kcore.lib.KCore_SessionWrite(sid, extra.write, n, cb);
    return comp.future;
  }

  Future<void> _close() {
    if (_closed) return Future.value();
    _closed = true;
    final comp = Completer<void>();
    final cb = _CB.putCompleter(comp);
    Kcore.lib.KCore_SessionClose(sid, cb);
    return comp.future;
  }
}

abstract class SessionProvider {
  Session get session;
}

mixin SessionMixin on SessionProvider {
  Future<void> close() => session._close();

  Future<void> writeUint8List(Uint8List list) async {
    var start = 0;
    while (start < list.length) {
      final end = min(start + _SESSION_BUFSIZE, list.length);
      await session._write(list.sublist(start, end));
      start = end;
    }
  }

  Future<Uint8List> readUint8List(int n) async {
    if (n <= _SESSION_BUFSIZE) {
      return session._read(n);
    } else {
      final retList = Uint8List(n);
      var start = 0;
      while (start < n) {
        final end = min(start + _SESSION_BUFSIZE, n);
        final chunk = await session._read(end - start);
        retList.setRange(start, end, chunk);
        start = end;
      }
      return retList;
    }
  }

  Future<void> writeU32(int x) {
    final buf = Uint8List(4)..buffer.asByteData().setUint32(0, x, Endian.big);
    return writeUint8List(buf);
  }

  Future<void> writeU64(int x) {
    final buf = Uint8List(8)..buffer.asByteData().setUint64(0, x, Endian.big);
    return writeUint8List(buf);
  }

  Future<void> writeIntList(List<int> list) {
    final buf = Uint8List.fromList(list);
    return writeUint8List(buf);
  }

  Future<void> writeString(String s) async {
    final buf = Uint8List.fromList(utf8.encode(s));
    await writeU32(buf.length);
    await writeUint8List(buf);
  }

  Future<int> readUint(int n) async {
    final buf = await readUint8List(n);
    var res = 0;
    for (var x in buf) {
      res = (res << 8) + x;
    }
    return res;
  }

  Future<int> readU32() async {
    final buf = await readUint8List(4);
    return ByteData.view(buf.buffer).getUint32(buf.offsetInBytes, Endian.big);
  }

  Future<int> readU64() async {
    final buf = await readUint8List(8);
    return ByteData.view(buf.buffer).getUint64(buf.offsetInBytes, Endian.big);
  }

  Future<String> readString() async {
    final length = await readU32();
    final buf = await readUint8List(length);
    return utf8.decode(buf);
  }
}
