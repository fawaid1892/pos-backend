import 'base_dao.dart';

/// DAO for products table.
class ProductDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'products';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;

  /// Search products by name or barcode locally.
  Future<List<Map<String, dynamic>>> search(String query) async {
    final db = await _db;
    final q = '%$query%';
    final maps = await db.query(
      tableName,
      where: 'name LIKE ? OR barcode LIKE ?',
      whereArgs: [q, q],
      orderBy: 'name ASC',
    );
    return maps;
  }
}
