import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/product.dart';
import '../providers/cart_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/sync_provider.dart';
import '../services/mock_api_service.dart';
import '../widgets/product_tile.dart';
import '../widgets/cart_item_tile.dart';
import '../widgets/sync_status_widget.dart';

class PosScreen extends StatefulWidget {
  const PosScreen({super.key});

  @override
  State<PosScreen> createState() => _PosScreenState();
}

class _PosScreenState extends State<PosScreen> {
  final _searchController = TextEditingController();
  final _api = MockApiService();
  List<Product> _products = [];
  List<Product> _filteredProducts = [];
  bool _isLoadingProducts = true;
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
    setState(() => _isLoadingProducts = true);
    final auth = context.read<AuthProvider>();
    final branchId = auth.branchId ?? 'branch_001';
    final products = await _api.getProducts(branchId);
    if (!mounted) return;
    setState(() {
      _products = products;
      _filteredProducts = products;
      _isLoadingProducts = false;
    });
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
    // barcode_scan2 integration placeholder
    // In production:
    //   import 'package:barcode_scan2/barcode_scan2.dart';
    //   final result = await BarcodeScanner.scan();
    //   final product = await _api.getProductByBarcode(result.rawContent, branchId);
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

    return Scaffold(
      appBar: AppBar(
        title: Text(auth.branchName ?? 'POS Multi Branch'),
        actions: [
          // Sync status icon button
          IconButton(
            icon: const SyncStatusIcon(),
            onPressed: () => Navigator.pushNamed(context, '/sync-status'),
            tooltip: 'Sync Status',
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
      body: _showCart ? _buildCartView(cart) : _buildProductView(cart),
    );
  }

  Widget _buildProductView(CartProvider cart) {
    return Column(
      children: [
        Padding(
          padding: const EdgeInsets.all(12),
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
                    fillColor: Theme.of(context).colorScheme.surfaceVariant,
                  ),
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: _isLoadingProducts
              ? const Center(child: CircularProgressIndicator())
              : _filteredProducts.isEmpty
                  ? const Center(child: Text('Produk tidak ditemukan'))
                  : RefreshIndicator(
                      onRefresh: _loadProducts,
                      child: ListView.builder(
                        itemCount: _filteredProducts.length,
                        itemBuilder: (context, index) => ProductTile(
                          product: _filteredProducts[index],
                          onAdd: () => _addToCart(_filteredProducts[index]),
                        ),
                      ),
                    ),
        ),
        if (!cart.isEmpty)
          SafeArea(
            child: Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
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
                          style: const TextStyle(color: Colors.grey),
                        ),
                        Text(
                          'Rp ${_formatPrice(cart.grandTotal)}',
                          style: const TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
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

  Widget _buildCartView(CartProvider cart) {
    if (cart.isEmpty) {
      return const Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.shopping_cart_outlined, size: 80, color: Colors.grey),
            SizedBox(height: 16),
            Text('Keranjang kosong', style: TextStyle(color: Colors.grey)),
          ],
        ),
      );
    }

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
          padding: const EdgeInsets.all(16),
          decoration: BoxDecoration(
            color: Theme.of(context).colorScheme.surface,
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
                        color: Theme.of(context).colorScheme.error,
                      ),
                    ),
                  ],
                ),
              const Divider(),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  const Text('Grand Total:',
                      style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                  Text(
                    'Rp ${_formatPrice(cart.grandTotal)}',
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Theme.of(context).colorScheme.primary,
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
