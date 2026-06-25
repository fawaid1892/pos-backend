import 'base_dao.dart';

/// DAO for users table.
class UserDao extends BaseDao<Map<String, dynamic>> {
  @override
  String get tableName => 'users';

  @override
  Map<String, dynamic> toMap(Map<String, dynamic> item) => item;

  @override
  Map<String, dynamic> fromMap(Map<String, dynamic> map) => map;
}
