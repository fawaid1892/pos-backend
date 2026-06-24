import '../models/report.dart';
import '../models/product.dart';

/// Mock service for report-related API calls.
/// In production, replace with actual HTTP client.
class MockReportService {
  static final MockReportService _instance = MockReportService._();
  factory MockReportService() => _instance;
  MockReportService._();

  /// Mock transactions data
  final List<Map<String, dynamic>> _mockTransactions = List.generate(
    30,
    (i) => {
      'id': 'trx_${(1000 + i).toString()}',
      'branchId': i < 12
          ? 'branch_001'
          : (i < 22 ? 'branch_002' : 'branch_003'),
      'items': [
        {
          'productId': 'prod_${(i % 8 + 1).toString().padLeft(3, '0')}',
          'productName': [
            'Kopi Hitam', 'Kopi Susu', 'Nasi Goreng',
            'Mie Goreng', 'Air Mineral', 'Teh Manis',
            'Kentang Goreng', 'Jus Jeruk'
          ][i % 8],
          'quantity': (i % 3) + 1,
          'price': [15000, 20000, 35000, 25000, 5000, 8000, 20000, 18000][i % 8],
          'total': ([15000, 20000, 35000, 25000, 5000, 8000, 20000, 18000][i % 8]) * ((i % 3) + 1),
        }
      ],
      'total': 50000.0 + (i * 10000),
      'grandTotal': 50000.0 + (i * 10000),
      'createdAt': DateTime.now()
          .subtract(Duration(days: i ~/ 2, hours: i * 2))
          .toIso8601String(),
    },
  );

  /// GET /api/v1/branches/:id/reports/sales?start=&end=
  Future<SalesReport> getSalesReport({
    required String branchId,
    required DateTime startDate,
    required DateTime endDate,
  }) async {
    await Future.delayed(const Duration(milliseconds: 400));

    final branchName = _getBranchName(branchId);
    final filtered = _mockTransactions.where((t) {
      final date = DateTime.parse(t['createdAt'] as String);
      return t['branchId'] == branchId &&
          date.isAfter(startDate.subtract(const Duration(days: 1))) &&
          date.isBefore(endDate.add(const Duration(days: 1)));
    }).toList();

    double totalSales = 0;
    int totalItems = 0;
    final Map<String, SalesReportItem> itemMap = {};

    for (final t in filtered) {
      totalSales += (t['grandTotal'] as num).toDouble();
      for (final item in t['items'] as List) {
        final qty = item['quantity'] as int;
        final tot = (item['total'] as num).toDouble();
        totalItems += qty;
        final pid = item['productId'] as String;
        final pName = item['productName'] as String;
        if (itemMap.containsKey(pid)) {
          itemMap[pid] = SalesReportItem(
            productId: pid,
            productName: pName,
            quantity: itemMap[pid]!.quantity + qty,
            total: itemMap[pid]!.total + tot,
          );
        } else {
          itemMap[pid] = SalesReportItem(
            productId: pid,
            productName: pName,
            quantity: qty,
            total: tot,
          );
        }
      }
    }

    return SalesReport(
      branchId: branchId,
      branchName: branchName,
      startDate: startDate,
      endDate: endDate,
      totalSales: totalSales,
      totalTransactions: filtered.length,
      averageTransactionValue: filtered.isEmpty ? 0 : totalSales / filtered.length,
      totalItemsSold: totalItems,
      items: itemMap.values.toList(),
    );
  }

