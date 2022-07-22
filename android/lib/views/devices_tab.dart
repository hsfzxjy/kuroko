// Flutter imports:
import 'package:flutter/material.dart';

// Package imports:
import 'package:kcore_android/kcore.dart';
import 'package:provider/provider.dart';

// Project imports:
import 'package:kuroko/models/bridges.dart';
import 'package:kuroko/models/peers.dart';
import 'package:kuroko/views/bluetooth.dart';
import 'package:kuroko/views/upload_tab.dart';

class DevicesTab extends StatelessWidget {
  const DevicesTab({super.key});

  void _onTap(BuildContext context) async {
    final Bridge? bridge = await Navigator.push(
      context,
      MaterialPageRoute(builder: (_) => const FindBluetoothDeviceScreen()),
    );
    if (bridge == null) return;
    PendingUploadFiles.i.resolve(bridge);
  }

  Widget _buildConnectNewDeviceCard(BuildContext context) {
    return Card(
      child: Column(children: [
        const ListTile(
            title: Text('Connect new device',
                style: TextStyle(fontWeight: FontWeight.bold))),
        const Divider(height: 0),
        ListTile(
          title: const Text('Bluetooth...'),
          trailing: const Icon(Icons.open_in_new_sharp),
          leading: const Icon(Icons.bluetooth),
          onTap: () => _onTap(context),
        ),
      ]),
    );
  }

  Widget _buildBridgeListTile(BuildContext context, Bridge bridge) {
    var typeName = '';
    IconData? icon;

    if (bridge.type == TransportType.bluetooth) {
      typeName = 'Bluetooth';
      icon = Icons.bluetooth_sharp;
    }

    return ListTile(
      title: Text(typeName),
      subtitle: Text(bridge.address),
      leading: Icon(icon),
      onTap: () => PendingUploadFiles.i.resolve(bridge),
    );
  }

  Widget _buildPeerCard(BuildContext context, Peer peer) {
    return Card(
      child: Column(children: [
        ListTile(
            title: Text(peer.name,
                style: const TextStyle(fontWeight: FontWeight.bold))),
        const Divider(height: 0),
        ...peer.bridges!.map((br) => _buildBridgeListTile(context, br)),
      ]),
    );
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context).textTheme;

    Widget buildSplitter(String text) {
      return Container(
        margin: const EdgeInsets.only(top: 10, bottom: 10),
        child: Row(
          children: [
            const Expanded(child: Divider()),
            Container(
              padding: const EdgeInsets.only(left: 5, right: 5),
              child: Text(text,
                  textAlign: TextAlign.center,
                  style: TextStyle(
                      fontWeight: FontWeight.w300,
                      fontSize: theme.subtitle1?.fontSize,
                      color: Colors.grey.shade600)),
            ),
            const Expanded(child: Divider()),
          ],
        ),
      );
    }

    const peersText = ['RECENT', 'OR', 'HISTORY'];

    return RefreshIndicator(
      child: Consumer<List<Peer>>(builder: (context, peers, _) {
        final peersLength = peers.length;
        return ListView.builder(
          itemCount: peersLength + 1,
          itemBuilder: (context, index) {
            Widget? splitter;
            if (peersLength > 0 && index < peersText.length) {
              splitter = buildSplitter(peersText.elementAt(index));
            }
            final Widget card = index == 1 || (index + peersLength == 0)
                ? _buildConnectNewDeviceCard(context)
                : _buildPeerCard(context, peers[index > 1 ? index - 1 : 0]);
            return Column(
              mainAxisSize: MainAxisSize.min,
              children: [if (splitter != null) splitter, card],
            );
          },
        );
      }),
      onRefresh: () async {
        await Provider.of<PeerList>(context, listen: false).refetch();
      },
    );
  }
}
