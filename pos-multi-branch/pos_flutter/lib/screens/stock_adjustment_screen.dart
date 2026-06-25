import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/stock_provider.dart';
import '../providers/theme_provider.dart';
import '../models/stock_adjustment.dart';
import '../widgets/error_state_widget.dart';
import '../widgets/empty_state_widget.dart';
import '../widgets/shimmer_loading.dart';
import '../utils/responsive.dart';

class StockAdjustmentScreen extends StatefulWidget {
  final ProductStock? initialProduct;

  const StockAdjustmentScreen({super.key, this.initialProduct});

  @override
  State<StockAdjustmentScreen> createState() => _StockAdjustmentScreenState();
}

class _StockAdjustmentScreenState extends State<StockAdjustmentScreen> {
  final _formKey = GlobalKey<FormState>();
  final _quantityController = TextEditingController();
  final _reasonController = TextEditingController();

  String _adjustmentType = 'in'; // 'in' or 'out'
  String? _selectedProductId;
  List<ProductStock> _products = [];
  bool _isLoading = true;
  String? _loadError;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadProducts());
  }

  @override
  void dispose() {
    _quantityController.dispose();
    _reasonController.dispose();
    super.dispose();
  }

  Future<void> _loadProducts() async {
    setState(() {
      _isLoading = true;
      _loadError = null;
    });
    final stockProv = context.read<StockProvider>();
    try {
      await stockProv.loadInventory();
      if (!mounted) return;
      setState(() {
        _products = stockProv.inventory;
        _isLoading = false;
        // Pre-select product if provided
        if (widget.initialProduct != null) {
          _selectedProductId = widget.initialProduct!.productId;
        }
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _isLoading = false;
        _loadError = e.toString();
      });
    }
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_selectedProductId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Pilih produk terlebih dahulu')),
      );
      return;
    }

    final stockProv = context.read<StockProvider>();
    final quantity = int.parse(_quantityController.text.trim());

    final result = await stockProv.submitAdjustment(
      productId: _selectedProductId!,
      quantity: quantity,
      reason: _reasonController.text.trim().isEmpty
          ? (_adjustmentType == 'in' ? 'Stok Masuk' : 'Stok Keluar')
          : _reasonController.text.trim(),
      type: _adjustmentType,
    );

    if (!mounted) return;

    if (result != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
              'Adjustment berhasil: ${result.productName} (${result.type == "in" ? "+" : ""}${result.quantity})'),
          backgroundColor: Colors.green,
        ),
      );
      Navigator.pop(context);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Gagal melakukan adjustment'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final stockProv = context.watch<StockProvider>();
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Adjustment Stok'),
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
      body: _isLoading
          ? const ShimmerPage(itemCount: 4)
          : _loadError != null
              ? ErrorStateWidget(
                  message: _loadError!,
                  title: 'Gagal memuat produk',
                  onRetry: _loadProducts,
                )
              : _products.isEmpty
                  ? const EmptyStateWidget.stock()
                  : _buildForm(stockProv, colorScheme),
    );
  }

  Widget _buildForm(StockProvider stockProv, ColorScheme colorScheme) {
    final isTablet = context.isTablet;

    return SingleChildScrollView(
      padding: EdgeInsets.all(isTablet ? 24 : 16),
      child: Form(
        key: _formKey,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            // Type selector
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Tipe Adjustment',
                        style: TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 12),
                    Row(
                      children: [
                        Expanded(
                          child: _buildTypeButton(
                            label: 'Stok Masuk',
                            icon: Icons.add_circle,
                            value: 'in',
                            color: Colors.green,
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: _buildTypeButton(
                            label: 'Stok Keluar',
                            icon: Icons.remove_circle,
                            value: 'out',
                            color: Colors.red,
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Product selector
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Produk',
                        style: TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 8),
                    DropdownButtonFormField<String>(
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        hintText: 'Pilih produk',
                        filled: true,
                        fillColor: colorScheme.surfaceVariant,
                      ),
                      value: _selectedProductId,
                      items: _products.map((p) {
                        return DropdownMenuItem(
                          value: p.productId,
                          child: Text(
                              '${p.productName} (stok: ${p.currentStock})'),
                        );
                      }).toList(),
                      onChanged: (val) {
                        setState(() => _selectedProductId = val);
                      },
                      validator: (val) =>
                          val == null ? 'Pilih produk' : null,
                    ),
                    if (_selectedProductId != null) ...[
                      const SizedBox(height: 8),
                      Text(
                        'Stok saat ini: ${_products.firstWhere((p) => p.productId == _selectedProductId!).currentStock}',
                        style: TextStyle(
                          color: colorScheme.primary,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                    ],
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Quantity input
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Jumlah',
                        style: TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _quantityController,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        hintText: 'Masukkan jumlah',
                        prefixIcon: const Icon(Icons.numbers),
                        filled: true,
                        fillColor: colorScheme.surfaceVariant,
                      ),
                      validator: (val) {
                        if (val == null || val.trim().isEmpty) {
                          return 'Masukkan jumlah';
                        }
                        final qty = int.tryParse(val.trim());
                        if (qty == null || qty <= 0) {
                          return 'Jumlah harus lebih dari 0';
                        }
                        if (_adjustmentType == 'out' &&
                            _selectedProductId != null) {
                          final product = _products.firstWhere(
                            (p) => p.productId == _selectedProductId,
                          );
                          if (qty > product.currentStock) {
                            return 'Stok tidak mencukupi (sisa: ${product.currentStock})';
                          }
                        }
                        return null;
                      },
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Reason input
            Card(
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Alasan',
                        style: TextStyle(
                            fontSize: 16, fontWeight: FontWeight.w600)),
                    const SizedBox(height: 8),
                    TextFormField(
                      controller: _reasonController,
                      maxLines: 3,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        hintText: 'Contoh: Rusak, Kadaluarsa, Restock, dll',
                        filled: true,
                        fillColor: colorScheme.surfaceVariant,
                      ),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            // Submit button
            SizedBox(
              height: 48,
              child: ElevatedButton.icon(
                onPressed:
                    stockProv.isSubmittingAdjustment ? null : _submit,
                icon: stockProv.isSubmittingAdjustment
                    ? const SizedBox(
                        width: 18,
                        height: 18,
                        child:
                            CircularProgressIndicator(strokeWidth: 2),
                      )
                    : const Icon(Icons.save),
                label: Text(
                  stockProv.isSubmittingAdjustment
                      ? 'Menyimpan...'
                      : 'Simpan Adjustment',
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor:
                      _adjustmentType == 'in' ? Colors.green : Colors.red,
                  foregroundColor: Colors.white,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTypeButton({
    required String label,
    required IconData icon,
    required String value,
    required Color color,
  }) {
    final selected = _adjustmentType == value;
    final brightness = Theme.of(context).brightness;

    return GestureDetector(
      onTap: () => setState(() => _adjustmentType = value),
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 16),
        decoration: BoxDecoration(
          color: selected
              ? color.withOpacity(0.1)
              : (brightness == Brightness.dark
                  ? Colors.grey.shade800
                  : Colors.grey.shade100),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: selected ? color : Colors.grey.shade300,
            width: selected ? 2 : 1,
          ),
        ),
        child: Column(
          children: [
            Icon(icon, color: selected ? color : Colors.grey, size: 32),
            const SizedBox(height: 4),
            Text(
              label,
              style: TextStyle(
                fontWeight: selected ? FontWeight.bold : FontWeight.normal,
                color: selected ? color : Colors.grey,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
