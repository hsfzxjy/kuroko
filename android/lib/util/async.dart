// Dart imports:
import 'dart:async';

// Flutter imports:
import 'package:flutter/widgets.dart';

class Poller {
  Completer<void> _completer = Completer();

  Future<void> awake() async {
    if (_completer.isCompleted) _completer = Completer();
    await _completer.future;
  }

  Future<void> get done => _completer.future;
  void poll() => _completer.complete();
}

class MemoizedStreamController<T> {
  MemoizedStreamController({required T initial}) {
    add(initial);
  }

  void add(T value) {
    _current = value;
    wrapped.add(value);
  }

  final wrapped = StreamController<T>.broadcast();
  late T _current;
  T get current => _current;
  Stream<T> get stream => wrapped.stream;

  Future<T> waitUntil(List<T> args) {
    final completer = Completer<T>();
    final targets = Set.from(args);

    StreamSubscription<T>? sub;
    sub = stream.listen((s) {
      if (targets.contains(s)) {
        completer.complete(s);
        sub?.cancel();
      }
    }, onDone: () {
      completer.completeError('done');
    }, onError: (err) {
      completer.completeError(err);
      sub?.cancel();
    });
    return completer.future;
  }

  StreamBuilder<T> asStreamBuilder(
    Widget Function(BuildContext ctx, AsyncSnapshot<T> snap) builder,
  ) {
    return StreamBuilder(
        builder: builder, stream: stream, initialData: current);
  }
}
