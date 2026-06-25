import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/product.dart';
import '../providers/cart_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/stock_provider.dart';
import '../providers/sync_provider.dart';
import '../providers/theme_provider.dart';
import '../services/product_service.dart';
import '../services/transaction_service.dart';
import '../widgets/product_tile.dart';
import '../widgets/cart_item_tile.dart';
import '../widgets/sync_status_widget.dart';
import '../widgets/shimmer_loading.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/empty_state_widget.dart';
import '../utils/responsive.dart';

class PosScreen extends StatefulWidget {
  const PosScreen({super.key});

  @override
  State<PosScreen> createState() => _PosScreenState();
}

class _PosScreenState extends State<PosScreen> {
  final _searchController = TextEditingController();
  final _productService = ProductService();
  final _transactionService = TransactionService();
  List<Product> _products = [];
  List<Product> _filteredProducts = [];
  bool _isLoadingProducts = true;
  String? _loadError;
  bool _showCart = false;

  @override
  void initState() {
    super.initState();
    _loadProducts();
    _searchController.addListener(_onSearchChanged);
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadProducts() async {
    setState(() {
      _isLoadingProducts = true;
      _loadError = null;
    });
    final auth = context.read<AuthProvider>();
    final branchId = auth.branchId ?? 'branch_001';
    try {
      final products = await _productService.getProducts(branchId);
      if (!mounted) return;
      setState(() {
        _products = products;
        _filteredProducts = products;
        _isLoadingProducts = false;
      });
      // Load stock alerts for badge
      if (mounted) {
        context.read<StockProvider>().loadAlerts(branchId: branchId);
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoadingProducts = false;
        _loadError = e.toString();
      });
    }
  }

  void _onSearchChanged() {
    final query = _searchController.text.trim();
    if (query.isEmpty) {
      setState(() => _filteredProducts = _products);
    } else {
      setState(() {
        _filteredProducts = _products
            .where((p) =>
                p.name.toLowerCase().contains(query.toLowerCase()) ||
                p.barcode.contains(query))
            .toList();
      });
    }
  }

