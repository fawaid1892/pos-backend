import '../database/local_database.dart';
import '../models/report.dart';
import 'pdf_export_service.dart';

/// Service for report queries backed by SQLite.
///
/// Replaces MockReportService with actual database queries.
class ReportService {
  static final ReportService _instance = ReportService._internal();
  factory ReportService() => _instance;
  ReportService._internal();

  final LocalDatabase _db = LocalDatabase();

  /// Get branch name by ID.
  Future<String> _getBranchName(String branchId) async {
    final db = await _db.database;
    final rows = await db.query('branches',
      where: 'id = ?', whereArgs: [branchId]);
    if (rows.isEmpty) return 'Cabang Lain';
    return rows.first['name'] as String;
  }

  /// GET sales report from local transactions.
  Future<SalesReport> getSalesReport({
    required String branchId,
    required DateTime startDate,
    required DateTime endDate,
  }) async {
    final db = await _db.database;
    final branchName = await _getBranchName(branchId);

    final startStr = startDate.toIso8601String();
    final endStr = endDate.toIso8601String();

    // Get transactions in date range
    final transactions = await db.rawQuery('''
      SELECT * FROM transactions
      WHERE branch_id = ?
        AND created_at >= ?
        AND created_at <= ?
      ORDER BY created_at DESC
    ''', [branchId, startStr, endStr]);

    double totalSales = 0;
    int totalItems = 0;
    final Map<String, SalesReportItem> itemMap = {};

    for (final t in transactions) {
      totalSales += (t['grand_total'] as num?)?.toDouble() ?? 0;

      // Get items for this transaction
      final items = await db.rawQuery('''
        SELECT * FROM transaction_items
        WHERE transaction_id = ?
      ''', [t['id']]);

      for (final item in items) {
        final qty = (item['quantity'] as num?)?.toInt() ?? 0;
        final tot = (item['subtotal'] as num?)?.toDouble() ?? 0;
        totalItems += qty;

        final pid = item['product_id'] as String;
        final pName = item['product_name'] as String;

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

    final trxCount = transactions.length;

    return SalesReport(
      branchId: branchId,
      branchName: branchName,
      startDate: startDate,
      endDate: endDate,
      totalSales: totalSales,
      totalTransactions: trxCount,
      averageTransactionValue:
          trxCount > 0 ? totalSales / trxCount : 0,
      totalItemsSold: totalItems,
      items: itemMap.values.toList(),
    );
  }

  /// GET stock report from branch_products + products.
  Future<StockReport> getStockReport(String branchId) async {
    final db = await _db.database;
    final branchName = await _getBranchName(branchId);

    final maps = await db.rawQuery('''
      SELECT
        bp.product_id,
        p.name AS product_name,
        p.category,
        p.price,
        bp.stock AS current_stock,
        bp.minimum_stock
      FROM branch_products bp
      INNER JOIN products p ON p.id = bp.product_id
      WHERE bp.branch_id = ?
      ORDER BY p.name ASC
    ''', [branchId]);

    final reports = maps.map((m) {
      final stock = (m['current_stock'] as num).toInt();
      final price = (m['price'] as num).toDouble();
      return ProductStockReport(
        productId: m['product_id'] as String,
        productName: m['product_name'] as String,
        category: m['category'] as String? ?? 'General',
        currentStock: stock,
        minimumStock: (m['minimum_stock'] as num?)?.toInt() ?? 5,
        price: price,
        stockValue: stock * price,
      );
    }).toList();

    int lowStock = reports.where((r) => r.isLowStock && !r.isOutOfStock).length;
    int outOfStock = reports.where((r) => r.isOutOfStock).length;
    int totalValue = reports.fold(0, (sum, r) => sum + r.currentStock * r.price.toInt());

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

  /// GET profit-loss report from transactions.
  Future<ProfitLossReport> getProfitLossReport({
    required String branchId,
    required DateTime startDate,
    required DateTime endDate,
  }) async {
    final db = await _db.database;
    final branchName = await _getBranchName(branchId);

    final startStr = startDate.toIso8601String();
    final endStr = endDate.toIso8601String();

    final result = await db.rawQuery('''
      SELECT COALESCE(SUM(grand_total), 0) AS total_revenue
      FROM transactions
      WHERE branch_id = ?
        AND created_at >= ?
        AND created_at <= ?
    ''', [branchId, startStr, endStr]);

    final totalRevenue = (result.first['total_revenue'] as num?)?.toDouble() ?? 0;
    final totalCost = totalRevenue * 0.6; // Approximation: 60% COGS
    final grossProfit = totalRevenue - totalCost;
    final totalExpenses = totalRevenue * 0.2; // Approximation: 20% operating expenses
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
        ProfitLossItem(name: 'Total Pendapatan', category: 'revenue', amount: totalRevenue),
        ProfitLossItem(name: 'HPP (COGS)', category: 'cost', amount: totalCost),
        ProfitLossItem(name: 'Laba Kotor', category: 'revenue', amount: grossProfit),
        ProfitLossItem(name: 'Biaya Operasional', category: 'expense', amount: totalExpenses),
        ProfitLossItem(name: 'Laba Bersih', category: 'revenue', amount: netProfit),
      ],
    );
  }

  /// Export report — generates actual PDF file for 'pdf' format.
  Future<String> exportReport({
    required String branchId,
    required String format,
    required String reportType,
    required DateTime startDate,
    required DateTime endDate,
  }) async {
    if (format == 'pdf') {
      final pdfService = PdfExportService();

      switch (reportType) {
        case 'sales':
          final report = await getSalesReport(
              branchId: branchId, startDate: startDate, endDate: endDate);
          return await pdfService.exportSalesReport(report);

        case 'stock':
          final report = await getStockReport(branchId);
          return await pdfService.exportStockReport(report);

        case 'profit_loss':
          final report = await getProfitLossReport(
              branchId: branchId, startDate: startDate, endDate: endDate);
          return await pdfService.exportProfitLossReport(report);
      }
    }

    // For non-PDF formats, return stub (Excel would be generated here)
    await Future.delayed(const Duration(milliseconds: 500));
    return 'Laporan terexport: $reportType - $format - '
        '${startDate.toIso8601String()} to ${endDate.toIso8601String()}';
  }
}
