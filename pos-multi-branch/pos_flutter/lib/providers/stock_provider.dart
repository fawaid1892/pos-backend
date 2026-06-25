import 'package:flutter/foundation.dart';
import '../models/stock_adjustment.dart';
import '../services/stock_service.dart';

/// State management for stock inventory, adjustments, transfers, and alerts.
///
/// Uses SQLite-backed StockService instead of mock data.
class StockProvider extends ChangeNotifier {
  final StockService _stockService = StockService();

  // Current branch context
  String? _currentBranchId;

  // Inventory state
  List<ProductStock> _inventory = [];
  List<ProductStock> _filteredInventory = [];
  bool _isLoadingInventory = false;
  String? _inventoryError;

  // Alerts state
  List<ProductStock> _alerts = [];
  bool _isLoadingAlerts = false;

  // Adjustment state
  bool _isSubmittingAdjustment = false;

  // Transfer state
  bool _isSubmittingTransfer = false;

  // Getters
  String? get currentBranchId => _currentBranchId;
  List<ProductStock> get inventory => _filteredInventory;
  bool get isLoadingInventory => _isLoadingInventory;
  String? get inventoryError => _inventoryError;
  List<ProductStock> get alerts => _alerts;
  bool get isLoadingAlerts => _isLoadingAlerts;
  bool get isSubmittingAdjustment => _isSubmittingAdjustment;
  bool get isSubmittingTransfer => _isSubmittingTransfer;

  void setBranch(String branchId) {
    _currentBranchId = branchId;
    notifyListeners();
  }

  /// Load inventory for the current branch from SQLite.
  Future<void> loadInventory({String? branchId}) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return;

    _isLoadingInventory = true;
    _inventoryError = null;
    notifyListeners();

    try {
      _inventory = await _stockService.getInventory(bid);
      _filteredInventory = List.from(_inventory);
      _isLoadingInventory = false;
    } catch (e) {
      debugPrint('StockProvider.loadInventory error: $e');
      _inventoryError = 'Gagal memuat inventory: ${e.toString()}';
      _isLoadingInventory = false;
    }
    notifyListeners();
  }

  /// Search inventory locally by name or barcode.
  void searchInventory(String query) {
    if (query.trim().isEmpty) {
      _filteredInventory = List.from(_inventory);
    } else {
      final q = query.toLowerCase();
      _filteredInventory = _inventory.where((p) {
        return p.productName.toLowerCase().contains(q) ||
            p.barcode.toLowerCase().contains(q);
      }).toList();
    }
    notifyListeners();
  }

  /// Load stock alerts for the current branch from SQLite.
  Future<void> loadAlerts({String? branchId}) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return;

    _isLoadingAlerts = true;
    notifyListeners();

    try {
      _alerts = await _stockService.getStockAlerts(bid);
      _isLoadingAlerts = false;
    } catch (e) {
      debugPrint('StockProvider.loadAlerts error: $e');
      _isLoadingAlerts = false;
    }
    notifyListeners();
  }

  /// Submit a stock adjustment (stock in / stock out) via SQLite.
  Future<StockAdjustment?> submitAdjustment({
    required String productId,
    required int quantity,
    required String reason,
    required String type,
    String? branchId,
  }) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return null;

    _isSubmittingAdjustment = true;
    notifyListeners();

    try {
      final result = await _stockService.adjustStock(
        branchId: bid,
        productId: productId,
        quantity: quantity,
        reason: reason,
        type: type,
      );
      // Reload inventory to reflect changes
      await loadInventory();
      _isSubmittingAdjustment = false;
      notifyListeners();
      return result;
    } catch (e) {
      debugPrint('StockProvider.submitAdjustment error: $e');
      _isSubmittingAdjustment = false;
      notifyListeners();
      return null;
    }
  }

  /// Submit a stock transfer between branches via SQLite.
  Future<StockTransfer?> submitTransfer({
    required String sourceBranchId,
    required String targetBranchId,
    required String productId,
    required int quantity,
  }) async {
    _isSubmittingTransfer = true;
    notifyListeners();

    try {
      final result = await _stockService.transferStock(
        sourceBranchId: sourceBranchId,
        targetBranchId: targetBranchId,
        productId: productId,
        quantity: quantity,
      );
      // Reload if current branch is source or target
      if (_currentBranchId == sourceBranchId) {
        await loadInventory();
      }
      _isSubmittingTransfer = false;
      notifyListeners();
      return result;
    } catch (e) {
      debugPrint('StockProvider.submitTransfer error: $e');
      _isSubmittingTransfer = false;
      notifyListeners();
      return null;
    }
  }

  /// Get product name by ID from current inventory.
  String? getProductName(String productId) {
    final idx = _inventory.indexWhere((p) => p.productId == productId);
    return idx >= 0 ? _inventory[idx].productName : null;
  }
}
