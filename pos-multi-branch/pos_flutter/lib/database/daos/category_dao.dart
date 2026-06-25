import 'base_dao.dart';

/// DAO for categories table.
class CategoryDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'categories';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;
}
