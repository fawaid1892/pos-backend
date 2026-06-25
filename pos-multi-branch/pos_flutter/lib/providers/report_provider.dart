import 'package:flutter/foundation.dart';
import '../models/report.dart';
import '../services/report_service.dart';

/// State management for sales, stock, and profit-loss reports.
///
/// Uses SQLite-backed ReportService instead of mock data.
class ReportProvider extends ChangeNotifier {
  final ReportService _reportService = ReportService();

  // Current branch context
  String? _currentBranchId;

  // Sales report state
  SalesReport? _salesReport;
  bool _isLoadingSalesReport = false;
  String? _salesReportError;

  // Stock report state
  StockReport? _stockReport;
  bool _isLoadingStockReport = false;
  String? _stockReportError;

  // Profit-loss report state
  ProfitLossReport? _profitLossReport;
  bool _isLoadingProfitLoss = false;
  String? _profitLossError;

  // Export state
  bool _isExporting = false;
  String? _exportResult;
  String? _exportError;

  // Active date range (for all reports)
  DateTimeRange? _activeDateRange;

  // Getters
  String? get currentBranchId => _currentBranchId;
  SalesReport? get salesReport => _salesReport;
  bool get isLoadingSalesReport => _isLoadingSalesReport;
  String? get salesReportError => _salesReportError;
  StockReport? get stockReport => _stockReport;
  bool get isLoadingStockReport => _isLoadingStockReport;
  String? get stockReportError => _stockReportError;
  ProfitLossReport? get profitLossReport => _profitLossReport;
  bool get isLoadingProfitLoss => _isLoadingProfitLoss;
  String? get profitLossError => _profitLossError;
  bool get isExporting => _isExporting;
  String? get exportResult => _exportResult;
  String? get exportError => _exportError;
  DateTimeRange? get activeDateRange => _activeDateRange;

  void setBranch(String branchId) {
    _currentBranchId = branchId;
    notifyListeners();
  }

  /// Set date range filter, default to current month.
  void setDateRange(DateTimeRange range) {
    _activeDateRange = range;
    notifyListeners();
  }

  void _ensureDateRange() {
    if (_activeDateRange == null) {
      final now = DateTime.now();
      _activeDateRange = DateTimeRange(
        start: DateTime(now.year, now.month, 1),
        end: now,
      );
    }
  }

  /// Load sales report for the given date range from SQLite.
  Future<void> loadSalesReport({
    String? branchId,
    DateTime? startDate,
    DateTime? endDate,
  }) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return;

    _ensureDateRange();
    final start = startDate ?? _activeDateRange!.start;
    final end = endDate ?? _activeDateRange!.end;

    _isLoadingSalesReport = true;
    _salesReportError = null;
    notifyListeners();

    try {
      _salesReport = await _reportService.getSalesReport(
        branchId: bid,
        startDate: start,
        endDate: end,
      );
      _isLoadingSalesReport = false;
    } catch (e) {
      debugPrint('ReportProvider.loadSalesReport error: $e');
      _salesReportError = 'Gagal memuat laporan penjualan: ${e.toString()}';
      _isLoadingSalesReport = false;
    }
    notifyListeners();
  }

  /// Load stock report from SQLite.
  Future<void> loadStockReport({String? branchId}) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return;

    _isLoadingStockReport = true;
    _stockReportError = null;
    notifyListeners();

    try {
      _stockReport = await _reportService.getStockReport(bid);
      _isLoadingStockReport = false;
    } catch (e) {
      debugPrint('ReportProvider.loadStockReport error: $e');
      _stockReportError = 'Gagal memuat laporan stok: ${e.toString()}';
      _isLoadingStockReport = false;
    }
    notifyListeners();
  }

  /// Load profit-loss report for the given date range from SQLite.
  Future<void> loadProfitLossReport({
    String? branchId,
    DateTime? startDate,
    DateTime? endDate,
  }) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return;

    _ensureDateRange();
    final start = startDate ?? _activeDateRange!.start;
    final end = endDate ?? _activeDateRange!.end;

    _isLoadingProfitLoss = true;
    _profitLossError = null;
    notifyListeners();

    try {
      _profitLossReport = await _reportService.getProfitLossReport(
        branchId: bid,
        startDate: start,
        endDate: end,
      );
      _isLoadingProfitLoss = false;
    } catch (e) {
      debugPrint('ReportProvider.loadProfitLoss error: $e');
      _profitLossError = 'Gagal memuat laporan laba rugi: ${e.toString()}';
      _isLoadingProfitLoss = false;
    }
    notifyListeners();
  }

  /// Export report in the given format.
  Future<void> exportReport({
    required String reportType,
    required String format,
    String? branchId,
    DateTime? startDate,
    DateTime? endDate,
  }) async {
    final bid = branchId ?? _currentBranchId;
    if (bid == null) return;

    _ensureDateRange();
    final start = startDate ?? _activeDateRange!.start;
    final end = endDate ?? _activeDateRange!.end;

    _isExporting = true;
    _exportResult = null;
    _exportError = null;
    notifyListeners();

    try {
      final result = await _reportService.exportReport(
        branchId: bid,
        format: format,
        reportType: reportType,
        startDate: start,
        endDate: end,
      );
      _exportResult = result;
      _isExporting = false;
    } catch (e) {
      debugPrint('ReportProvider.exportReport error: $e');
      _exportError = 'Gagal mengexport laporan: ${e.toString()}';
      _isExporting = false;
    }
    notifyListeners();
  }
}
