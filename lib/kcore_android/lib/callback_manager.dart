part of kcore_android;

const _INT32_MAX = (1 << 31) - 1;
const _INT32_WRAPPING = 1 << 32;

const _CALL_ARRAY = 1;
const _CALL_WITH_CODE = 2;
const _CALL_MULTI = 4;
const _CALL_FUT_REJECTED = 8;

class _CallbackManager {
  static int counter = 0;
  late final ReceivePort port;

  final map = <int, Function>{};

  _CallbackManager();

  int put(Function cb) {
    while (map.containsKey(counter)) {
      counter++;
      if (counter > _INT32_MAX) counter -= _INT32_WRAPPING;
    }
    map[counter] = cb;
    return counter;
  }

  int putCompleter(Completer comp) => put((int code, res) {
        if (code & _CALL_FUT_REJECTED != 0) {
          comp.completeError(res);
        } else {
          comp.complete(res);
        }
      });

  void bindPort(ReceivePort port) {
    this.port = port;
    port.listen((args) {
      invoke(args);
    });
  }

  void invoke(List args) {
    final int cbid = args[0];
    final int code = args[1];

    late Function cb;
    if (code & _CALL_MULTI != 0) {
      cb = map[cbid]!;
    } else {
      cb = map.remove(cbid)!;
    }
    if (code & _CALL_ARRAY != 0) {
      if (code & _CALL_WITH_CODE != 0) {
        Function.apply(cb, [code, args.sublist(2)]);
      } else {
        Function.apply(cb, [args.sublist(2)]);
      }
    } else {
      if (code & _CALL_WITH_CODE != 0) {
        Function.apply(cb, args.sublist(1));
      } else {
        Function.apply(cb, args.sublist(2));
      }
    }
  }
}

final _CB = _CallbackManager();
