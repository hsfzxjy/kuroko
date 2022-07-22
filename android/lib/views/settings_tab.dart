// Flutter imports:
import 'package:flutter/material.dart';

// Package imports:
import 'package:kcore_android/kcore.dart';

class SettingsTab extends StatelessWidget {
  const SettingsTab({super.key});

  @override
  Widget build(BuildContext context) {
    return ListView(children: [
      StreamBuilder<AccepterState>(
        builder: (ctx, snap) {
          final state = snap.data ?? AccepterState.ended;
          return SwitchListTile(
            title: const Text('Bluetooth Accepter'),
            value: state.isStart,
            onChanged: state.isTransient
                ? null
                : (_) {
                    switch (state) {
                      case AccepterState.started:
                        accepters[TransportType.bluetooth]!.stop();
                        break;
                      case AccepterState.error:
                      case AccepterState.ended:
                        accepters[TransportType.bluetooth]!.start();
                        break;
                      default:
                    }
                  },
          );
        },
        stream: accepters[TransportType.bluetooth]!.state,
      ),
    ]);
  }
}
