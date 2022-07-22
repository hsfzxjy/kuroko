// Package imports:
import 'package:kcore_android/kcore.dart';
import 'package:sqflite/sqflite.dart';

// Project imports:
import 'package:kuroko/core/db.dart';

abstract class BridgeLike {
  String get address;
  TransportType get type;
}

mixin BridgeMixin on BridgeLike {
  Bridge? _bridge;
  Bridge get bridge {
    _bridge ??= Bridge.fromBridgeLike(this);
    return _bridge!;
  }
}

class Bridge implements BridgeLike {
  Bridge.fromBluetoothDevice(BluetoothDevice device)
      : type = TransportType.bluetooth,
        address = device.address;

  Bridge.fromBridgeLike(BridgeLike bi)
      : type = bi.type,
        address = bi.address;

  Bridge.fromMap(Map map)
      : type = TransportType.of(map['type']),
        address = map['address'] {
    fillFromMap(map);
  }

  Bridge(this.type, this.address);

  @override
  final TransportType type;
  @override
  final String address;
  int? bridgeId;
  String? name;
  int? peerid;
  int? peerTime;
  int? bridgeTime;

  bool get isAnonymous => bridgeId == null;

  void fillFromMap(Map map) {
    bridgeId = map['bridge_id'];
    name = map['name'];
    peerid = map['peerid'];
    peerTime = map['peer_time'];
    bridgeTime = map['bridge_time'];
  }

  Future<bool> fill(int? peerid, {DatabaseExecutor? db}) async {
    final map = await Bridge.getMap(
      type: type,
      address: address,
      db: db,
      peerid: peerid,
    );
    if (map == null) {
      return false;
    } else {
      fillFromMap(map);
      return true;
    }
  }

  static Future<List<Map>> _query(DatabaseExecutor db, String sql,
      [List<Object?>? params]) async {
    final stmt = '''
    SELECT
      peers.peerid as peerid,
      peers.name as name,
      bridges.bridge_id as bridge_id,
      bridges.peer_time as peer_time,
      bridges.bridge_time as bridge_time,
      bridges.type as type,
      bridges.address as address
    FROM
      peers
    LEFT OUTER JOIN
      bridges
    ON
      bridges.peerid = peers.peerid $sql''';
    return await db.rawQuery(stmt, params);
  }

  static Future<List<Bridge>> getList({DatabaseExecutor? db}) async {
    final ex = db ?? await openDb();
    final rows = await _query(ex, 'ORDER BY peer_time DESC, bridge_time DESC');
    return rows.map((map) => Bridge.fromMap(map)).toList();
  }

  static Future<Map?> getMap({
    required TransportType type,
    required String address,
    int? peerid,
    DatabaseExecutor? db,
  }) async {
    final ex = db ?? await openDb();
    var cond = 'WHERE type = ? AND address = ?';
    final params = [type.value, address];
    if (peerid != null) {
      cond += ' AND peers.peerid = ?';
      params.add(peerid);
    }
    final rows = await _query(ex, cond, params);
    switch (rows.length) {
      case 0:
        return null;
      case 1:
        return rows[0];
      default:
        throw 'multiple entries found';
    }
  }

  static Future<Bridge?> get({
    required TransportType type,
    required String address,
    int? peerid,
    DatabaseExecutor? db,
  }) async {
    final map =
        await getMap(address: address, type: type, peerid: peerid, db: db);
    return map == null ? null : Bridge.fromMap(map);
  }

  static Future<Bridge?> ensureExist({
    required int peerid,
    required String name,
    required TransportType type,
    required String address,
    bool noResult = true,
    DatabaseExecutor? db,
  }) async {
    final ex = db ?? await openDb();
    return await transaction(ex, (txn) async {
      final batch = txn.batch();
      batch.insert(
        'peers',
        {'peerid': peerid, 'name': name},
        conflictAlgorithm: ConflictAlgorithm.replace,
      );
      batch.insert(
        'bridges',
        {'peerid': peerid, 'address': address, 'type': type.value},
        conflictAlgorithm: ConflictAlgorithm.ignore,
      );
      await batch.commit(noResult: true);
      if (noResult) return null;
      return await get(type: type, address: address, db: txn);
    });
  }
}
