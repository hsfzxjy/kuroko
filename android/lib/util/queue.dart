enum QueuedItemState {
  queuing,
  running,
  canceling,
  dequed,
}

abstract class QueuedItem {
  Future<void> onQueuedItemCanceled(bool started);
  void onQueuedItemScheduled(void Function() onFinished);
  void onQueuedItemEnqueued();
}

class _QueuedItemWrapper {
  final QueuedItem item;
  final ExclusiveQueue queue;
  var state = QueuedItemState.queuing;

  _QueuedItemWrapper(this.item, this.queue) {
    item.onQueuedItemEnqueued();
  }

  Future<void>? _cancelFut;

  Future<void> cancel() {
    switch (state) {
      case QueuedItemState.running:
      case QueuedItemState.queuing:
        final started = state == QueuedItemState.running;
        state = QueuedItemState.canceling;
        _cancelFut = item.onQueuedItemCanceled(started).whenComplete(() {
          queue._remove(item);
          state = QueuedItemState.dequed;
        });
        return _cancelFut!;

      default:
        return _cancelFut ?? Future.value();
    }
  }

  void start() {
    if (state != QueuedItemState.queuing) return;
    state = QueuedItemState.running;
    item.onQueuedItemScheduled(() {
      queue._remove(item);
      state = QueuedItemState.dequed;
      queue._poll();
    });
  }
}

class ExclusiveQueue {
  final _list = <QueuedItem>[];
  final _map = <QueuedItem, _QueuedItemWrapper>{};

  void _remove(QueuedItem item) {
    _map.remove(item);
    _list.remove(item);
  }

  Future<void> clear() {
    final copy = <_QueuedItemWrapper>[];
    copy.addAll(_map.values);
    return Future.wait(copy.map((w) => w.cancel()));
  }

  void _poll() {
    if (_list.isEmpty) return;
    final first = _list.first;
    final firstWrapper = _map[first]!;
    if (firstWrapper.state != QueuedItemState.queuing) return;
    firstWrapper.start();
  }

  void dequeue(QueuedItem item) {
    if (!_map.containsKey(item)) return;
    final wrapper = _map[item]!;
    wrapper.cancel();
  }

  void enqueue(QueuedItem item) {
    if (_map.containsKey(item)) return;
    final fut = clear();
    _list.add(item);
    _map[item] = _QueuedItemWrapper(item, this);
    fut.whenComplete(() => _poll());
  }

  void toggle(QueuedItem item) {
    if (_map.containsKey(item)) {
      dequeue(item);
    } else {
      enqueue(item);
    }
  }
}
