import 'base_dao.dart';

/// DAO for stock_mutations table (adjustments + transfers).
class StockMutationDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'stock_mutations';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;

  /// Get mutations by branch.
  Future<List<Map<String, dynamic>>> getByBranch(String branchId,
      {int limit = 50}) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'branch_id = ?',
      whereArgs: [branchId],
      orderBy: 'created_at DESC',
      limit: limit,
    );
    return maps;
  }

  /// Get mutations by type (in/out/transfer_in/transfer_out).
  Future<List<Map<String, dynamic>>> getByType(
      String branchId, String type) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'branch_id = ? AND type = ?',
      whereArgs: [branchId, type],
      orderBy: 'created_at DESC',
    );
    return maps;
  }
}
