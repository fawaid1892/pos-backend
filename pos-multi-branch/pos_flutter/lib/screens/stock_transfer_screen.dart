import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../models/stock_adjustment.dart';
import '../models/branch.dart';
import '../providers/stock_provider.dart';
import '../services/mock_stock_service.dart';

class StockTransferScreen extends StatefulWidget {
  final ProductStock? initialProduct;

  const StockTransferScreen({super.key, this.initialProduct});

  @override
  State<StockTransferScreen> createState() => _StockTransferScreenState();
}

class _StockTransferScreenState extends State<StockTransferScreen> {
  final _formKey = GlobalKey<FormState>();
  final _quantityController = TextEditingController();

  final MockStockService _stockService = MockStockService();

  String? _sourceBranchId;
  String? _targetBranchId;
  String? _selectedProductId;

  List<Branch> _branches = [];
  List<ProductStock> _sourceProducts = [];
  bool _isLoadingBranches = true;
  bool _isLoadingProducts = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) => _loadBranches());
  }

  @override
  void dispose() {
    _quantityController.dispose();
    super.dispose();
  }

  Future<void> _loadBranches() async {
    setState(() {
      _branches = _stockService.getBranches();
      _isLoadingBranches = false;
    });
  }

  Future<void> _loadSourceProducts() async {
    if (_sourceBranchId == null) return;
    setState(() => _isLoadingProducts = true);

    final items = await _stockService.getInventory(_sourceBranchId!);
    if (!mounted) return;
    setState(() {
      _sourceProducts = items;
      _selectedProductId = null;
      _isLoadingProducts = false;
      // Pre-select if applicable
      if (widget.initialProduct != null) {
        final match = _sourceProducts
            .where((p) => p.productId == widget.initialProduct!.productId)
            .toList();
        if (match.isNotEmpty) {
          _selectedProductId = match.first.productId;
        }
      }
    });
  }

  List<Branch> get _availableTargetBranches {
    return _branches.where((b) => b.id != _sourceBranchId).toList();
  }

  Future<void> _submit() async {
    if (!_formKey.currentState!.validate()) return;
    if (_sourceBranchId == null ||
        _targetBranchId == null ||
        _selectedProductId == null) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Lengkapi semua field')),
      );
      return;
    }

    final stockProv = context.read<StockProvider>();
    final quantity = int.parse(_quantityController.text.trim());

    final result = await stockProv.submitTransfer(
      sourceBranchId: _sourceBranchId!,
      targetBranchId: _targetBranchId!,
      productId: _selectedProductId!,
      quantity: quantity,
    );

    if (!mounted) return;

    if (result != null) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(
            'Transfer berhasil: ${result.productName} x${result.quantity} '
            '(${result.sourceBranchName} → ${result.targetBranchName})',
          ),
          backgroundColor: Colors.green,
        ),
      );
      Navigator.pop(context);
    } else {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Gagal melakukan transfer'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final stockProv = context.watch<StockProvider>();

    return Scaffold(
      appBar: AppBar(
        title: const Text('Transfer Antar Cabang'),
      ),
      body: _isLoadingBranches
          ? const Center(child: CircularProgressIndicator())
          : SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: Form(
                key: _formKey,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Source branch selector
                    _buildSection('Cabang Asal', [
                      DropdownButtonFormField<String>(
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                          prefixIcon: Icon(Icons.store),
                          hintText: 'Pilih cabang asal',
                        ),
                        value: _sourceBranchId,
                        items: _branches.map((b) {
                          return DropdownMenuItem(
                            value: b.id,
                            child: Text(b.name),
                          );
                        }).toList(),
                        onChanged: (val) {
                          setState(() {
                            _sourceBranchId = val;
                            _targetBranchId = null;
                          });
                          _loadSourceProducts();
                        },
                        validator: (val) =>
                            val == null ? 'Pilih cabang asal' : null,
                      ),
                    ]),

                    const SizedBox(height: 16),

                    // Target branch selector
                    _buildSection('Cabang Tujuan', [
                      DropdownButtonFormField<String>(
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                          prefixIcon: Icon(Icons.store_mall_directory),
                          hintText: 'Pilih cabang tujuan',
                        ),
                        value: _targetBranchId,
                        items: _availableTargetBranches.map((b) {
                          return DropdownMenuItem(
                            value: b.id,
                            child: Text(b.name),
                          );
                        }).toList(),
                        onChanged: (val) {
                          setState(() => _targetBranchId = val);
                        },
                        validator: (val) =>
                            val == null ? 'Pilih cabang tujuan' : null,
                      ),
                    ]),

                    const SizedBox(height: 16),

                    // Arrow visual
                    if (_sourceBranchId != null && _targetBranchId != null)
                      Padding(
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Text(
                              _branches
                                      .firstWhere(
                                          (b) => b.id == _sourceBranchId)
                                      .name,
                              style: const TextStyle(
                                  fontWeight: FontWeight.w500),
                            ),
                            const Padding(
                              padding: EdgeInsets.symmetric(horizontal: 12),
                              child: Icon(Icons.arrow_forward,
                                  color: Colors.blue, size: 32),
                            ),
                            Text(
                              _branches
                                      .firstWhere(
                                          (b) => b.id == _targetBranchId)
                                      .name,
                              style: const TextStyle(
                                  fontWeight: FontWeight.w500),
                            ),
                          ],
                        ),
                      ),

                    const SizedBox(height: 16),

                    // Product selector
                    _buildSection('Produk', [
                      if (_sourceBranchId == null)
                        const Padding(
                          padding: EdgeInsets.all(12),
                          child: Text(
                            'Pilih cabang asal terlebih dahulu',
                            style: TextStyle(color: Colors.grey),
                          ),
                        )
                      else if (_isLoadingProducts)
                        const Center(child: CircularProgressIndicator())
                      else
                        DropdownButtonFormField<String>(
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                            prefixIcon: Icon(Icons.inventory_2),
                            hintText: 'Pilih produk',
                          ),
                          value: _selectedProductId,
                          items: _sourceProducts.map((p) {
                            return DropdownMenuItem(
                              value: p.productId,
                              child: Text(
                                '${p.productName} (stok: ${p.currentStock})',
                              ),
                            );
                          }).toList(),
                          onChanged: (val) {
                            setState(() => _selectedProductId = val);
                          },
                          validator: (val) =>
                              val == null ? 'Pilih produk' : null,
                        ),
                    ]),

                    const SizedBox(height: 16),

                    // Quantity input
                    _buildSection('Jumlah', [
                      TextFormField(
                        controller: _quantityController,
                        keyboardType: TextInputType.number,
                        decoration: const InputDecoration(
                          border: OutlineInputBorder(),
                          hintText: 'Masukkan jumlah transfer',
                          prefixIcon: Icon(Icons.numbers),
                        ),
                        validator: (val) {
                          if (val == null || val.trim().isEmpty) {
                            return 'Masukkan jumlah';
                          }
                          final qty = int.tryParse(val.trim());
                          if (qty == null || qty <= 0) {
                            return 'Jumlah harus lebih dari 0';
                          }
                          if (_selectedProductId != null) {
                            final product = _sourceProducts.firstWhere(
                              (p) => p.productId == _selectedProductId,
                            );
                            if (qty > product.currentStock) {
                              return 'Stok tidak mencukupi (sisa: ${product.currentStock})';
                            }
                          }
                          return null;
                        },
                      ),
                    ]),

                    const SizedBox(height: 24),

                    // Current stock info
                    if (_selectedProductId != null && _sourceProducts.isNotEmpty)
                      Card(
                        color: Colors.blue.shade50,
                        child: Padding(
                          padding: const EdgeInsets.all(12),
                          child: Row(
                            children: [
                              const Icon(Icons.info_outline,
                                  color: Colors.blue),
                              const SizedBox(width: 8),
                              Text(
                                'Stok saat ini: ${_sourceProducts.firstWhere((p) => p.productId == _selectedProductId!).currentStock}',
                                style: const TextStyle(
                                    fontWeight: FontWeight.w500),
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
                            stockProv.isSubmittingTransfer ? null : _submit,
                        icon: stockProv.isSubmittingTransfer
                            ? const SizedBox(
                                width: 18,
                                height: 18,
                                child:
                                    CircularProgressIndicator(strokeWidth: 2),
                              )
                            : const Icon(Icons.swap_horiz),
                        label: Text(
                          stockProv.isSubmittingTransfer
                              ? 'Memproses Transfer...'
                              : 'Proses Transfer',
                        ),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.blue,
                          foregroundColor: Colors.white,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
    );
  }

  Widget _buildSection(String title, List<Widget> children) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title,
                style: const TextStyle(
                    fontSize: 16, fontWeight: FontWeight.w600)),
            const SizedBox(height: 12),
            ...children,
          ],
        ),
      ),
    );
  }
}
