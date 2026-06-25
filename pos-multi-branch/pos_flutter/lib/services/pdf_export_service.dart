import 'dart:io';
import 'dart:typed_data';
import 'package:flutter/foundation.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:path_provider/path_provider.dart';
import 'package:open_file/open_file.dart';
import '../models/report.dart';
import '../models/transaction.dart';

/// Service for generating and exporting PDF reports.
class PdfExportService {
  static final PdfExportService _instance = PdfExportService._internal();
  factory PdfExportService() => _instance;
  PdfExportService._internal();

  /// Generate and open a sales report PDF.
  Future<String> exportSalesReport(SalesReport report) async {
    final pdf = pw.Document();

    pdf.addPage(
      pw.MultiPage(
        pageFormat: PdfPageFormat.a4,
        margin: const pw.EdgeInsets.all(24),
        header: (context) => _buildHeader(context, 'Laporan Penjualan'),
        footer: (context) => _buildFooter(context),
        build: (context) => [
          _buildDateRange(report.startDate, report.endDate),
          pw.SizedBox(height: 16),
          // Summary cards
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              _buildSummaryBox('Total Penjualan',
                  'Rp ${_formatCurrency(report.totalSales)}', PdfColors.green),
              _buildSummaryBox('Transaksi', '${report.totalTransactions}',
                  PdfColors.blue),
              _buildSummaryBox('Item Terjual', '${report.totalItemsSold}',
                  PdfColors.orange),
            ],
          ),
          pw.SizedBox(height: 8),
          _buildSummaryBox(
            'Rata-rata Transaksi',
            'Rp ${_formatCurrency(report.averageTransactionValue)}',
            PdfColors.purple,
            fullWidth: true,
          ),
          pw.SizedBox(height: 24),
          // Table header
          pw.Header(text: 'Detail Produk Terjual', level: 1),
          pw.SizedBox(height: 8),
          _buildTableHeader(['Produk', 'Qty', 'Total']),
          ...report.items.map((item) => _buildTableRow(
                [item.productName, '${item.quantity}',
                    'Rp ${_formatCurrency(item.total)}'],
                isAlt: report.items.indexOf(item).isOdd,
              )),
          if (report.items.isEmpty)
            pw.Center(
              child: pw.Text('Belum ada data penjualan',
                  style: pw.TextStyle(color: PdfColors.grey)),
            ),
        ],
      ),
    );

    return await _saveAndShare(pdf, 'laporan_penjualan');
  }

  /// Generate and open a stock report PDF.
  Future<String> exportStockReport(StockReport report) async {
    final pdf = pw.Document();

    pdf.addPage(
      pw.MultiPage(
        pageFormat: PdfPageFormat.a4,
        margin: const pw.EdgeInsets.all(24),
        header: (context) => _buildHeader(context, 'Laporan Stok Inventory'),
        footer: (context) => _buildFooter(context),
        build: (context) => [
          pw.Header(
              text: 'Cabang: ${report.branchName}', level: 2),
          pw.SizedBox(height: 8),
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              _buildSummaryBox('Total Produk', '${report.totalProducts}',
                  PdfColors.blue),
              _buildSummaryBox(
                  'Stok Minimal', '${report.lowStockCount}', PdfColors.orange),
              _buildSummaryBox(
                  'Stok Habis', '${report.outOfStockCount}', PdfColors.red),
            ],
          ),
          pw.SizedBox(height: 8),
          _buildSummaryBox(
            'Nilai Stok',
            'Rp ${_formatCurrency(report.totalStockValue.toDouble())}',
            PdfColors.teal,
            fullWidth: true,
          ),
          pw.SizedBox(height: 24),
          pw.Header(text: 'Detail Stok Produk', level: 1),
          pw.SizedBox(height: 8),
          _buildTableHeader(['Produk', 'Stok', 'Min', 'Status', 'Nilai']),
          ...report.products.map((product) {
            final status = product.isOutOfStock
                ? 'Habis'
                : product.isLowStock
                    ? 'Minimal'
                    : 'Aman';
            return _buildTableRow(
              [
                product.productName,
                '${product.currentStock}',
                '${product.minimumStock}',
                status,
                'Rp ${_formatCurrency(product.stockValue)}',
              ],
              isAlt: report.products.indexOf(product).isOdd,
            );
          }),
        ],
      ),
    );

    return await _saveAndShare(pdf, 'laporan_stok');
  }

  /// Generate and open a profit-loss report PDF.
  Future<String> exportProfitLossReport(ProfitLossReport report) async {
    final pdf = pw.Document();

    pdf.addPage(
      pw.MultiPage(
        pageFormat: PdfPageFormat.a4,
        margin: const pw.EdgeInsets.all(24),
        header: (context) => _buildHeader(context, 'Laporan Laba Rugi'),
        footer: (context) => _buildFooter(context),
        build: (context) => [
          _buildDateRange(report.startDate, report.endDate),
          pw.SizedBox(height: 16),
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              _buildSummaryBox('Total Pendapatan',
                  'Rp ${_formatCurrency(report.totalRevenue)}', PdfColors.green),
              _buildSummaryBox(
                  'HPP', 'Rp ${_formatCurrency(report.totalCost)}', PdfColors.red),
            ],
          ),
          pw.SizedBox(height: 8),
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              _buildSummaryBox('Biaya Operasional',
                  'Rp ${_formatCurrency(report.totalExpenses)}', PdfColors.orange),
              _buildSummaryBox('Laba Bersih',
                  'Rp ${_formatCurrency(report.netProfit)}',
                  report.netProfit >= 0 ? PdfColors.green : PdfColors.red),
            ],
          ),
          pw.SizedBox(height: 24),
          pw.Header(text: 'Rincian', level: 1),
          pw.SizedBox(height: 8),
          _buildTableHeader(['Item', 'Kategori', 'Jumlah']),
          ...report.items.map((item) {
            final catLabel = switch (item.category) {
              'revenue' => 'Pendapatan',
              'cost' => 'Biaya',
              'expense' => 'Operasional',
              _ => item.category,
            };
            return _buildTableRow(
              [item.name, catLabel,
                  'Rp ${_formatCurrency(item.amount)}'],
              isAlt: report.items.indexOf(item).isOdd,
            );
          }),
          pw.SizedBox(height: 16),
          pw.Divider(thickness: 2),
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Text('Margin Laba:',
                  style: pw.TextStyle(
                      fontWeight: pw.FontWeight.bold, fontSize: 14)),
              pw.Text(
                '${report.totalRevenue > 0 ? ((report.netProfit / report.totalRevenue) * 100).toStringAsFixed(1) : '0'}%',
                style: pw.TextStyle(
                  fontWeight: pw.FontWeight.bold,
                  fontSize: 14,
                  color: report.netProfit >= 0
                      ? PdfColors.green
                      : PdfColors.red,
                ),
              ),
            ],
          ),
        ],
      ),
    );

    return await _saveAndShare(pdf, 'laporan_laba_rugi');
  }

  /// Helper: build header for each page.
  pw.Widget _buildHeader(pw.Context context, String title) {
    return pw.Column(
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      children: [
        pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
          children: [
            pw.Text('TOKO KAMI - POS Multi Branch',
                style: pw.TextStyle(
                    fontSize: 16, fontWeight: pw.FontWeight.bold)),
            pw.Text('${DateTime.now().day}/${DateTime.now().month}/${DateTime.now().year}',
                style: const pw.TextStyle(fontSize: 10, color: PdfColors.grey)),
          ],
        ),
        pw.Divider(thickness: 1.5),
        pw.SizedBox(height: 8),
        pw.Header(text: title, level: 0),
      ],
    );
  }

  /// Helper: build footer for each page.
  pw.Widget _buildFooter(pw.Context context) {
    return pw.Column(
      children: [
        pw.Divider(),
        pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
          children: [
            pw.Text('POS Multi Branch',
                style: const pw.TextStyle(fontSize: 8, color: PdfColors.grey)),
            pw.Text('Halaman ${context.pageNumber}',
                style: const pw.TextStyle(fontSize: 8, color: PdfColors.grey)),
          ],
        ),
      ],
    );
  }

  /// Helper: build date range text.
  pw.Widget _buildDateRange(DateTime start, DateTime end) {
    return pw.Text(
      'Periode: ${start.day}/${start.month}/${start.year} - '
      '${end.day}/${end.month}/${end.year}',
      style: const pw.TextStyle(fontSize: 10, color: PdfColors.grey),
    );
  }

  /// Helper: build a summary value box.
  pw.Widget _buildSummaryBox(String label, String value, PdfColor color,
      {bool fullWidth = false}) {
    return pw.Container(
      width: fullWidth ? null : 170,
      padding: const pw.EdgeInsets.all(12),
      margin: const pw.EdgeInsets.only(bottom: 8),
      decoration: pw.BoxDecoration(
        color: PdfColor.fromInt(0xFFF5F5F5),
        border: pw.Border(
          left: pw.BorderSide(color: color, width: 4),
        ),
      ),
      child: pw.Column(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Text(label,
              style: const pw.TextStyle(fontSize: 10, color: PdfColors.grey)),
          pw.SizedBox(height: 4),
          pw.Text(value,
              style: pw.TextStyle(
                  fontSize: fullWidth ? 18 : 16,
                  fontWeight: pw.FontWeight.bold,
                  color: color)),
        ],
      ),
    );
  }

  /// Helper: build table header row.
  pw.Widget _buildTableHeader(List<String> columns) {
    return pw.Container(
      decoration: const pw.BoxDecoration(
        color: PdfColors.grey800,
        borderRadius: pw.BorderRadius.all(pw.Radius.circular(4)),
      ),
      padding: const pw.EdgeInsets.symmetric(horizontal: 8, vertical: 8),
      child: pw.Row(
        children: columns
            .map((col) => pw.Expanded(
                  child: pw.Text(col,
                      style: const pw.TextStyle(
                          color: PdfColors.white,
                          fontWeight: pw.FontWeight.bold,
                          fontSize: 10)),
                ))
            .toList(),
      ),
    );
  }

  /// Helper: build a table data row.
  pw.Widget _buildTableRow(List<String> columns, {bool isAlt = false}) {
    return pw.Container(
      decoration: pw.BoxDecoration(
        color: isAlt ? PdfColors.grey100 : null,
        border: pw.Border(
          bottom: pw.BorderSide(color: PdfColors.grey300, width: 0.5),
        ),
      ),
      padding: const pw.EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      child: pw.Row(
        children: columns
            .map((col) => pw.Expanded(
                  child: pw.Text(col,
                      style: const pw.TextStyle(fontSize: 9)),
                ))
            .toList(),
      ),
    );
  }

  /// Save PDF to temp directory and open with system viewer.
  Future<String> _saveAndShare(pw.Document pdf, String filename) async {
    try {
      final dir = await getTemporaryDirectory();
      final filePath = '${dir.path}/${filename}_${DateTime.now().millisecondsSinceEpoch}.pdf';
      final file = File(filePath);
      await file.writeAsBytes(await pdf.save());

      // Open the file using system viewer
      await OpenFile.open(filePath);

      // Also offer share option
      try {
        await Share.shareXFiles(
          [XFile(filePath)],
          subject: filename,
        );
      } catch (_) {
        // Share dialog may be cancelled by user — that's fine
      }

      return 'PDF tersimpan di:\n$filePath';
    } catch (e) {
      debugPrint('PdfExportService._saveAndShare error: $e');
      rethrow;
    }
  }

  String _formatCurrency(double amount) {
    return amount
        .toStringAsFixed(0)
        .replaceAllMapped(
            RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.');
  }
}
