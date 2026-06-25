import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/report_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/theme_provider.dart';
import '../utils/responsive.dart';

class ExportScreen extends StatefulWidget {
  const ExportScreen({super.key});

  @override
  State<ExportScreen> createState() => _ExportScreenState();
}

class _ExportScreenState extends State<ExportScreen> {
  String _selectedReportType = 'sales';
  String _selectedFormat = 'xlsx';
  DateTimeRange? _dateRange;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final reportProv = context.read<ReportProvider>();
      if (reportProv.activeDateRange != null) {
        setState(() => _dateRange = reportProv.activeDateRange);
      }
    });
  }

  Future<void> _pickDateRange() async {
    final now = DateTime.now();
    final picked = await showDateRangePicker(
      context: context,
      firstDate: DateTime(2024),
      lastDate: now,
      initialDateRange: _dateRange ??
          DateTimeRange(
            start: DateTime(now.year, now.month, 1),
            end: now,
          ),
    );
    if (picked != null) {
      setState(() => _dateRange = picked);
    }
  }

  Future<void> _export() async {
    if (_dateRange == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pilih rentang tanggal terlebih dahulu')),
      );
      return;
    }

    final reportProv = context.read<ReportProvider>();
    await reportProv.exportReport(
      reportType: _selectedReportType,
      format: _selectedFormat,
      startDate: _dateRange!.start,
      endDate: _dateRange!.end,
    );

    if (!mounted) return;

    if (reportProv.exportResult != null) {
      _showSuccessDialog(reportProv.exportResult!);
    } else if (reportProv.exportError != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(reportProv.exportError!),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  void _showSuccessDialog(String message) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        icon: const Icon(Icons.check_circle, color: Colors.green, size: 48),
        title: const Text('Export Berhasil'),
        content: Text(message),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.pop(context);
              Navigator.pop(context);
            },
            child: const Text('OK'),
          ),
        ],
      ),
    );
  }

  String _reportTypeLabel(String type) {
    switch (type) {
      case 'sales':
        return 'Laporan Penjualan';
      case 'stock':
        return 'Laporan Stok';
      case 'profit_loss':
        return 'Laporan Laba Rugi';
      default:
        return type;
    }
  }

  IconData _reportTypeIcon(String type) {
    switch (type) {
      case 'sales':
        return Icons.shopping_cart;
      case 'stock':
        return Icons.inventory_2;
      case 'profit_loss':
        return Icons.trending_up;
      default:
        return Icons.description;
    }
  }

  @override
  Widget build(BuildContext context) {
    final reportProv = context.watch<ReportProvider>();
    final colorScheme = Theme.of(context).colorScheme;
    final isTablet = context.isTablet;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Export Laporan'),
        actions: [
          Consumer<ThemeProvider>(
            builder: (context, themeProv, _) => IconButton(
              icon: Icon(
                themeProv.isDarkMode ? Icons.light_mode : Icons.dark_mode,
              ),
              onPressed: () => themeProv.toggleTheme(),
              tooltip: 'Toggle Dark Mode',
            ),
          ),
        ],
      ),
      body: ListView(
        padding: EdgeInsets.all(isTablet ? 24 : 16),
        children: [
          // Report type
          const Text('Jenis Laporan',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          ...['sales', 'stock', 'profit_loss'].map((type) {
            final selected = _selectedReportType == type;
            return Card(
              color: selected ? colorScheme.primaryContainer : null,
              child: RadioListTile<String>(
                title: Row(
                  children: [
                    Icon(
                      _reportTypeIcon(type),
                      size: 20,
                      color:
                          selected ? colorScheme.primary : null,
                    ),
                    const SizedBox(width: 8),
                    Text(_reportTypeLabel(type)),
                  ],
                ),
                value: type,
                groupValue: _selectedReportType,
                onChanged: (val) {
                  setState(() => _selectedReportType = val!);
                },
              ),
            );
          }),

          const SizedBox(height: 24),

          // Format
          const Text('Format File',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: _buildFormatCard(
                  label: 'PDF',
                  icon: Icons.picture_as_pdf,
                  color: Colors.red,
                  selected: _selectedFormat == 'pdf',
                  onTap: () => setState(() => _selectedFormat = 'pdf'),
                ),
              ),
              const SizedBox(width: 12),
              Expanded(
                child: _buildFormatCard(
                  label: 'Excel (XLSX)',
                  icon: Icons.table_chart,
                  color: Colors.green,
                  selected: _selectedFormat == 'xlsx',
                  onTap: () => setState(() => _selectedFormat = 'xlsx'),
                ),
              ),
            ],
          ),

          const SizedBox(height: 24),

          // Date range
          const Text('Rentang Tanggal',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          Card(
            child: ListTile(
              leading: const Icon(Icons.date_range),
              title: Text(
                _dateRange != null
                    ? '${_dateRange!.start.day}/${_dateRange!.start.month}/${_dateRange!.start.year} '
                        '- ${_dateRange!.end.day}/${_dateRange!.end.month}/${_dateRange!.end.year}'
                    : 'Pilih rentang tanggal',
              ),
              trailing: const Icon(Icons.chevron_right),
              onTap: _pickDateRange,
            ),
          ),

          const SizedBox(height: 32),

          // Export button
          SizedBox(
            height: 52,
            child: ElevatedButton.icon(
              onPressed: reportProv.isExporting ? null : _export,
              icon: reportProv.isExporting
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.file_download),
              label: Text(
                reportProv.isExporting
                    ? 'Mengexport...'
                    : 'Export ${_reportTypeLabel(_selectedReportType)}',
                style: const TextStyle(fontSize: 16),
              ),
              style: ElevatedButton.styleFrom(
                backgroundColor: colorScheme.primary,
                foregroundColor: colorScheme.onPrimary,
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFormatCard({
    required String label,
    required IconData icon,
    required Color color,
    required bool selected,
    required VoidCallback onTap,
  }) {
    final brightness = Theme.of(context).brightness;

    return GestureDetector(
      onTap: onTap,
      child: Card(
        color: selected
            ? color.withOpacity(0.1)
            : null,
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
          side: BorderSide(
            color: selected ? color : Colors.grey.shade300,
            width: selected ? 2 : 1,
          ),
        ),
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: 20),
          child: Column(
            children: [
              Icon(icon,
                  color: selected ? color : Colors.grey,
                  size: 40),
              const SizedBox(height: 8),
              Text(
                label,
                style: TextStyle(
                  fontWeight:
                      selected ? FontWeight.bold : FontWeight.normal,
                  color: selected ? color : Colors.grey,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
