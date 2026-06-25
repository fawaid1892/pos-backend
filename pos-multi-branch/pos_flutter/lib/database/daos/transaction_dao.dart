import 'dart:convert';
import 'base_dao.dart';

/// DAO for transactions table.
class TransactionDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'transactions';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;

  /// Get transactions by branch.
  Future<List<Map<String, dynamic>>> getByBranch(String branchId,
      {int limit = 50, int offset = 0}) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'branch_id = ?',
      whereArgs: [branchId],
      orderBy: 'created_at DESC',
      limit: limit,
      offset: offset,
    );
    return maps;
  }

  /// Get transactions by date range.
  Future<List<Map<String, dynamic>>> getByDateRange(
      String branchId, DateTime start, DateTime end) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'branch_id = ? AND created_at >= ? AND created_at <= ?',
      whereArgs: [branchId, start.toIso8601String(), end.toIso8601String()],
      orderBy: 'created_at DESC',
    );
    return maps;
  }

  /// Get today's transactions for a branch.
  Future<List<Map<String, dynamic>>> getToday(String branchId) async {
    final todayStart = DateTime.now();
    final todayStartStr =
        '${todayStart.year}-${todayStart.month.toString().padLeft(2, '0')}-${todayStart.day.toString().padLeft(2, '0')}T00:00:00';
    return getByDateRange(branchId, DateTime.parse(todayStartStr), DateTime.now());
  }
}

/// DAO for transaction_items table.
class TransactionItemDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'transaction_items';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;

  /// Get items for a specific transaction.
  Future<List<Map<String, dynamic>>> getByTransaction(
      String transactionId) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'transaction_id = ?',
      whereArgs: [transactionId],
    );
    return maps;
  }

  /// Insert all items for a transaction in batch.
  Future<void> insertBatch(List<Map<String, dynamic>> items) async {
    final db = await _db;
    final batch = db.batch();
    for (final item in items) {
      item['pending_sync'] = 1;
      item['sync_status'] = 'pending';
      batch.insert(tableName, item, conflictAlgorithm: ConflictAlgorithm.replace);
    }
    await batch.commit(noResult: true);
  }
}
