import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/transaction.dart';

class ReceiptScreen extends StatelessWidget {
  const ReceiptScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final transaction =
        ModalRoute.of(context)!.settings.arguments as Transaction;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Struk Pembayaran'),
        automaticallyImplyLeading: false,
        actions: [
          IconButton(
            icon: const Icon(Icons.print),
            onPressed: () {
              // esc_pos_bluetooth integration placeholder
              // await BluetoothManager.instance.startScan();
              ScaffoldMessenger.of(context).showSnackBar(
                const SnackBar(
                    content: Text('Cetak struk via Bluetooth printer')),
              );
            },
          ),
        ],
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(24),
        child: Column(
          children: [
            Icon(
              Icons.check_circle,
              size: 64,
              color: Colors.green[600],
            ),
            const SizedBox(height: 8),
            const Text(
              'Pembayaran Berhasil!',
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
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
                        color: Theme.of(context).colorScheme.primary,
                      ),
                    ),
                    const Text('POS Multi Branch',
                        style: TextStyle(color: Colors.grey)),
                    const Divider(),

                    // Receipt info
                    _infoRow('No. Struk', transaction.receiptNumber ?? '-'),
                    _infoRow(
                        'Tanggal',
                        '${transaction.createdAt.day}/${transaction.createdAt.month}/${transaction.createdAt.year} '
                            '${transaction.createdAt.hour.toString().padLeft(2, '0')}:'
                            '${transaction.createdAt.minute.toString().padLeft(2, '0')}'),
                    _infoRow('Kasir', transaction.cashierName),
                    _infoRow('Metode Bayar', transaction.paymentMethod),
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
                    _totalRow('Subtotal', transaction.total),
                    if (transaction.discountTotal > 0)
                      _totalRow('Diskon', -transaction.discountTotal,
                          color: Colors.red),
                    _totalRow('Grand Total', transaction.grandTotal,
                        bold: true, color: Colors.green),
                    const Divider(),
                    _totalRow('Bayar', transaction.amountPaid),
                    _totalRow('Kembali', transaction.change,
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
                      // Share receipt stub
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

  Widget _infoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: const TextStyle(color: Colors.grey)),
          Text(value, style: const TextStyle(fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }

  Widget _totalRow(String label, double amount,
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
                color: color),
          ),
        ],
      ),
    );
  }

  String _formatPrice(double amount) {
    return amount.abs()
        .toStringAsFixed(0)
        .replaceAllMapped(
            RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.');
  }
}