  void _addToCart(Product product) {
    context.read<CartProvider>().addProduct(product);
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text('${product.name} ditambahkan ke keranjang'),
        duration: const Duration(milliseconds: 800),
      ),
    );
  }

  Future<void> _scanBarcode() async {
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(
        content: Text('Barcode scanner akan aktif di perangkat Android'),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final cart = context.watch<CartProvider>();
    final auth = context.watch<AuthProvider>();
    final theme = Theme.of(context);
    final isTablet = context.isTablet;

    return Scaffold(
      appBar: AppBar(
        title: Text(auth.branchName ?? 'POS Multi Branch'),
        actions: [
          // Dark mode toggle
          Consumer<ThemeProvider>(
            builder: (context, themeProv, _) => IconButton(
              icon: Icon(
                themeProv.isDarkMode ? Icons.light_mode : Icons.dark_mode,
              ),
              onPressed: () => themeProv.toggleTheme(),
              tooltip: 'Toggle Dark Mode',
            ),
          ),
          // Sync status icon button
          IconButton(
            icon: const SyncStatusIcon(),
            onPressed: () => Navigator.pushNamed(context, '/sync-status'),
            tooltip: 'Sync Status',
          ),
          // Low stock alert badge
          Consumer<StockProvider>(
            builder: (context, stockProv, _) {
              final alertCount = stockProv.alerts.length;
              return Badge(
                isLabelVisible: alertCount > 0,
                label: Text('$alertCount'),
                child: IconButton(
                  icon: const Icon(Icons.inventory_2),
                  onPressed: () => Navigator.pushNamed(
                      context, '/low-stock-alert'),
                  tooltip: 'Peringatan Stok',
                ),
              );
            },
          ),
          // Manual sync button
          Consumer<SyncProvider>(
            builder: (context, sync, _) => IconButton(
              icon: sync.isSyncing
                  ? const SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : const Icon(Icons.sync),
              onPressed: sync.isSyncing
                  ? null
                  : () async {
                      final result = await sync.triggerSync();
                      if (context.mounted) {
                        ScaffoldMessenger.of(context).showSnackBar(
                          SnackBar(
                            content: Text(result.success
                                ? 'Sync selesai: ${result.summary}'
                                : 'Sync gagal: ${result.error ?? ''}'),
                            backgroundColor:
                                result.success ? Colors.green : Colors.red,
                            duration: const Duration(seconds: 3),
                          ),
                        );
                      }
                    },
            ),
          ),
          IconButton(
            icon: Icon(_showCart ? Icons.search : Icons.shopping_cart),
            onPressed: () => setState(() => _showCart = !_showCart),
          ),
          if (_showCart && !cart.isEmpty)
            Badge(
              label: Text('${cart.itemCount}'),
              child: IconButton(
                icon: const Icon(Icons.point_of_sale),
                onPressed: () => Navigator.pushNamed(context, '/checkout'),
              ),
            ),
          IconButton(
            icon: const Icon(Icons.logout),
            onPressed: () async {
              await auth.logout();
              if (context.mounted) {
                Navigator.pushReplacementNamed(context, '/login');
              }
            },
          ),
        ],
      ),
      body: _showCart ? _buildCartView(cart) : _buildProductView(cart, isTablet),
    );
  }

  Widget _buildProductView(CartProvider cart, bool isTablet) {
    return Column(
      children: [
        // Search bar
        Padding(
          padding: EdgeInsets.all(isTablet ? 16 : 12),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    hintText: 'Cari produk (nama / barcode)',
                    prefixIcon: const Icon(Icons.search),
                    suffixIcon: _searchController.text.isNotEmpty
                        ? IconButton(
                            icon: const Icon(Icons.clear),
                            onPressed: () {
                              _searchController.clear();
                            },
                          )
                        : IconButton(
                            icon: const Icon(Icons.qr_code_scanner),
                            onPressed: _scanBarcode,
                          ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                    filled: true,
                    fillColor: theme.colorScheme.surfaceVariant,
                  ),
                ),
              ),
            ],
          ),
        ),

        // Products list / grid
        Expanded(
          child: _buildProductList(cart),
        ),

        // Bottom cart bar
        if (!cart.isEmpty)
          SafeArea(
            child: Container(
              padding: EdgeInsets.all(isTablet ? 16 : 12),
              decoration: BoxDecoration(
                color: theme.colorScheme.surface,
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withOpacity(0.1),
                    blurRadius: 8,
                    offset: const Offset(0, -2),
                  ),
                ],
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Text(
                          '${cart.itemCount} item',
                          style: TextStyle(
                            color: theme.colorScheme.onSurfaceVariant,
                          ),
                        ),
                        Text(
                          'Rp ${_formatPrice(cart.grandTotal)}',
                          style: TextStyle(
                            fontSize: isTablet ? 22 : 20,
                            fontWeight: FontWeight.bold,
                            color: theme.colorScheme.onSurface,
                          ),
                        ),
                      ],
                    ),
                  ),
                  ElevatedButton.icon(
                    onPressed: () => Navigator.pushNamed(context, '/checkout'),
                    icon: const Icon(Icons.shopping_cart_checkout),
                    label: const Text('Bayar'),
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 24, vertical: 12),
                    ),
                  ),
                ],
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildProductList(CartProvider cart) {
    if (_isLoadingProducts) {
      return const ShimmerPage(itemCount: 8);
    }

    if (_loadError != null) {
      return ErrorStateWidget(
        message: _loadError!,
        title: 'Gagal memuat produk',
        onRetry: _loadProducts,
      );
    }

    if (_filteredProducts.isEmpty) {
      return const EmptyStateWidget.search();
    }

    final isTablet = context.isTablet;

    if (isTablet) {
      // Grid layout for tablet
      return RefreshIndicator(
        onRefresh: _loadProducts,
        child: GridView.builder(
          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
          gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
            crossAxisCount: 3,
            childAspectRatio: 1.2,
            crossAxisSpacing: 8,
            mainAxisSpacing: 8,
          ),
          itemCount: _filteredProducts.length,
          itemBuilder: (context, index) => _buildProductGridTile(
              _filteredProducts[index]),
        ),
      );
    }

    return RefreshIndicator(
      onRefresh: _loadProducts,
      child: ListView.builder(
        itemCount: _filteredProducts.length,
        itemBuilder: (context, index) => ProductTile(
          product: _filteredProducts[index],
          onAdd: () => _addToCart(_filteredProducts[index]),
        ),
      ),
    );
  }

  Widget _buildProductGridTile(Product product) {
    return Card(
      child: InkWell(
        borderRadius: BorderRadius.circular(12),
        onTap: () => _addToCart(product),
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              CircleAvatar(
                backgroundColor: theme.colorScheme.primaryContainer,
                child: Text(
                  product.name[0].toUpperCase(),
                  style: TextStyle(
                    color: theme.colorScheme.onPrimaryContainer,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              const SizedBox(height: 6),
              Text(
                product.name,
                style: const TextStyle(
                    fontWeight: FontWeight.w600, fontSize: 13),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
              const Spacer(),
              Text(
                'Rp ${_formatPrice(product.price)}',
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 13,
                  color: theme.colorScheme.primary,
                ),
              ),
              if (product.stock <= 5)
                Text(
                  'Stok: ${product.stock}',
                  style: TextStyle(
                    fontSize: 11,
                    color: theme.colorScheme.error,
                  ),
                ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCartView(CartProvider cart) {
    if (cart.isEmpty) {
      return const EmptyStateWidget.cart();
    }

    final isTablet = context.isTablet;

    return Column(
      children: [
        Expanded(
          child: ListView.builder(
            itemCount: cart.items.length,
            itemBuilder: (context, index) {
              final item = cart.items[index];
              return CartItemTile(
                item: item,
                onIncrement: () =>
                    cart.updateQuantity(item.productId, item.quantity + 1),
                onDecrement: () =>
                    cart.updateQuantity(item.productId, item.quantity - 1),
                onRemove: () => cart.removeItem(item.productId),
              );
            },
          ),
        ),
        Container(
          padding: EdgeInsets.all(isTablet ? 20 : 16),
          decoration: BoxDecoration(
            color: theme.colorScheme.surface,
            boxShadow: [
              BoxShadow(
                color: Colors.black.withOpacity(0.1),
                blurRadius: 8,
                offset: const Offset(0, -2),
              ),
            ],
          ),
          child: Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Total:', style: TextStyle(fontSize: 16)),
                  Text(
                    'Rp ${_formatPrice(cart.total)}',
                    style: const TextStyle(
                        fontSize: 16, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              if (cart.discountTotal > 0)
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Diskon:', style: TextStyle(fontSize: 16)),
                    Text(
                      '- Rp ${_formatPrice(cart.discountTotal)}',
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                        color: theme.colorScheme.error,
                      ),
                    ),
                  ],
                ),
              const Divider(),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Grand Total:',
                      style: TextStyle(
                          fontSize: 18, fontWeight: FontWeight.bold)),
                  Text(
                    'Rp ${_formatPrice(cart.grandTotal)}',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: theme.colorScheme.primary,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: ElevatedButton.icon(
                  onPressed: () => Navigator.pushNamed(context, '/checkout'),
                  icon: const Icon(Icons.payment),
                  label: const Text('Lanjut ke Pembayaran',
                      style: TextStyle(fontSize: 16)),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(vertical: 14),
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  String _formatPrice(double amount) {
    return amount
        .toStringAsFixed(0)
        .replaceAllMapped(
            RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.');
  }
}
