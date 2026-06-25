import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/stock_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/sync_provider.dart';
import '../providers/theme_provider.dart';
import '../models/stock_adjustment.dart';
import '../widgets/sync_status_widget.dart';
import '../widgets/shimmer_loading.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/empty_state_widget.dart';
import '../utils/responsive.dart';

class StockScreen extends StatefulWidget {
  const StockScreen({super.key});

  @override
  State<StockScreen> createState() => _StockScreenState();
}

class _StockScreenState extends State<StockScreen>
    with SingleTickerProviderStateMixin {
  final _searchController = TextEditingController();
  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _searchController.addListener(_onSearchChanged);
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadData());
  }

  @override
  void dispose() {
    _searchController.dispose();
    _tabController.dispose();
    super.dispose();
  }

  void _loadData() {
    final auth = context.read<AuthProvider>();
    final stockProv = context.read<StockProvider>();
    final branchId = auth.branchId ?? 'branch_001';
    stockProv.setBranch(branchId);
    stockProv.loadInventory();
    stockProv.loadAlerts();
  }

  void _onSearchChanged() {
    context.read<StockProvider>().searchInventory(_searchController.text);
  }

  @override
  Widget build(BuildContext context) {
    final stockProv = context.watch<StockProvider>();
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Stok Inventory'),
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
            icon: const SyncStatusIcon(),
            onPressed: () => Navigator.pushNamed(context, '/sync-status'),
            tooltip: 'Sync Status',
          ),
          IconButton(
            icon: const Icon(Icons.add_circle_outline),
            tooltip: 'Adjustment Stok',
            onPressed: () => Navigator.pushNamed(context, '/stock-adjustment'),
          ),
          IconButton(
            icon: const Icon(Icons.swap_horiz),
            tooltip: 'Transfer Stok',
            onPressed: () => Navigator.pushNamed(context, '/stock-transfer'),
          ),
        ],
        bottom: TabBar(
          controller: _tabController,
          tabs: [
            const Tab(text: 'Inventory'),
            Tab(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Text('Alert'),
                  if (stockProv.alerts.isNotEmpty) ...[
                    const SizedBox(width: 4),
                    Container(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: colorScheme.error,
                        borderRadius: BorderRadius.circular(10),
                      ),
                      child: Text(
                        '${stockProv.alerts.length}',
                        style: TextStyle(
                          color: colorScheme.onError,
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [
          _buildInventoryTab(stockProv),
          _buildAlertTab(stockProv),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        heroTag: 'stock_adjust',
        onPressed: () => Navigator.pushNamed(context, '/stock-adjustment'),
        child: const Icon(Icons.add),
      ),
    );
  }

  Widget _buildInventoryTab(StockProvider stockProv) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
          child: TextField(
            controller: _searchController,
            decoration: InputDecoration(
              hintText: 'Cari produk...',
              prefixIcon: const Icon(Icons.search),
              suffixIcon: _searchController.text.isNotEmpty
                  ? IconButton(
                      icon: const Icon(Icons.clear),
                      onPressed: () => _searchController.clear(),
                    )
                  : null,
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              filled: true,
              fillColor: Theme.of(context).colorScheme.surfaceVariant,
            ),
          ),
        ),
        Expanded(
          child: _buildInventoryContent(stockProv),
        ),
      ],
    );
  }

  Widget _buildInventoryContent(StockProvider stockProv) {
    if (stockProv.isLoadingInventory) {
      return const ShimmerPage(itemCount: 8);
    }

    if (stockProv.inventoryError != null) {
      return ErrorStateWidget(
        message: stockProv.inventoryError!,
        title: 'Gagal memuat inventory',
        onRetry: () => stockProv.loadInventory(),
      );
    }

    if (stockProv.inventory.isEmpty) {
      return const EmptyStateWidget.stock();
    }

    return RefreshIndicator(
      onRefresh: () => context.read<StockProvider>().loadInventory(),
      child: ListView.builder(
        itemCount: stockProv.inventory.length,
        itemBuilder: (context, index) {
          final item = stockProv.inventory[index];
          return _buildStockTile(item);
        },
      ),
    );
  }

  Widget _buildAlertTab(StockProvider stockProv) {
    if (stockProv.isLoadingAlerts) {
      return const ShimmerPage(itemCount: 4);
    }

    if (stockProv.alerts.isEmpty) {
      return const EmptyStateWidget.alert();
    }

    return ListView.builder(
      itemCount: stockProv.alerts.length,
      itemBuilder: (context, index) {
        final item = stockProv.alerts[index];
        return _buildAlertTile(item);
      },
    );
  }

  Widget _buildStockTile(ProductStock item) {
    final isLow = item.currentStock <= item.minimumStock;
    final isOut = item.currentStock == 0;

    Color statusColor;
    IconData statusIcon;
    String statusText;

    if (isOut) {
      statusColor = Colors.red;
      statusIcon = Icons.block;
      statusText = 'Habis';
    } else if (isLow) {
      statusColor = Colors.orange;
      statusIcon = Icons.warning_amber;
      statusText = 'Hampir Habis';
    } else {
      statusColor = Colors.green;
      statusIcon = Icons.check_circle;
      statusText = 'Aman';
    }

    final theme = Theme.of(context);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: statusColor.withOpacity(0.15),
          child: Icon(statusIcon, color: statusColor),
        ),
        title: Text(item.productName,
            style: const TextStyle(fontWeight: FontWeight.w500)),
        subtitle: Row(
          children: [
            Text('Stok: ${item.currentStock}'),
            const SizedBox(width: 8),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
              decoration: BoxDecoration(
                color: statusColor.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8),
              ),
              child: Text(statusText,
                  style: TextStyle(color: statusColor, fontSize: 11)),
            ),
          ],
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            IconButton(
              icon: const Icon(Icons.edit_note, size: 20),
              tooltip: 'Adjustment',
              onPressed: () => Navigator.pushNamed(
                context,
                '/stock-adjustment',
                arguments: item,
              ),
            ),
            IconButton(
              icon: const Icon(Icons.swap_horiz, size: 20),
              tooltip: 'Transfer',
              onPressed: () => Navigator.pushNamed(
                context,
                '/stock-transfer',
                arguments: item,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAlertTile(ProductStock item) {
    final theme = Theme.of(context);
    final isDark = theme.brightness == Brightness.dark;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      color: isDark
          ? (item.currentStock == 0
              ? Colors.red.shade900
              : Colors.orange.shade900)
          : (item.currentStock == 0
              ? Colors.red.shade50
              : Colors.orange.shade50),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor:
              item.currentStock == 0 ? Colors.red : Colors.orange,
          child: Icon(
            item.currentStock == 0 ? Icons.block : Icons.warning,
            color: Colors.white,
          ),
        ),
        title: Text(item.productName,
            style: const TextStyle(fontWeight: FontWeight.w500)),
        subtitle: Text(
          item.currentStock == 0
              ? 'Stok habis!'
              : 'Stok: ${item.currentStock} (min: ${item.minimumStock})',
          style: TextStyle(
            color: isDark
                ? (item.currentStock == 0
                    ? Colors.red.shade200
                    : Colors.orange.shade200)
                : (item.currentStock == 0
                    ? Colors.red
                    : Colors.orange.shade800),
          ),
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
