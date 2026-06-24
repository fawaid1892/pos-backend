import 'package:flutter/foundation.dart';

class AuthProvider extends ChangeNotifier {
  String? _token;
  String? _userId;
  String? _userName;
  String? _branchId;
  String? _branchName;
  String? _role; // 'owner' or 'cashier'
  bool _isLoading = false;

  String? get token => _token;
  String? get userId => _userId;
  String? get userName => _userName;
  String? get branchId => _branchId;
  String? get branchName => _branchName;
  String? get role => _role;
  bool get isLoading => _isLoading;
  bool get isLoggedIn => _token != null;

  Future<bool> login(String email, String password) async {
    _isLoading = true;
    notifyListeners();

    try {
      // Simulate API call delay
      await Future.delayed(const Duration(milliseconds: 800));

      // Mock login — in production, call Supabase Auth
      if (email.contains('owner')) {
        _token = 'mock_token_owner_${DateTime.now().millisecondsSinceEpoch}';
        _userId = 'user_001';
        _userName = 'Owner';
        _branchId = null; // owner sees all branches
        _branchName = null;
        _role = 'owner';
      } else {
        _token = 'mock_token_kasir_${DateTime.now().millisecondsSinceEpoch}';
        _userId = 'user_002';
        _userName = 'Kasir';
        _branchId = 'branch_001';
        _branchName = 'Cabang Utama';
        _role = 'cashier';
      }

      _isLoading = false;
      notifyListeners();
      return true;
    } catch (e) {
      _isLoading = false;
      notifyListeners();
      return false;
    }
  }

  Future<void> logout() async {
    _token = null;
    _userId = null;
    _userName = null;
    _branchId = null;
    _branchName = null;
    _role = null;
    notifyListeners();
  }

  void setBranch(String id, String name) {
    _branchId = id;
    _branchName = name;
    notifyListeners();
  }
}