  /// GET /api/v1/branches/:id/reports/stock
  Future<StockReport> getStockReport(String branchId) async {
    await Future.delayed(const Duration(milliseconds: 300));

    final branchName = _getBranchName(branchId);
    final stockService = MockReportService._instance;

    // Simplified product stub data
    final products = [
      {'id': 'prod_001', 'name': 'Kopi Hitam', 'cat': 'Minuman', 'price': 15000, 'stock': branchId == 'branch_001' ? 50 : (branchId == 'branch_002' ? 12 : 30)},
      {'id': 'prod_002', 'name': 'Kopi Susu', 'cat': 'Minuman', 'price': 20000, 'stock': branchId == 'branch_001' ? 40 : (branchId == 'branch_002' ? 8 : 20)},
      {'id': 'prod_003', 'name': 'Nasi Goreng', 'cat': 'Makanan', 'price': 35000, 'stock': branchId == 'branch_001' ? 25 : (branchId == 'branch_002' ? 5 : 0)},
      {'id': 'prod_004', 'name': 'Mie Goreng', 'cat': 'Makanan', 'price': 25000, 'stock': branchId == 'branch_001' ? 30 : (branchId == 'branch_002' ? 0 : 10)},
      {'id': 'prod_005', 'name': 'Air Mineral', 'cat': 'Minuman', 'price': 5000, 'stock': branchId == 'branch_001' ? 100 : (branchId == 'branch_002' ? 45 : 80)},
      {'id': 'prod_006', 'name': 'Teh Manis', 'cat': 'Minuman', 'price': 8000, 'stock': branchId == 'branch_001' ? 60 : (branchId == 'branch_002' ? 22 : 15)},
      {'id': 'prod_007', 'name': 'Kentang Goreng', 'cat': 'Makanan', 'price': 20000, 'stock': branchId == 'branch_001' ? 20 : (branchId == 'branch_002' ? 2 : 7)},
      {'id': 'prod_008', 'name': 'Jus Jeruk', 'cat': 'Minuman', 'price': 18000, 'stock': branchId == 'branch_001' ? 3 : (branchId == 'branch_002' ? 15 : 25)},
    ];

    final reports = products.map((p) {
      final stock = p['stock'] as int;
      return ProductStockReport(
        productId: p['id'] as String,
        productName: p['name'] as String,
        category: p['cat'] as String,
        currentStock: stock,
        minimumStock: 5,
        price: (p['price'] as num).toDouble(),
        stockValue: stock * (p['price'] as int),
      );
    }).toList();

    int lowStock = reports.where((r) => r.isLowStock && !r.isOutOfStock).length;
    int outOfStock = reports.where((r) => r.isOutOfStock).length;
    int totalValue = reports.fold(0, (sum, r) => sum + (r.currentStock * r.price.toInt()));

    return StockReport(
      branchId: branchId,
      branchName: branchName,
      totalProducts: reports.length,
      lowStockCount: lowStock,
      outOfStockCount: outOfStock,
      totalStockValue: totalValue,
      products: reports,
    );
  }

  /// GET /api/v1/branches/:id/reports/profit-loss?start=&end=
  Future<ProfitLossReport> getProfitLossReport({
    required String branchId,
    required DateTime startDate,
    required DateTime endDate,
  }) async {
    await Future.delayed(const Duration(milliseconds: 400));

    final branchName = _getBranchName(branchId);
    final filtered = _mockTransactions.where((t) {
      final date = DateTime.parse(t['createdAt'] as String);
      return t['branchId'] == branchId &&
          date.isAfter(startDate.subtract(const Duration(days: 1))) &&
          date.isBefore(endDate.add(const Duration(days: 1)));
    }).toList();

    final totalRevenue =
        filtered.fold(0.0, (sum, t) => sum + (t['grandTotal'] as num).toDouble());
    final totalCost = totalRevenue * 0.6; // Mock: 60% cost of goods sold
    final grossProfit = totalRevenue - totalCost;
    final totalExpenses = totalRevenue * 0.2; // Mock: 20% operating expenses
    final netProfit = grossProfit - totalExpenses;

    return ProfitLossReport(
      branchId: branchId,
      branchName: branchName,
      startDate: startDate,
      endDate: endDate,
      totalRevenue: totalRevenue,
      totalCost: totalCost,
      grossProfit: grossProfit,
      totalExpenses: totalExpenses,
      netProfit: netProfit,
      items: [
        ProfitLossItem(
            name: 'Total Pendapatan',
            category: 'revenue',
            amount: totalRevenue),
        ProfitLossItem(name: 'HPP (COGS)', category: 'cost', amount: totalCost),
        ProfitLossItem(
            name: 'Laba Kotor', category: 'revenue', amount: grossProfit),
        ProfitLossItem(
            name: 'Biaya Operasional',
            category: 'expense',
            amount: totalExpenses),
        ProfitLossItem(
            name: 'Laba Bersih', category: 'revenue', amount: netProfit),
      ],
    );
  }

  /// GET /api/v1/branches/:id/reports/sales/export?format=pdf|xlsx
  Future<String> exportReport({
    required String branchId,
    required String format,
    required String reportType,
    required DateTime startDate,
    required DateTime endDate,
  }) async {
    await Future.delayed(const Duration(milliseconds: 800));
    // In production, would return download URL or file path
    return 'Laporan exported: $reportType - $format - '
        '${startDate.toIso8601String()} to ${endDate.toIso8601String()}';
  }

  String _getBranchName(String branchId) {
    const names = {
      'branch_001': 'Cabang Utama',
      'branch_002': 'Cabang Kedua',
      'branch_003': 'Cabang Ketiga',
    };
    return names[branchId] ?? 'Cabang Lain';
  }
}
