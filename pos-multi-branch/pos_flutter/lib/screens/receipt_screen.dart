import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../models/transaction.dart';
import '../providers/theme_provider.dart';
import '../utils/responsive.dart';

class ReceiptScreen extends StatelessWidget {
  const ReceiptScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final transaction =
        ModalRoute.of(context)!.settings.arguments as Transaction;
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final isTablet = context.isTablet;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Struk Pembayaran'),
        automaticallyImplyLeading: false,
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
            icon: const Icon(Icons.print),
            onPressed: () {
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                    content: Text('Cetak struk via Bluetooth printer')),
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: EdgeInsets.all(isTablet ? 32 : 24),
        child: Column(
          children: [
            Icon(
              Icons.check_circle,
              size: 64,
              color: Colors.green[600],
            ),
            const SizedBox(height: 8),
            Text(
              'Pembayaran Berhasil!',
              style: TextStyle(
                fontSize: 20,
                fontWeight: FontWeight.bold,
                color: colorScheme.onSurface,
              ),
            ),
            const SizedBox(height: 24),

            // Receipt card
            Card(
              elevation: 2,
              child: Padding(
                padding: const EdgeInsets.all(16),
                child: Column(
                  children: [
                    // Header
                    Text(
                      'TOKO KAMI',
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: colorScheme.primary,
                      ),
                    ),
                    Text('POS Multi Branch',
                        style: TextStyle(color: colorScheme.onSurfaceVariant)),
                    const Divider(),

                    // Receipt info
                    _infoRow('No. Struk', transaction.receiptNumber ?? '-',
                        colorScheme),
                    _infoRow(
                        'Tanggal',
                        '${transaction.createdAt.day}/${transaction.createdAt.month}/${transaction.createdAt.year} '
                            '${transaction.createdAt.hour.toString().padLeft(2, '0')}:'
                            '${transaction.createdAt.minute.toString().padLeft(2, '0')}',
                        colorScheme),
                    _infoRow('Kasir', transaction.cashierName, colorScheme),
                    _infoRow('Metode Bayar', transaction.paymentMethod,
                        colorScheme),
                    const Divider(),

                    // Items
                    const Text('Pesanan',
                        style: TextStyle(fontWeight: FontWeight.bold)),
                    const SizedBox(height: 8),
                    ...transaction.items.map((item) => Padding(
                          padding: const EdgeInsets.symmetric(vertical: 2),
                          child: Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Expanded(
                                child: Text(item.productName),
                              ),
                              Text('${item.quantity}x'),
                              const SizedBox(width: 8),
                              Text(
                                'Rp ${_formatPrice(item.subtotal)}',
                                style: const TextStyle(
                                    fontWeight: FontWeight.w600),
                              ),
                            ],
                          ),
                        )),
                    const Divider(),

                    // Totals
                    _totalRow('Subtotal', transaction.total, colorScheme),
                    if (transaction.discountTotal > 0)
                      _totalRow('Diskon', -transaction.discountTotal,
                          colorScheme, color: Colors.red),
                    _totalRow('Grand Total', transaction.grandTotal, colorScheme,
                        bold: true, color: Colors.green),
                    const Divider(),
                    _totalRow('Bayar', transaction.amountPaid, colorScheme),
                    _totalRow('Kembali', transaction.change, colorScheme,
                        bold: true, color: Colors.blue),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 24),

            Row(
              children: [
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () {
                      Navigator.pushNamedAndRemoveUntil(
                          context, '/pos', (route) => false);
                    },
                    icon: const Icon(Icons.add_shopping_cart),
                    label: const Text('Transaksi Baru'),
                  ),
                ),
                const SizedBox(width: 12),
                Expanded(
                  child: OutlinedButton.icon(
                    onPressed: () {
                      Clipboard.setData(ClipboardData(
                          text:
                              'Struk: ${transaction.receiptNumber}\nTotal: Rp ${_formatPrice(transaction.grandTotal)}'));
                      ScaffoldMessenger.of(context).showSnackBar(
                        const SnackBar(content: Text('Struk disalin')),
                      );
                    },
                    icon: const Icon(Icons.copy),
                    label: const Text('Salin Struk'),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _infoRow(String label, String value, ColorScheme colorScheme) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: TextStyle(color: colorScheme.onSurfaceVariant)),
          Text(value,
              style: const TextStyle(fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }

  Widget _totalRow(String label, double amount, ColorScheme colorScheme,
      {bool bold = false, Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: TextStyle(
                  fontWeight: bold ? FontWeight.bold : FontWeight.w500)),
          Text(
            'Rp ${_formatPrice(amount)}',
            style: TextStyle(
                fontWeight: bold ? FontWeight.bold : FontWeight.w500,
                color: color ?? colorScheme.onSurface),
          ),
        ],
      ),
    );
  }

  String _formatPrice(double amount) {
    return amount.abs().toStringAsFixed(0).replaceAllMapped(
        RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.');
  }
}
