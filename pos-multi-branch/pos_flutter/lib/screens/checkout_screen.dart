import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/cart_provider.dart';
import '../providers/auth_provider.dart';
import '../providers/sync_provider.dart';
import '../providers/theme_provider.dart';
import '../services/transaction_service.dart';
import '../widgets/sync_status_widget.dart';
import '../utils/responsive.dart';

class CheckoutScreen extends StatefulWidget {
  const CheckoutScreen({super.key});

  @override
  State<CheckoutScreen> createState() => _CheckoutScreenState();
}

class _CheckoutScreenState extends State<CheckoutScreen> {
  final _amountController = TextEditingController();
  final _discountController = TextEditingController();
  final _transactionService = TransactionService();
  final List<Map<String, dynamic>> _paymentMethods = const [
    {'key': 'Tunai', 'icon': Icons.money},
    {'key': 'QRIS', 'icon': Icons.qr_code},
    {'key': 'Transfer', 'icon': Icons.account_balance},
    {'key': 'Kartu Debit', 'icon': Icons.credit_card},
  ];
  bool _isProcessing = false;

  @override
  void dispose() {
    _amountController.dispose();
    _discountController.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    final cart = context.read<CartProvider>();
    _amountController.text = cart.grandTotal.toStringAsFixed(0);
  }

  double get _amountPaid =>
      double.tryParse(_amountController.text.replaceAll('.', '')) ?? 0;

  double get _change => _amountPaid - context.read<CartProvider>().grandTotal;

  bool get _isPaymentValid =>
      _amountPaid >= context.read<CartProvider>().grandTotal;

  Future<void> _processPayment() async {
    if (!_isPaymentValid) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('Jumlah bayar kurang dari total belanja'),
          backgroundColor: Colors.red,
        ),
      );
      return;
    }

    setState(() => _isProcessing = true);

    final cart = context.read<CartProvider>();
    final auth = context.read<AuthProvider>();

    // Apply optional discount
    final disc =
        double.tryParse(_discountController.text.replaceAll('.', '')) ?? 0;
    if (disc > 0) {
      cart.setCartDiscount(disc);
    }

    final transaction = cart.buildTransaction(
      id: 'trx_${DateTime.now().millisecondsSinceEpoch}',
      branchId: auth.branchId ?? 'branch_001',
      cashierId: auth.userId ?? 'unknown',
      cashierName: auth.userName ?? 'Kasir',
      amountPaid: _amountPaid,
      change: _change,
      receiptNumber: 'REC-${DateTime.now().millisecondsSinceEpoch}',
    );

    try {
      // Save transaction to local SQLite
      await _transactionService.submitTransaction(transaction);
      if (!mounted) return;
      cart.clearCart();
      Navigator.pushReplacementNamed(context, '/receipt', arguments: transaction);
    } catch (e) {
      setState(() => _isProcessing = false);
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Gagal memproses: $e'),
          backgroundColor: Colors.red,
        ),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final cart = context.watch<CartProvider>();
    final theme = Theme.of(context);
    final isTablet = context.isTablet;
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Pembayaran'),
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
        ],
      ),
      body: SingleChildScrollView(
        padding: EdgeInsets.all(isTablet ? 24 : 16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Order items summary
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Pesanan',
                        style: TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 16)),
                    const SizedBox(height: 8),
                    ...cart.items.map((item) => Padding(
                          padding: const EdgeInsets.symmetric(vertical: 2),
                          child: Row(
                            mainAxisAlignment: MainAxisAlignment.spaceBetween,
                            children: [
                              Text('${item.productName} x${item.quantity}'),
                              Text('Rp ${_formatPrice(item.subtotal)}'),
                            ],
                          ),
                        )),
                    const Divider(),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text('Subtotal',
                            style: TextStyle(fontWeight: FontWeight.w500)),
                        Text('Rp ${_formatPrice(cart.total)}'),
                      ],
                    ),
                    if (cart.discountTotal > 0)
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text('Diskon',
                              style: TextStyle(fontWeight: FontWeight.w500)),
                          Text('- Rp ${_formatPrice(cart.discountTotal)}',
                              style: const TextStyle(color: Colors.red)),
                        ],
                      ),
                    const Divider(),
                    Row(
                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                      children: [
                        const Text('Grand Total',
                            style: TextStyle(
                                fontWeight: FontWeight.bold, fontSize: 18)),
                        Text('Rp ${_formatPrice(cart.grandTotal)}',
                            style: TextStyle(
                                fontWeight: FontWeight.bold,
                                fontSize: 18,
                                color: colorScheme.primary)),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Discount
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Diskon (Opsional)',
                        style: TextStyle(fontWeight: FontWeight.w600)),
                    const SizedBox(height: 8),
                    TextField(
                      controller: _discountController,
                      decoration: InputDecoration(
                        labelText: 'Diskon (Rp)',
                        prefixText: 'Rp ',
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        hintText: '0',
                        filled: true,
                        fillColor: colorScheme.surfaceVariant,
                      ),
                      keyboardType: TextInputType.number,
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Payment method
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Metode Bayar',
                        style: TextStyle(fontWeight: FontWeight.w600)),
                    const SizedBox(height: 8),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: _paymentMethods.map((method) {
                        final isSelected =
                            cart.selectedPaymentMethod == method['key'];
                        return ChoiceChip(
                          label: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(method['icon'] as IconData, size: 18),
                              const SizedBox(width: 4),
                              Text(method['key'] as String),
                            ],
                          ),
                          selected: isSelected,
                          onSelected: (_) =>
                              cart.setPaymentMethod(method['key'] as String),
                        );
                      }).toList(),
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Amount paid
            Card(
              child: Padding(
                padding: const EdgeInsets.all(12),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Jumlah Bayar',
                        style: TextStyle(fontWeight: FontWeight.w600)),
                    const SizedBox(height: 8),
                    TextField(
                      controller: _amountController,
                      decoration: InputDecoration(
                        labelText: 'Nominal bayar',
                        prefixText: 'Rp ',
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        filled: true,
                        fillColor: colorScheme.surfaceVariant,
                      ),
                      keyboardType: TextInputType.number,
                      onChanged: (_) => setState(() {}),
                    ),
                    if (_amountPaid > 0) ...[
                      const SizedBox(height: 8),
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          const Text('Kembalian:',
                              style: TextStyle(fontSize: 16)),
                          Text(
                            'Rp ${_formatPrice(_change)}',
                            style: TextStyle(
                              fontSize: 18,
                              fontWeight: FontWeight.bold,
                              color: _isPaymentValid
                                  ? colorScheme.primary
                                  : colorScheme.error,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),
            SizedBox(
              width: double.infinity,
              height: 52,
              child: ElevatedButton.icon(
                onPressed: _isProcessing ? null : _processPayment,
                icon: _isProcessing
                    ? const SizedBox(
                        height: 20,
                        width: 20,
                        child: CircularProgressIndicator(strokeWidth: 2))
                    : const Icon(Icons.check_circle),
                label: Text(
                  _isProcessing ? 'Memproses...' : 'Konfirmasi Pembayaran',
                  style: const TextStyle(fontSize: 16),
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.green,
                  foregroundColor: Colors.white,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  String _formatPrice(double amount) {
    return amount
        .toStringAsFixed(0)
        .replaceAllMapped(
            RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.');
  }
}
