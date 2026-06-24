import '../models/branch.dart';
import '../models/stock_adjustment.dart';

/// Mock service for stock-related API calls.
/// In production, replace with actual HTTP client.
class MockStockService {
  static final MockStockService _instance = MockStockService._();
  factory MockStockService() => _instance;
  MockStockService._();

  /// Mock branches data
  final List<Branch> _branches = [
    Branch(id: 'branch_001', name: 'Cabang Utama', address: 'Jl. Merdeka No.1', phone: '021-1234567'),
    Branch(id: 'branch_002', name: 'Cabang Kedua', address: 'Jl. Sudirman No.45', phone: '021-7654321'),
    Branch(id: 'branch_003', name: 'Cabang Ketiga', address: 'Jl. Gatot Subroto No.88', phone: '021-5551234'),
  ];

  /// Stock per product per branch.
  /// Inner map: productId -> stock count
  final Map<String, Map<String, int>> _stockByBranch = {
    'branch_001': {
      'prod_001': 50,
      'prod_002': 40,
      'prod_003': 25,
      'prod_004': 30,
      'prod_005': 100,
      'prod_006': 60,
      'prod_007': 20,
      'prod_008': 3, // low stock
    },
    'branch_002': {
      'prod_001': 12,
      'prod_002': 8,
      'prod_003': 5,
      'prod_004': 0, // out of stock
      'prod_005': 45,
      'prod_006': 22,
      'prod_007': 2, // low stock
      'prod_008': 15,
    },
    'branch_003': {
      'prod_001': 30,
      'prod_002': 20,
      'prod_003': 0, // out of stock
      'prod_004': 10,
      'prod_005': 80,
      'prod_006': 15,
      'prod_007': 7,
      'prod_008': 25,
    },
  };

  /// Product name lookup
  final Map<String, String> _productNames = {
    'prod_001': 'Kopi Hitam',
    'prod_002': 'Kopi Susu',
    'prod_003': 'Nasi Goreng',
    'prod_004': 'Mie Goreng',
    'prod_005': 'Air Mineral',
    'prod_006': 'Teh Manis',
    'prod_007': 'Kentang Goreng',
    'prod_008': 'Jus Jeruk',
  };

  final Map<String, String> _productBarcodes = {
    'prod_001': '8991001001001',
    'prod_002': '8991001001002',
    'prod_003': '8991001001003',
    'prod_004': '8991001001004',
    'prod_005': '8991001001005',
    'prod_006': '8991001001006',
    'prod_007': '8991001001007',
    'prod_008': '8991001001008',
  };

  final Map<String, String> _productCategories = {
    'prod_001': 'Minuman',
    'prod_002': 'Minuman',
    'prod_003': 'Makanan',
    'prod_004': 'Makanan',
    'prod_005': 'Minuman',
    'prod_006': 'Minuman',
    'prod_007': 'Makanan',
    'prod_008': 'Minuman',
  };

  List<Branch> getBranches() => List.unmodifiable(_branches);

  /// GET /api/v1/branches/:id/inventory
  Future<List<ProductStock>> getInventory(String branchId) async {
    await Future.delayed(const Duration(milliseconds: 300));
    final stockMap = _stockByBranch[branchId] ?? {};
    return stockMap.entries.map((e) {
      return ProductStock(
        productId: e.key,
        productName: _productNames[e.key] ?? 'Unknown',
        barcode: _productBarcodes[e.key] ?? '',
        currentStock: e.value,
        minimumStock: 5,
        branchId: branchId,
      );
    }).toList();
  }

  /// POST /api/v1/branches/:id/inventory/adjustment
  Future<StockAdjustment> adjustStock({
    required String branchId,
    required String productId,
    required int quantity,
    required String reason,
    required String type,
  }) async {
    await Future.delayed(const Duration(milliseconds: 400));
    final adjustedQty = type == 'in' ? quantity : -quantity;

    _stockByBranch[branchId] ??= {};
    _stockByBranch[branchId]![productId] =
        (_stockByBranch[branchId]![productId] ?? 0) + adjustedQty;

    return StockAdjustment(
      id: 'adj_${DateTime.now().millisecondsSinceEpoch}',
      productId: productId,
      productName: _productNames[productId] ?? 'Unknown',
      branchId: branchId,
      quantity: adjustedQty,
      reason: reason,
      type: type,
      createdAt: DateTime.now(),
    );
  }

  /// POST /api/v1/inventory/transfer
  Future<StockTransfer> transferStock({
    required String sourceBranchId,
    required String targetBranchId,
    required String productId,
    required int quantity,
  }) async {
    await Future.delayed(const Duration(milliseconds: 400));

    // Deduct from source
    _stockByBranch[sourceBranchId] ??= {};
    _stockByBranch[sourceBranchId]![productId] =
        (_stockByBranch[sourceBranchId]![productId] ?? 0) - quantity;

    // Add to target
    _stockByBranch[targetBranchId] ??= {};
    _stockByBranch[targetBranchId]![productId] =
        (_stockByBranch[targetBranchId]![productId] ?? 0) + quantity;

    final sourceName =
        _branches.firstWhere((b) => b.id == sourceBranchId).name;
    final targetName =
        _branches.firstWhere((b) => b.id == targetBranchId).name;

    return StockTransfer(
      id: 'trf_${DateTime.now().millisecondsSinceEpoch}',
      sourceBranchId: sourceBranchId,
      sourceBranchName: sourceName,
      targetBranchId: targetBranchId,
      targetBranchName: targetName,
      productId: productId,
      productName: _productNames[productId] ?? 'Unknown',
      quantity: quantity,
      status: 'completed',
      createdAt: DateTime.now(),
    );
  }

  /// Get stock alerts for a branch (stock <= minimumStock)
  Future<List<ProductStock>> getStockAlerts(String branchId,
      {int minimumStock = 5}) async {
    await Future.delayed(const Duration(milliseconds: 200));
    final stockMap = _stockByBranch[branchId] ?? {};
    return stockMap.entries
        .where((e) => e.value <= minimumStock)
        .map((e) => ProductStock(
              productId: e.key,
              productName: _productNames[e.key] ?? 'Unknown',
              barcode: _productBarcodes[e.key] ?? '',
              currentStock: e.value,
              minimumStock: minimumStock,
              branchId: branchId,
            ))
        .toList();
  }

  /// Search inventory
  Future<List<ProductStock>> searchInventory(
      String branchId, String query) async {
    await Future.delayed(const Duration(milliseconds: 200));
    final stockMap = _stockByBranch[branchId] ?? {};
    final q = query.toLowerCase();
    return stockMap.entries
        .where((e) {
          final name = _productNames[e.key]?.toLowerCase() ?? '';
          final barcode = _productBarcodes[e.key]?.toLowerCase() ?? '';
          return name.contains(q) || barcode.contains(q);
        })
        .map((e) => ProductStock(
              productId: e.key,
              productName: _productNames[e.key] ?? 'Unknown',
              barcode: _productBarcodes[e.key] ?? '',
              currentStock: e.value,
              minimumStock: 5,
              branchId: branchId,
            ))
        .toList();
  }
}
