import 'base_dao.dart';

/// DAO for branch_products (inventory per branch) table.
class BranchProductDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'branch_products';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;

  /// Get inventory for a specific branch.
  Future<List<Map<String, dynamic>>> getByBranch(String branchId) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'branch_id = ?',
      whereArgs: [branchId],
      orderBy: 'updated_at DESC',
    );
    return maps;
  }

  /// Get a specific product stock in a branch.
  Future<Map<String, dynamic>?> getByBranchAndProduct(
      String branchId, String productId) async {
    final db = await _db;
    final maps = await db.query(
      tableName,
      where: 'branch_id = ? AND product_id = ?',
      whereArgs: [branchId, productId],
    );
    if (maps.isEmpty) return null;
    return maps.first;
  }

  /// Update stock quantity for a product in a branch.
  Future<void> updateStock(
      String branchId, String productId, int newStock) async {
    final db = await _db;
    await db.update(
      tableName,
      {
        'stock': newStock,
        'pending_sync': 1,
        'sync_status': 'pending',
        'updated_at': DateTime.now().toIso8601String(),
      },
      where: 'branch_id = ? AND product_id = ?',
      whereArgs: [branchId, productId],
    );
  }

  /// Get low stock products for a branch.
  Future<List<Map<String, dynamic>>> getLowStock(String branchId) async {
    final db = await _db;
    final maps = await db.rawQuery('''
      SELECT * FROM $tableName
      WHERE branch_id = ? AND stock <= minimum_stock
      ORDER BY stock ASC
    ''', [branchId]);
    return maps;
  }
}
