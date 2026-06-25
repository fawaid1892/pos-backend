import 'package:sqflite/sqflite.dart';
import '../local_database.dart';

/// Abstract base DAO providing common CRUD operations with sync fields.
abstract class BaseDao<T> {
  String get tableName;

  /// Convert a database row map to the model.
  T fromMap(Map<String, dynamic> map);

  /// Convert the model to a database row map.
  Map<String, dynamic> toMap(T item);

  Future<Database> get _db => LocalDatabase().database;

  /// Get all records from the table.
  Future<List<T>> getAll() async {
    final db = await _db;
    final maps = await db.query(tableName, orderBy: 'updated_at DESC');
    return maps.map(fromMap).toList();
  }

  /// Get a record by its primary key (assumes 'id' column).
  Future<T?> getById(String id) async {
    final db = await _db;
    final maps = await db.query(tableName, where: 'id = ?', whereArgs: [id]);
    if (maps.isEmpty) return null;
    return fromMap(maps.first);
  }

  /// Insert a new record.
  Future<int> insert(T item, {String? idField}) async {
    final db = await _db;
    final map = toMap(item);
    // Ensure pending_sync is set to 1 for offline-first
    if (map.containsKey('pending_sync')) {
      map['pending_sync'] = 1;
    }
    if (map.containsKey('sync_status')) {
      map['sync_status'] = 'pending';
    }
    return await db.insert(tableName, map,
        conflictAlgorithm: ConflictAlgorithm.replace);
  }

  /// Update an existing record by its 'id' column.
  Future<int> update(T item) async {
    final db = await _db;
    final map = toMap(item);
    // Mark as pending sync
    if (map.containsKey('pending_sync')) {
      map['pending_sync'] = 1;
    }
    if (map.containsKey('sync_status')) {
      map['sync_status'] = 'pending';
    }
    return await db.update(
      tableName,
      map,
      where: 'id = ?',
      whereArgs: [map['id']],
    );
  }

  /// Delete a record by its 'id' column.
  Future<int> delete(String id) async {
    final db = await _db;
    return await db.delete(tableName, where: 'id = ?', whereArgs: [id]);
  }

  /// Get all records with pending_sync = true.
  Future<List<T>> getPendingSync() async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'pending_sync = ?',
      whereArgs: [1],
    );
    return maps.map(fromMap).toList();
  }

  /// Mark a record as synced.
  Future<void> markSynced(String id) async {
    final db = await _db;
    await db.update(
      tableName,
      {
        'pending_sync': 0,
        'synced_at': DateTime.now().toIso8601String(),
        'sync_status': 'synced',
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  /// Mark a record as conflicted.
  Future<void> markConflict(String id) async {
    final db = await _db;
    await db.update(
      tableName,
      {
        'sync_status': 'conflict',
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  /// Get count of pending sync records.
  Future<int> countPendingSync() async {
    final db = await _db;
    final result = await db.rawQuery(
      'SELECT COUNT(*) as count FROM $tableName WHERE pending_sync = 1',
    );
    return Sqflite.firstIntValue(result) ?? 0;
  }

  /// Insert or replace a record (for sync pull / upsert).
  Future<int> upsert(Map<String, dynamic> map) async {
    final db = await _db;
    // When upserting from server, set pending_sync = 0 and sync_status = 'synced'
    if (map.containsKey('pending_sync')) {
      map['pending_sync'] = 0;
    }
    if (map.containsKey('synced_at')) {
      map['synced_at'] = DateTime.now().toIso8601String();
    }
    if (map.containsKey('sync_status')) {
      map['sync_status'] = 'synced';
    }
    return await db.insert(
      tableName,
      map,
      conflictAlgorithm: ConflictAlgorithm.replace,
    );
  }
}
