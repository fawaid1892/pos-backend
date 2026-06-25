import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:flutter_bluetooth_basic/flutter_bluetooth_basic.dart';
import '../models/transaction.dart';
import '../providers/theme_provider.dart';
import '../services/thermal_print_service.dart';
import '../utils/responsive.dart';

class ReceiptScreen extends StatefulWidget {
  const ReceiptScreen({super.key});

  @override
  State<ReceiptScreen> createState() => _ReceiptScreenState();
}

class _ReceiptScreenState extends State<ReceiptScreen> {
  final _printService = ThermalPrintService();
  bool _isPrinting = false;

  Future<void> _showPrintDialog(Transaction transaction) async {
    final devices = await _printService.getBondedDevices();

    if (!mounted) return;

    if (devices.isEmpty) {
      _showNoPrinterDialog();
      return;
    }

    showModalBottomSheet(
      context: context,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      builder: (ctx) => _buildPrinterListSheet(ctx, transaction, devices),
    );
  }

  void _showNoPrinterDialog() {
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        icon: const Icon(Icons.print_disabled, size: 48, color: Colors.orange),
        title: const Text('Printer Tidak Ditemukan'),
        content: const Text(
          'Tidak ada printer Bluetooth yang terhubung.\n\n'
          'Pastikan printer thermal sudah di-pairing dengan perangkat ini '
          'melalui pengaturan Bluetooth.',
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx),
            child: const Text('OK'),
          ),
          TextButton.icon(
            onPressed: () {
              Navigator.pop(ctx);
              _tryWindowsPrint(transactionFromParent: ctx);
            },
            icon: const Icon(Icons.computer, size: 16),
            label: const Text('Coba USB (Windows)'),
          ),
        ],
      ),
    );
  }

  Transaction? _cachedTransaction;

  void _tryWindowsPrint({BuildContext? transactionFromParent}) {
    final transaction = _cachedTransaction;
    if (transaction == null) return;
    _printViaUsb(transaction);
  }

  Widget _buildPrinterListSheet(
      BuildContext ctx, Transaction transaction, List<BluetoothDevice> devices) {
    return Padding(
      padding: const EdgeInsets.all(16),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              const Icon(Icons.print, size: 20),
              const SizedBox(width: 8),
              const Text('Pilih Printer',
                  style: TextStyle(
                      fontSize: 16, fontWeight: FontWeight.bold)),
              const Spacer(),
              IconButton(
                icon: const Icon(Icons.refresh),
                onPressed: () {
                  Navigator.pop(ctx);
                  _showPrintDialog(transaction);
                },
                tooltip: 'Scan ulang',
              ),
            ],
          ),
          const SizedBox(height: 8),
          const Text('Printer Bluetooth terdeteksi:',
              style: TextStyle(color: Colors.grey, fontSize: 13)),
          const SizedBox(height: 8),
          ...devices.map((device) => ListTile(
                leading: const CircleAvatar(
                  child: Icon(Icons.bluetooth),
                ),
                title: Text(device.name ?? 'Printer'),
                subtitle: Text(device.address ?? ''),
                trailing: const Icon(Icons.chevron_right),
                onTap: () async {
                  Navigator.pop(ctx);
                  await _printViaBluetooth(device, transaction);
                },
              )),
          const Divider(),
          ListTile(
            leading: const CircleAvatar(
              child: Icon(Icons.usb),
            ),
            title: const Text('USB / Windows Printer'),
            subtitle: const Text('Cetak via port USB (LPT/USB)'),
            trailing: const Icon(Icons.chevron_right),
            onTap: () async {
              Navigator.pop(ctx);
              await _printViaUsb(transaction);
            },
          ),
          const SizedBox(height: 8),
        ],
      ),
    );
  }

  Future<void> _printViaBluetooth(
      BluetoothDevice device, Transaction transaction) async {
    setState(() => _isPrinting = true);

    final success = await _printService.printReceiptBluetooth(
      device: device,
      transaction: transaction,
      storeName: 'TOKO KAMI',
      storeAddress: 'POS Multi Branch',
      storePhone: '021-12345678',
    );

    if (!mounted) return;
    setState(() => _isPrinting = false);

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
            success ? 'Struk berhasil dicetak!' : 'Gagal mencetak struk'),
        backgroundColor: success ? Colors.green : Colors.red,
      ),
    );
  }

  Future<void> _printViaUsb(Transaction transaction) async {
    setState(() => _isPrinting = true);

    final success = await _printService.printReceiptUsb(
      transaction: transaction,
      storeName: 'TOKO KAMI',
      storeAddress: 'POS Multi Branch',
      storePhone: '021-12345678',
    );

    if (!mounted) return;
    setState(() => _isPrinting = false);

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Text(
            success
                ? 'Struk dikirim ke printer USB'
                : 'USB printing tidak tersedia di platform ini'),
        backgroundColor: success ? Colors.green : Colors.orange,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final transaction =
        ModalRoute.of(context)!.settings.arguments as Transaction;
    _cachedTransaction = transaction;
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
          Stack(
            children: [
              IconButton(
                icon: const Icon(Icons.print),
                onPressed:
                    _isPrinting ? null : () => _showPrintDialog(transaction),
                tooltip: 'Cetak Struk',
              ),
              if (_isPrinting)
                const Positioned(
                  right: 8,
                  top: 8,
                  child: SizedBox(
                    width: 14,
                    height: 14,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
                ),
            ],
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
                    if (transaction.paymentReference != null &&
                        transaction.paymentReference!.isNotEmpty)
                      _infoRow('Referensi', transaction.paymentReference!,
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
                    if (transaction.taxRate > 0)
                      _totalRow(
                          'Pajak (${(transaction.taxRate * 100).toStringAsFixed(0)}%)',
                          transaction.taxAmount,
                          colorScheme),
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

            // Action buttons
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
            const SizedBox(height: 12),

            // Print button (prominent)
            SizedBox(
              width: double.infinity,
              height: 48,
              child: ElevatedButton.icon(
                onPressed: _isPrinting
                    ? null
                    : () => _showPrintDialog(transaction),
                icon: _isPrinting
                    ? const SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(
                            strokeWidth: 2, color: Colors.white),
                      )
                    : const Icon(Icons.print),
                label: Text(
                  _isPrinting ? 'Mencetak...' : 'Cetak Struk',
                  style: const TextStyle(fontSize: 16),
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor: colorScheme.primary,
                  foregroundColor: colorScheme.onPrimary,
                ),
              ),
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
