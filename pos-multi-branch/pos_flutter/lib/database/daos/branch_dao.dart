import 'base_dao.dart';

/// DAO for branches table.
class BranchDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'branches';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;
}
