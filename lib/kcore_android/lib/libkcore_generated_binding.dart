// AUTO GENERATED FILE, DO NOT EDIT.
//
// Generated by `package:ffigen`.
import 'dart:ffi' as ffi;

class LibKcore {
  /// Holds the symbol lookup function.
  final ffi.Pointer<T> Function<T extends ffi.NativeType>(String symbolName)
      _lookup;

  /// The symbols are looked up in [dynamicLibrary].
  LibKcore(ffi.DynamicLibrary dynamicLibrary) : _lookup = dynamicLibrary.lookup;

  /// The symbols are looked up with [lookup].
  LibKcore.fromLookup(
      ffi.Pointer<T> Function<T extends ffi.NativeType>(String symbolName)
          lookup)
      : _lookup = lookup;

  int InitializeDartFFI(
    ffi.Pointer<ffi.Void> arg0,
    int arg1,
  ) {
    return _InitializeDartFFI(
      arg0,
      arg1,
    );
  }

  late final _InitializeDartFFIPtr = _lookup<
      ffi.NativeFunction<
          ffi.IntPtr Function(
              ffi.Pointer<ffi.Void>, ffi.Int64)>>('InitializeDartFFI');
  late final _InitializeDartFFI = _InitializeDartFFIPtr.asFunction<
      int Function(ffi.Pointer<ffi.Void>, int)>();

  void CFree(
    ffi.Pointer<ffi.Void> arg0,
  ) {
    return _CFree(
      arg0,
    );
  }

  late final _CFreePtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Pointer<ffi.Void>)>>(
          'CFree');
  late final _CFree =
      _CFreePtr.asFunction<void Function(ffi.Pointer<ffi.Void>)>();

  void KCore_AccepterSetStateCallback(
    int ctt,
    int ccb,
  ) {
    return _KCore_AccepterSetStateCallback(
      ctt,
      ccb,
    );
  }

  late final _KCore_AccepterSetStateCallbackPtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint8, DartCallback)>>(
          'KCore_AccepterSetStateCallback');
  late final _KCore_AccepterSetStateCallback =
      _KCore_AccepterSetStateCallbackPtr.asFunction<void Function(int, int)>();

  void KCore_AccepterStart(
    int ctt,
  ) {
    return _KCore_AccepterStart(
      ctt,
    );
  }

  late final _KCore_AccepterStartPtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint8)>>(
          'KCore_AccepterStart');
  late final _KCore_AccepterStart =
      _KCore_AccepterStartPtr.asFunction<void Function(int)>();

  void KCore_AccepterStop(
    int ctt,
  ) {
    return _KCore_AccepterStop(
      ctt,
    );
  }

  late final _KCore_AccepterStopPtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint8)>>(
          'KCore_AccepterStop');
  late final _KCore_AccepterStop =
      _KCore_AccepterStopPtr.asFunction<void Function(int)>();

  void KCore_AccepterAccept(
    int ccb,
  ) {
    return _KCore_AccepterAccept(
      ccb,
    );
  }

  late final _KCore_AccepterAcceptPtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(DartCallback)>>(
          'KCore_AccepterAccept');
  late final _KCore_AccepterAccept =
      _KCore_AccepterAcceptPtr.asFunction<void Function(int)>();

  void KCore_SessionRelease(
    int extrap,
  ) {
    return _KCore_SessionRelease(
      extrap,
    );
  }

  late final _KCore_SessionReleasePtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint64)>>(
          'KCore_SessionRelease');
  late final _KCore_SessionRelease =
      _KCore_SessionReleasePtr.asFunction<void Function(int)>();

  void KCore_CancelDial(
    int token,
  ) {
    return _KCore_CancelDial(
      token,
    );
  }

  late final _KCore_CancelDialPtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint32)>>(
          'KCore_CancelDial');
  late final _KCore_CancelDial =
      _KCore_CancelDialPtr.asFunction<void Function(int)>();

  int KCore_SessionDial(
    int x1,
    int x2,
    int x3,
    int ccb,
  ) {
    return _KCore_SessionDial(
      x1,
      x2,
      x3,
      ccb,
    );
  }

  late final _KCore_SessionDialPtr = _lookup<
      ffi.NativeFunction<
          ffi.Uint32 Function(ffi.Uint64, ffi.Uint64, ffi.Uint64,
              DartCallback)>>('KCore_SessionDial');
  late final _KCore_SessionDial =
      _KCore_SessionDialPtr.asFunction<int Function(int, int, int, int)>();

  void KCore_SessionRead(
    int csid,
    ffi.Pointer<ffi.Uint8> buf_ptr,
    int buf_size,
    int ccb,
  ) {
    return _KCore_SessionRead(
      csid,
      buf_ptr,
      buf_size,
      ccb,
    );
  }

  late final _KCore_SessionReadPtr = _lookup<
      ffi.NativeFunction<
          ffi.Void Function(ffi.Uint32, ffi.Pointer<ffi.Uint8>, ffi.Uint64,
              DartCallback)>>('KCore_SessionRead');
  late final _KCore_SessionRead = _KCore_SessionReadPtr.asFunction<
      void Function(int, ffi.Pointer<ffi.Uint8>, int, int)>();

  void KCore_SessionWrite(
    int csid,
    ffi.Pointer<ffi.Uint8> buf_ptr,
    int buf_size,
    int ccb,
  ) {
    return _KCore_SessionWrite(
      csid,
      buf_ptr,
      buf_size,
      ccb,
    );
  }

  late final _KCore_SessionWritePtr = _lookup<
      ffi.NativeFunction<
          ffi.Void Function(ffi.Uint32, ffi.Pointer<ffi.Uint8>, ffi.Uint64,
              DartCallback)>>('KCore_SessionWrite');
  late final _KCore_SessionWrite = _KCore_SessionWritePtr.asFunction<
      void Function(int, ffi.Pointer<ffi.Uint8>, int, int)>();

  void KCore_SessionClose(
    int csid,
    int ccb,
  ) {
    return _KCore_SessionClose(
      csid,
      ccb,
    );
  }

  late final _KCore_SessionClosePtr =
      _lookup<ffi.NativeFunction<ffi.Void Function(ffi.Uint32, DartCallback)>>(
          'KCore_SessionClose');
  late final _KCore_SessionClose =
      _KCore_SessionClosePtr.asFunction<void Function(int, int)>();

  int KCore_SessionGetTransportKey(
    int csid,
    ffi.Pointer<ffi.Uint8> buf_ptr,
  ) {
    return _KCore_SessionGetTransportKey(
      csid,
      buf_ptr,
    );
  }

  late final _KCore_SessionGetTransportKeyPtr = _lookup<
      ffi.NativeFunction<
          ffi.Int32 Function(ffi.Uint32,
              ffi.Pointer<ffi.Uint8>)>>('KCore_SessionGetTransportKey');
  late final _KCore_SessionGetTransportKey = _KCore_SessionGetTransportKeyPtr
      .asFunction<int Function(int, ffi.Pointer<ffi.Uint8>)>();
}

class GoInterface extends ffi.Struct {
  external ffi.Pointer<ffi.Void> t;

  external ffi.Pointer<ffi.Void> v;
}

class GoSlice extends ffi.Struct {
  external ffi.Pointer<ffi.Void> data;

  @ffi.Int64()
  external int len;

  @ffi.Int64()
  external int cap;
}

typedef DartCallback = ffi.Uint32;
