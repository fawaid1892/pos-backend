import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/report.dart';
import '../providers/report_provider.dart';
import '../providers/auth_provider.dart';

class ReportScreen extends StatefulWidget {
  const ReportScreen({super.key});

  @override
  State<ReportScreen> createState() => _ReportScreenState();
}

class _ReportScreenState extends State<ReportScreen>
    with SingleTickerProviderStateMixin {
  late TabController _tabController;
  DateTimeRange? _dateRange;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _tabController.addListener(_onTabChanged);
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadInitialData());
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _onTabChanged() {
    if (!_tabController.indexIsChanging) {
      _loadCurrentTab();
    }
  }

  void _loadInitialData() {
    final auth = context.read<AuthProvider>();
    final reportProv = context.read<ReportProvider>();
    final branchId = auth.branchId ?? 'branch_001';
    reportProv.setBranch(branchId);

    final now = DateTime.now();
    _dateRange = DateTimeRange(
      start: DateTime(now.year, now.month, 1),
      end: now,
    );
    reportProv.setDateRange(_dateRange!);

    _loadCurrentTab();
  }

  void _loadCurrentTab() {
    final reportProv = context.read<ReportProvider>();
    switch (_tabController.index) {
      case 0:
        reportProv.loadSalesReport();
        break;
      case 1:
        reportProv.loadStockReport();
        break;
      case 2:
        reportProv.loadProfitLossReport();
        break;
    }
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
      final reportProv = context.read<ReportProvider>();
      reportProv.setDateRange(picked);
      _loadCurrentTab();
    }
  }

  @override
  Widget build(BuildContext context) {
    final reportProv = context.watch<ReportProvider>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Laporan'),
        actions: [
          // Date range picker
          TextButton.icon(
            onPressed: _pickDateRange,
            icon: const Icon(Icons.date_range),
            label: Text(
              _dateRange != null
                  ? '${_dateRange!.start.day}/${_dateRange!.start.month} '
                      '- ${_dateRange!.end.day}/${_dateRange!.end.month}'
                  : 'Filter Tanggal',
              style: const TextStyle(fontSize: 12),
            ),
          ),
          // Export button
          IconButton(
            icon: const Icon(Icons.file_download),
            tooltip: 'Export Laporan',
            onPressed: () => Navigator.pushNamed(context, '/export-report'),
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          tabs: const [
            Tab(icon: Icon(Icons.shopping_cart), text: 'Penjualan'),
            Tab(icon: Icon(Icons.inventory_2), text: 'Stok'),
            Tab(icon: Icon(Icons.trending_up), text: 'Laba Rugi'),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildSalesReportTab(reportProv),
          _buildStockReportTab(reportProv),
          _buildProfitLossTab(reportProv),
        ],
      ),
    );
  }

  // ─────────────────────────────
  // SALES REPORT TAB
  // ─────────────────────────────
  Widget _buildSalesReportTab(ReportProvider reportProv) {
    if (reportProv.isLoadingSalesReport) {
      return const Center(child: CircularProgressIndicator());
    }

    final report = reportProv.salesReport;
    if (report == null) {
      return const Center(child: Text('Pilih tanggal untuk melihat laporan'));
    }

    return RefreshIndicator(
      onRefresh: () => reportProv.loadSalesReport(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Summary cards
          _buildSummaryCard(
            title: 'Total Penjualan',
            value: _formatCurrency(report.totalSales),
            icon: Icons.attach_money,
            color: Colors.green,
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: _buildSmallSummaryCard(
                  title: 'Transaksi',
                  value: '${report.totalTransactions}',
                  icon: Icons.receipt,
                  color: Colors.blue,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: _buildSmallSummaryCard(
                  title: 'Item Terjual',
                  value: '${report.totalItemsSold}',
                  icon: Icons.shopping_bag,
                  color: Colors.orange,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          _buildSummaryCard(
            title: 'Rata-rata Transaksi',
            value: _formatCurrency(report.averageTransactionValue),
            icon: Icons.analytics,
            color: Colors.purple,
          ),

          const SizedBox(height: 16),

          // Detail items
          const Text('Detail Produk Terjual',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          ...report.items.map((item) => Card(
                child: ListTile(
                  leading: CircleAvatar(
                    child: Text('${item.quantity}',
                        style: const TextStyle(fontSize: 14)),
                  ),
                  title: Text(item.productName),
                  trailing: Text(
                    _formatCurrency(item.total),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
              )),
          if (report.items.isEmpty)
            const Card(
              child: Padding(
                padding: EdgeInsets.all(24),
                child: Center(child: Text('Belum ada data penjualan')),
              ),
            ),
        ],
      ),
    );
  }

  // ─────────────────────────────
  // STOCK REPORT TAB
  // ─────────────────────────────
  Widget _buildStockReportTab(ReportProvider reportProv) {
    if (reportProv.isLoadingStockReport) {
      return const Center(child: CircularProgressIndicator());
    }

    final report = reportProv.stockReport;
    if (report == null) {
      return const Center(child: Text('Memuat laporan stok...'));
    }

    return RefreshIndicator(
      onRefresh: () => reportProv.loadStockReport(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // Summary
          Row(
            children: [
              Expanded(
                child: _buildSmallSummaryCard(
                  title: 'Total Produk',
                  value: '${report.totalProducts}',
                  icon: Icons.inventory,
                  color: Colors.blue,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: _buildSmallSummaryCard(
                  title: 'Stok Minimal',
                  value: '${report.lowStockCount}',
                  icon: Icons.warning_amber,
                  color: Colors.orange,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: _buildSmallSummaryCard(
                  title: 'Stok Habis',
                  value: '${report.outOfStockCount}',
                  icon: Icons.block,
                  color: Colors.red,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: _buildSmallSummaryCard(
                  title: 'Nilai Stok',
                  value: _formatCurrency(report.totalStockValue.toDouble()),
                  icon: Icons.account_balance,
                  color: Colors.teal,
                  smallFont: true,
                ),
              ),
            ],
          ),

          const SizedBox(height: 16),

          // Product stock list
          const Text('Detail Stok Produk',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          ...report.products.map((product) {
            Color statusColor;
            IconData statusIcon;

            if (product.isOutOfStock) {
              statusColor = Colors.red;
              statusIcon = Icons.block;
            } else if (product.isLowStock) {
              statusColor = Colors.orange;
              statusIcon = Icons.warning_amber;
            } else {
              statusColor = Colors.green;
              statusIcon = Icons.check_circle;
            }

            return Card(
              child: ListTile(
                leading: CircleAvatar(
                  backgroundColor: statusColor.withOpacity(0.15),
                  child: Icon(statusIcon, color: statusColor, size: 20),
                ),
                title: Text(product.productName),
                subtitle: Row(
                  children: [
                    Text('Stok: ${product.currentStock}'),
                    const SizedBox(width: 8),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: statusColor.withOpacity(0.1),
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        product.isOutOfStock
                            ? 'Habis'
                            : product.isLowStock
                                ? 'Minimal'
                                : 'Aman',
                        style:
                            TextStyle(color: statusColor, fontSize: 11),
                      ),
                    ),
                  ],
                ),
                trailing: Text(
                  _formatCurrency(product.stockValue),
                  style: const TextStyle(fontSize: 12, color: Colors.grey),
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  // ─────────────────────────────
  // PROFIT/LOSS TAB
  // ─────────────────────────────
  Widget _buildProfitLossTab(ReportProvider reportProv) {
    if (reportProv.isLoadingProfitLoss) {
      return const Center(child: CircularProgressIndicator());
    }

    final report = reportProv.profitLossReport;
    if (report == null) {
      return const Center(
          child: Text('Pilih tanggal untuk melihat laporan laba rugi'));
    }

    return RefreshIndicator(
      onRefresh: () => reportProv.loadProfitLossReport(),
      child: ListView(
        padding: const EdgeInsets.all(16),
        children: [
          // P&L Summary cards
          Card(
            color: Colors.green.shade50,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  const Text('Total Pendapatan',
                      style: TextStyle(color: Colors.grey)),
                  const SizedBox(height: 4),
                  Text(
                    _formatCurrency(report.totalRevenue),
                    style: const TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                      color: Colors.green,
                    ),
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),

          Card(
            color: Colors.red.shade50,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceEvenly,
                children: [
                  Column(
                    children: [
                      const Text('HPP', style: TextStyle(color: Colors.grey)),
                      Text(
                        _formatCurrency(report.totalCost),
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.red,
                        ),
                      ),
                    ],
                  ),
                  Container(height: 40, width: 1, color: Colors.grey.shade300),
                  Column(
                    children: [
                      const Text('Biaya Operasional',
                          style: TextStyle(color: Colors.grey)),
                      Text(
                        _formatCurrency(report.totalExpenses),
                        style: const TextStyle(
                          fontWeight: FontWeight.bold,
                          color: Colors.red,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          const SizedBox(height: 8),

          Card(
            color: report.netProfit >= 0
                ? Colors.green.shade50
                : Colors.red.shade50,
            child: Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                children: [
                  const Text('Laba Bersih',
                      style: TextStyle(color: Colors.grey)),
                  const SizedBox(height: 4),
                  Text(
                    _formatCurrency(report.netProfit),
                    style: TextStyle(
                      fontSize: 28,
                      fontWeight: FontWeight.bold,
                      color: report.netProfit >= 0 ? Colors.green : Colors.red,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    report.netProfit >= 0
                        ? '${_profitMargin(report.netProfit, report.totalRevenue)}% margin'
                        : 'Rugi',
                    style: TextStyle(
                      color: report.netProfit >= 0
                          ? Colors.green.shade700
                          : Colors.red.shade700,
                    ),
                  ),
                ],
              ),
            ),
          ),

          const SizedBox(height: 24),

          // Detailed breakdown
          const Text('Rincian',
              style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
          const SizedBox(height: 8),
          ...report.items.map((item) {
            Color itemColor;
            IconData itemIcon;
            switch (item.category) {
              case 'revenue':
                itemColor = Colors.green;
                itemIcon = Icons.arrow_upward;
                break;
              case 'cost':
                itemColor = Colors.red;
                itemIcon = Icons.arrow_downward;
                break;
              case 'expense':
                itemColor = Colors.orange;
                itemIcon = Icons.money_off;
                break;
              default:
                itemColor = Colors.grey;
                itemIcon = Icons.help;
            }

            return Card(
              child: ListTile(
                leading: CircleAvatar(
                  backgroundColor: itemColor.withOpacity(0.15),
                  child: Icon(itemIcon, color: itemColor, size: 20),
                ),
                title: Text(item.name),
                trailing: Text(
                  _formatCurrency(item.amount),
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: itemColor,
                  ),
                ),
              ),
            );
          }),
        ],
      ),
    );
  }

  // ─────────────────────────────
  // REUSABLE WIDGETS
  // ─────────────────────────────
  Widget _buildSummaryCard({
    required String title,
    required String value,
    required IconData icon,
    required Color color,
  }) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Row(
          children: [
            CircleAvatar(
              backgroundColor: color.withOpacity(0.15),
              child: Icon(icon, color: color),
            ),
            const SizedBox(width: 16),
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: const TextStyle(color: Colors.grey)),
                Text(value,
                    style: const TextStyle(
                        fontSize: 20, fontWeight: FontWeight.bold)),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSmallSummaryCard({
    required String title,
    required String value,
    required IconData icon,
    required Color color,
    bool smallFont = false,
  }) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          children: [
            Icon(icon, color: color, size: 24),
            const SizedBox(height: 8),
            Text(
              value,
              style: TextStyle(
                fontSize: smallFont ? 14 : 18,
                fontWeight: FontWeight.bold,
              ),
            ),
            Text(title,
                style: const TextStyle(fontSize: 11, color: Colors.grey)),
          ],
        ),
      ),
    );
  }

  String _profitMargin(double profit, double revenue) {
    if (revenue == 0) return '0';
    return ((profit / revenue) * 100).toStringAsFixed(1);
  }

  String _formatCurrency(double amount) {
    return 'Rp ${amount.toStringAsFixed(0).replaceAllMapped(RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.')}';
  }
}
