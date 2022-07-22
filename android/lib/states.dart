// Flutter imports:
import 'package:flutter/foundation.dart';

// Package imports:
import 'package:receive_sharing_intent/receive_sharing_intent.dart';

// Project imports:
import 'package:kuroko/views/upload_tab.dart';

class TabIndex extends ChangeNotifier {
  TabIndex._();
  static final TabIndex i = TabIndex._();

  var _value = 0;
  int get value => _value;
  set value(int v) {
    if (v != _value) {
      _value = v;
      notifyListeners();
    }
  }

  void toDevicesTab() {
    value = 0;
  }

  void toUploadTab() {
    value = 1;
  }

  void toDownloadTab() {
    value = 2;
  }

  void toSettingsTab() {
    value = 3;
  }
}

class SharingManager {
  SharingManager._() {
    ReceiveSharingIntent.getInitialMedia().then(_handleFiles);
    ReceiveSharingIntent.getMediaStream().listen(_handleFiles);
  }
  static final i = SharingManager._();

  void _handleFiles(List<SharedMediaFile> files) {
    PendingUploadFiles.i.replace(files.map((f) => f.path).toList());
    TabIndex.i.toDevicesTab();
  }
}
