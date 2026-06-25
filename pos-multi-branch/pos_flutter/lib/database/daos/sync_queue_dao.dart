import 'base_dao.dart';

/// DAO for sync_queue table.
class SyncQueueDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'sync_queue';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;

  /// Add a new entry to the sync queue.
  Future<int> enqueue({
    required String tableName,
    required String recordId,
    required String action,
    required Map<String, dynamic> payload,
  }) async {
    final db = await _db;
    return await db.insert(
      'sync_queue',
      {
        'table_name': tableName,
        'record_id': recordId,
        'action': action,
        'payload': payload.toString(), // will store as JSON string
        'status': 'pending',
        'created_at': DateTime.now().toIso8601String(),
      },
    );
  }

  /// Get all pending sync queue entries.
  Future<List<Map<String, dynamic>>> getPending({
    int limit = 100,
  }) async {
    final db = await _db;
    return await db.query(
      'sync_queue',
      where: 'status = ?',
      whereArgs: ['pending'],
      orderBy: 'created_at ASC',
      limit: limit,
    );
  }

  /// Mark a queue entry as processed.
  Future<void> markProcessed(int id) async {
    final db = await _db;
    await db.update(
      'sync_queue',
      {
        'status': 'processed',
        'processed_at': DateTime.now().toIso8601String(),
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  /// Mark a queue entry as failed.
  Future<void> markFailed(int id, String errorMessage) async {
    final db = await _db;
    await db.update(
      'sync_queue',
      {
        'status': 'failed',
        'error_message': errorMessage,
        'processed_at': DateTime.now().toIso8601String(),
      },
      where: 'id = ?',
      whereArgs: [id],
    );
  }

  /// Get count of pending queue entries.
  Future<int> countPending() async {
    final db = await _db;
    final result = await db.rawQuery(
      "SELECT COUNT(*) as count FROM sync_queue WHERE status = 'pending'",
    );
    return Sqflite.firstIntValue(result) ?? 0;
  }

  /// Get all failed queue entries (Dead Letter Queue).
  Future<List<Map<String, dynamic>>> getFailed({int limit = 100}) async {
    final db = await _db;
    return await db.query(
      'sync_queue',
      where: 'status = ?',
      whereArgs: ['failed'],
      orderBy: 'created_at DESC',
      limit: limit,
    );
  }

  /// Get count of failed queue entries.
  Future<int> countFailed() async {
    final db = await _db;
    final result = await db.rawQuery(
      "SELECT COUNT(*) as count FROM sync_queue WHERE status = 'failed'",
    );
    return Sqflite.firstIntValue(result) ?? 0;
  }

  /// Retry a failed item by resetting its status back to pending.
  Future<bool> retryFailedItem(int id) async {
    final db = await _db;
    final updated = await db.update(
      'sync_queue',
      {
        'status': 'pending',
        'error_message': null,
        'processed_at': null,
      },
      where: 'id = ? AND status = ?',
      whereArgs: [id, 'failed'],
    );
    return updated > 0;
  }

  /// Delete a specific failed queue entry.
  Future<bool> deleteFailedItem(int id) async {
    final db = await _db;
    final deleted = await db.delete(
      'sync_queue',
      where: 'id = ? AND status = ?',
      whereArgs: [id, 'failed'],
    );
    return deleted > 0;
  }

  /// Clear all failed entries, optionally only those older than [olderThanDays].
  Future<int> clearResolved({int olderThanDays = 0}) async {
    final db = await _db;
    if (olderThanDays > 0) {
      final cutoff = DateTime.now()
          .subtract(Duration(days: olderThanDays))
          .toIso8601String();
      final result = await db.delete(
        'sync_queue',
        where: "status = 'failed' AND created_at < ?",
        whereArgs: [cutoff],
      );
      return result;
    } else {
      final result = await db.delete(
        'sync_queue',
        where: "status = 'failed'",
      );
      return result;
    }
  }

  /// Clear processed entries older than [olderThanDays] days.
  Future<void> cleanUp({int olderThanDays = 7}) async {
    final db = await _db;
    final cutoff = DateTime.now()
        .subtract(Duration(days: olderThanDays))
        .toIso8601String();
    await db.delete(
      'sync_queue',
      where: "status = 'processed' AND processed_at < ?",
      whereArgs: [cutoff],
    );
  }

  /// Get count of all entries per status.
  Future<Map<String, int>> getStatusCounts() async {
    final db = await _db;
    final result = <String, int>{
      'pending': 0,
      'failed': 0,
      'processed': 0,
    };
    for (final status in result.keys) {
      final count = await db.rawQuery(
        'SELECT COUNT(*) as count FROM sync_queue WHERE status = ?',
        [status],
      );
      result[status] = Sqflite.firstIntValue(count) ?? 0;
    }
    return result;
  }
}
