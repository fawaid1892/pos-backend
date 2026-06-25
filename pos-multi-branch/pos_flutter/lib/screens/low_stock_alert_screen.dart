import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/stock_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/theme_provider.dart';
import '../models/stock_adjustment.dart';
import '../widgets/shimmer_loading.dart';
import '../widgets/empty_state_widget.dart';
import '../utils/responsive.dart';

class LowStockAlertScreen extends StatefulWidget {
  const LowStockAlertScreen({super.key});

  @override
  State<LowStockAlertScreen> createState() => _LowStockAlertScreenState();
}

class _LowStockAlertScreenState extends State<LowStockAlertScreen> {
  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadAlerts());
  }

  void _loadAlerts() {
    final auth = context.read<AuthProvider>();
    final stockProv = context.read<StockProvider>();
    final branchId = auth.branchId ?? 'branch_001';
    stockProv.setBranch(branchId);
    stockProv.loadAlerts();
  }

  @override
  Widget build(BuildContext context) {
    final stockProv = context.watch<StockProvider>();
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Peringatan Stok'),
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
          IconButton(
            icon: const Icon(Icons.add_circle_outline),
            tooltip: 'Tambah Stok',
            onPressed: () => Navigator.pushNamed(context, '/stock-adjustment'),
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await context.read<StockProvider>().loadAlerts();
          await context.read<StockProvider>().loadInventory();
        },
        child: _buildAlertList(stockProv, colorScheme),
      ),
    );
  }

  Widget _buildAlertList(StockProvider stockProv, ColorScheme colorScheme) {
    if (stockProv.isLoadingAlerts) {
      return const ShimmerPage(itemCount: 6);
    }

    if (stockProv.alerts.isEmpty) {
      return const EmptyStateWidget.alert();
    }

    final lowStockItems =
        stockProv.alerts.where((a) => a.currentStock > 0).toList();
    final outOfStockItems =
        stockProv.alerts.where((a) => a.currentStock == 0).toList();

    return ListView(
      padding: const EdgeInsets.all(12),
      children: [
        // Summary cards
        Row(
          children: [
            Expanded(
              child: _buildSummaryCard(
                title: 'Stok Habis',
                value: '${outOfStockItems.length}',
                icon: Icons.block,
                color: Colors.red,
              ),
            ),
            const SizedBox(width: 8),
            Expanded(
              child: _buildSummaryCard(
                title: 'Stok Minimal',
                value: '${lowStockItems.length}',
                icon: Icons.warning_amber,
                color: Colors.orange,
              ),
            ),
          ],
        ),
        const SizedBox(height: 16),

        // Out of stock section
        if (outOfStockItems.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              children: [
                Icon(Icons.block, size: 18, color: Colors.red),
                const SizedBox(width: 6),
                Text('Stok Habis (${outOfStockItems.length})',
                    style: const TextStyle(
                        fontWeight: FontWeight.bold, fontSize: 15)),
              ],
            ),
          ),
          ...outOfStockItems
              .map((item) => _buildAlertCard(item, isOutOfStock: true)),
          const SizedBox(height: 12),
        ],

        // Low stock section
        if (lowStockItems.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.only(bottom: 8),
            child: Row(
              children: [
                Icon(Icons.warning_amber, size: 18, color: Colors.orange),
                const SizedBox(width: 6),
                Text('Stok Menipis (${lowStockItems.length})',
                    style: const TextStyle(
                        fontWeight: FontWeight.bold, fontSize: 15)),
              ],
            ),
          ),
          ...lowStockItems
              .map((item) => _buildAlertCard(item, isOutOfStock: false)),
        ],
      ],
    );
  }

  Widget _buildSummaryCard({
    required String title,
    required String value,
    required IconData icon,
    required Color color,
  }) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          children: [
            Icon(icon, color: color, size: 28),
            const SizedBox(height: 8),
            Text(
              value,
              style: TextStyle(
                fontSize: 24,
                fontWeight: FontWeight.bold,
                color: color,
              ),
            ),
            Text(title,
                style: const TextStyle(fontSize: 12, color: Colors.grey)),
          ],
        ),
      ),
    );
  }

  Widget _buildAlertCard(ProductStock item, {required bool isOutOfStock}) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;
    final color = isOutOfStock ? Colors.red : Colors.orange;

    return Card(
      margin: const EdgeInsets.symmetric(vertical: 4),
      color: isDark ? color.shade900 : color.shade50,
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: color,
          child: Icon(
            isOutOfStock ? Icons.block : Icons.warning,
            color: Colors.white,
            size: 20,
          ),
        ),
        title: Text(
          item.productName,
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const SizedBox(height: 4),
            Row(
              children: [
                Text(
                  'Stok: ${item.currentStock}',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: isDark ? color.shade200 : color.shade800,
                  ),
                ),
                const SizedBox(width: 8),
                Text(
                  'Min: ${item.minimumStock}',
                  style: const TextStyle(fontSize: 12, color: Colors.grey),
                ),
              ],
            ),
            // Progress bar
            const SizedBox(height: 6),
            ClipRRect(
              borderRadius: BorderRadius.circular(4),
              child: LinearProgressIndicator(
                value: item.minimumStock > 0
                    ? (item.currentStock / item.minimumStock).clamp(0.0, 1.0)
                    : 0,
                backgroundColor: color.shade100,
                valueColor: AlwaysStoppedAnimation<Color>(color),
                minHeight: 6,
              ),
            ),
          ],
        ),
        trailing: IconButton(
          icon: const Icon(Icons.add_shopping_cart),
          tooltip: 'Tambah Stok',
          onPressed: () => Navigator.pushNamed(
            context,
            '/stock-adjustment',
            arguments: item,
          ),
        ),
      ),
    );
  }
}
