// Flutter imports:
import 'package:flutter/widgets.dart';

// Package imports:
import 'package:path/path.dart' as path;
import 'package:sqflite/sqflite.dart';

Future<T> transaction<T>(
    DatabaseExecutor ex, Future<T> Function(Transaction txn) action) async {
  if (ex is Database) {
    return await ex.transaction(action);
  } else if (ex is Transaction) {
    return await action(ex);
  } else {
    throw 'unknown executor ${ex.runtimeType}';
  }
}

Future<Database> openDb({bool readOnly = false}) async {
  WidgetsFlutterBinding.ensureInitialized();

  Future<void> init(Database db) async {
    const statements = '''
    DROP TABLE IF EXISTS peers;
    CREATE TABLE IF NOT EXISTS peers (
      peerid INTEGER UNIQUE,
      name TEXT
    );
    CREATE INDEX IF NOT EXISTS peerid_index ON peers (peerid);

    DROP TABLE IF EXISTS bluetooth;
    DROP TABLE IF EXISTS bridges;
    DROP TABLE IF EXISTS peer_history;
    CREATE TABLE IF NOT EXISTS bridges (
      bridge_id INTEGER PRIMARY KEY,
      peerid INTEGER,
      type INTEGER,
      address TEXT,
      peer_time INTEGER,
      bridge_time INTEGER,
      UNIQUE(peerid, type, address)
    );
    CREATE INDEX IF NOT EXISTS bridge_id_index ON bridges (bridge_id);
    CREATE INDEX IF NOT EXISTS address_index ON bridges (address);
    CREATE INDEX IF NOT EXISTS peerid_index ON bridges (peerid);
    CREATE INDEX IF NOT EXISTS type_index ON bridges (type);
    CREATE INDEX IF NOT EXISTS peerid_type_index ON bridges (peerid, type);
    CREATE INDEX IF NOT EXISTS peer_time_index ON bridges (peer_time);
    CREATE INDEX IF NOT EXISTS bridge_time_index ON bridges (bridge_time);
    ''';
    final batch = db.batch();
    for (final statement in statements.split(';')) {
      if (statement.trim().isEmpty) continue;
      batch.execute(statement);
    }
    await batch.commit(noResult: true);
  }

  return await openDatabase(
    path.join(await getDatabasesPath(), 'main.db'),
    onCreate: (db, version) => init(db),
    onUpgrade: (db, _, __) => init(db),
    version: 11,
    readOnly: readOnly,
    singleInstance: true,
  );
}
