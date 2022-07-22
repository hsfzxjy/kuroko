// Flutter imports:
import 'package:flutter/material.dart';

// Package imports:
import 'package:kcore_android/kcore.dart';
import 'package:provider/provider.dart';

// Project imports:
import 'package:kuroko/core/bridge_session.dart';
import 'package:kuroko/endpoints/in.dart';
import 'package:kuroko/models/peers.dart';
import 'package:kuroko/states.dart';
import 'package:kuroko/views/devices_tab.dart';
import 'package:kuroko/views/settings_tab.dart';
import 'package:kuroko/views/upload_tab.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();
  final _ = SharingManager.i;
  await Kcore.init();

  Accepter.sessions.listen((sess) {
    routeEndpoint(BridgeSession(sess));
  });

  runApp(MultiProvider(
    providers: [
      ChangeNotifierProvider.value(value: PeerList.i),
      ChangeNotifierProvider.value(value: TabIndex.i),
      ChangeNotifierProvider.value(value: PendingUploadFiles.i),
      StreamProvider<List<Peer>>.value(
        value: PeerList.i.list,
        initialData: const [],
      ),
      StreamProvider<List<UploadFileGroup>>.value(
        value: UploadManager.i.groups.stream,
        initialData: UploadManager.i.groups.current,
      ),
    ],
    child: const KurokoApp(),
  ));
}

class KurokoApp extends StatelessWidget {
  const KurokoApp({super.key});

  static final _widgets = <Widget>[
    const DevicesTab(),
    const UploadTab(),
    const Text('Download'),
    const SettingsTab(),
  ];

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      theme: ThemeData(
          appBarTheme: const AppBarTheme(backgroundColor: Colors.teal)),
      title: 'Kuroko Teleporter',
      home: Scaffold(
        appBar: AppBar(title: const Text('Selector')),
        bottomNavigationBar: Consumer<TabIndex>(
          builder: (_, idx, _w) => BottomNavigationBar(
            selectedItemColor: Colors.teal,
            type: BottomNavigationBarType.fixed,
            items: const [
              BottomNavigationBarItem(
                  icon: Icon(Icons.devices_sharp), label: 'Devices'),
              BottomNavigationBarItem(
                  icon: Icon(Icons.upload_sharp), label: 'Upload'),
              BottomNavigationBarItem(
                  icon: Icon(Icons.download_sharp), label: 'Download'),
              BottomNavigationBarItem(
                  icon: Icon(Icons.settings_sharp), label: 'Settings'),
            ],
            currentIndex: idx.value,
            onTap: (index) {
              TabIndex.i.value = index;
            },
          ),
        ),
        body: Consumer<TabIndex>(
          builder: (_, idx, _w) => _widgets.elementAt(idx.value),
        ),
      ),
    );
  }
}
