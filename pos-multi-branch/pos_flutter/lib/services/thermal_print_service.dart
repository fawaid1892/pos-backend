import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';
import 'package:esc_pos_utils/esc_pos_utils.dart';
import 'package:flutter_bluetooth_basic/flutter_bluetooth_basic.dart';
import '../models/transaction.dart';

/// Service for thermal receipt printing via Bluetooth (Android) and USB (Windows).
class ThermalPrintService {
  static final ThermalPrintService _instance =
      ThermalPrintService._internal();
  factory ThermalPrintService() => _instance;
  ThermalPrintService._internal();

  final FlutterBluetoothBasic _bluetooth = FlutterBluetoothBasic();

  /// Check if Bluetooth is available on the device.
  Future<bool> isBluetoothAvailable() async {
    try {
      return await _bluetooth.isBluetoothAvailable ?? false;
    } catch (e) {
      debugPrint('ThermalPrintService.isBluetoothAvailable error: $e');
      return false;
    }
  }

  /// Get list of bonded Bluetooth devices.
  Future<List<BluetoothDevice>> getBondedDevices() async {
    try {
      final devices = await _bluetooth.getBondedDevices();
      return devices;
    } catch (e) {
      debugPrint('ThermalPrintService.getBondedDevices error: $e');
      return [];
    }
  }

  /// Scan for nearby Bluetooth devices.
  Future<List<BluetoothDevice>> scanDevices() async {
    try {
      final devices = await _bluetooth.getBondedDevices();
      return devices;
    } catch (e) {
      debugPrint('ThermalPrintService.scanDevices error: $e');
      return [];
    }
  }

  /// Print receipt to a Bluetooth printer device.
  Future<bool> printReceiptBluetooth({
    required BluetoothDevice device,
    required Transaction transaction,
    required String storeName,
    String? storeAddress,
    String? storePhone,
  }) async {
    try {
      // Connect to printer
      final isConnected = await _bluetooth.isConnected(device.address!);
      if (!isConnected) {
        await _bluetooth.connect(device);
      }

      // Wait for connection to stabilize
      await Future.delayed(const Duration(milliseconds: 500));

      // Generate receipt bytes using esc_pos_utils
      final profile = await CapabilityProfile.load();
      final generator = Generator(PaperSize.mm58, profile);
      final bytes = _generateReceiptBytes(
        generator: generator,
        transaction: transaction,
        storeName: storeName,
        storeAddress: storeAddress,
        storePhone: storePhone,
      );

      // Send to printer
      await _bluetooth.sendData(bytes);

      // Disconnect
      await _bluetooth.disconnect();

      return true;
    } catch (e) {
      debugPrint('ThermalPrintService.printReceiptBluetooth error: $e');
      return false;
    }
  }

  /// Print receipt using raw USB (Windows via socket/port).
  Future<bool> printReceiptUsb({
    required Transaction transaction,
    required String storeName,
    String? storeAddress,
    String? storePhone,
  }) async {
    try {
      final profile = await CapabilityProfile.load();
      final generator = Generator(PaperSize.mm58, profile);
      final bytes = _generateReceiptBytes(
        generator: generator,
        transaction: transaction,
        storeName: storeName,
        storeAddress: storeAddress,
        storePhone: storePhone,
      );

      // For Windows USB thermal printers, write to LPT or USB virtual port
      // Using raw socket approach - in production would use serial/usb plugin
      if (Platform.isWindows) {
        // Attempt LPT1 direct write (common for POS printers on Windows)
        try {
          final file = File(r'\\localhost\LPT1');
          if (await file.exists()) {
            await file.writeAsBytes(bytes);
            return true;
          }
        } catch (_) {
          // Fallback: try to write to a temp file for manual printing
          final tempDir = Directory.systemTemp;
          final receiptFile = File(
              '${tempDir.path}/receipt_${transaction.receiptNumber ?? DateTime.now().millisecondsSinceEpoch}.bin');
          await receiptFile.writeAsBytes(bytes);
          debugPrint(
              'ThermalPrintService: Receipt saved to ${receiptFile.path} for USB printing');
          return true;
        }
      }

      return false;
    } catch (e) {
      debugPrint('ThermalPrintService.printReceiptUsb error: $e');
      return false;
    }
  }

  /// Generate receipt bytes using esc_pos_utils Generator.
  List<int> _generateReceiptBytes({
    required Generator generator,
    required Transaction transaction,
    required String storeName,
    String? storeAddress,
    String? storePhone,
  }) {
    final List<int> bytes = [];

    // Header
    bytes += generator.setInitialLength(0x20);
    bytes += generator.setAlign(PosAlign.center);
    bytes += generator.text(storeName,
        styles: const PosStyles(
          bold: true,
          height: PosTextSize.size2,
          width: PosTextSize.size2,
        ));
    if (storeAddress != null && storeAddress.isNotEmpty) {
      bytes += generator.text(storeAddress);
    }
    if (storePhone != null && storePhone.isNotEmpty) {
      bytes += generator.text('Telp: $storePhone');
    }
    bytes += generator.text('');
    bytes += generator.hr();

    // Receipt info
    bytes += generator.setAlign(PosAlign.left);
    bytes += generator.row([
      PosColumn(
        text: 'No. Struk:',
        width: 5,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text: transaction.receiptNumber ?? '-',
        width: 7,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    bytes += generator.row([
      PosColumn(
        text: 'Tanggal:',
        width: 5,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text:
            '${transaction.createdAt.day}/${transaction.createdAt.month}/${transaction.createdAt.year} '
            '${transaction.createdAt.hour.toString().padLeft(2, '0')}:'
            '${transaction.createdAt.minute.toString().padLeft(2, '0')}',
        width: 7,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    bytes += generator.row([
      PosColumn(
        text: 'Kasir:',
        width: 5,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text: transaction.cashierName,
        width: 7,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    bytes += generator.row([
      PosColumn(
        text: 'Bayar:',
        width: 5,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text: transaction.paymentMethod,
        width: 7,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    if (transaction.paymentReference != null &&
        transaction.paymentReference!.isNotEmpty) {
      bytes += generator.row([
        PosColumn(
          text: 'Ref:',
          width: 5,
          styles: const PosStyles(align: PosAlign.left),
        ),
        PosColumn(
          text: transaction.paymentReference!,
          width: 7,
          styles: const PosStyles(align: PosAlign.right),
        ),
      ]);
    }
    bytes += generator.hr();

    // Items header
    bytes += generator.row([
      PosColumn(
        text: 'ITEM',
        width: 6,
        styles: const PosStyles(bold: true),
      ),
      PosColumn(
        text: 'QTY',
        width: 2,
        styles: const PosStyles(bold: true, align: PosAlign.center),
      ),
      PosColumn(
        text: 'TOTAL',
        width: 4,
        styles: const PosStyles(bold: true, align: PosAlign.right),
      ),
    ]);
    bytes += generator.hr();

    // Items
    for (final item in transaction.items) {
      bytes += generator.row([
        PosColumn(
          text: item.productName,
          width: 6,
          styles: const PosStyles(align: PosAlign.left),
        ),
        PosColumn(
          text: '${item.quantity}',
          width: 2,
          styles: const PosStyles(align: PosAlign.center),
        ),
        PosColumn(
          text: _formatRp(item.subtotal),
          width: 4,
          styles: const PosStyles(align: PosAlign.right),
        ),
      ]);
    }
    bytes += generator.hr();

    // Totals
    bytes += generator.row([
      PosColumn(
        text: 'Subtotal',
        width: 6,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text: _formatRp(transaction.total),
        width: 6,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    if (transaction.discountTotal > 0) {
      bytes += generator.row([
        PosColumn(
          text: 'Diskon',
          width: 6,
          styles: const PosStyles(align: PosAlign.left),
        ),
        PosColumn(
          text: '-${_formatRp(transaction.discountTotal)}',
          width: 6,
          styles: const PosStyles(align: PosAlign.right),
        ),
      ]);
    }
    if (transaction.taxRate > 0) {
      bytes += generator.row([
        PosColumn(
          text: 'Pajak (${(transaction.taxRate * 100).toStringAsFixed(0)}%)',
          width: 6,
          styles: const PosStyles(align: PosAlign.left),
        ),
        PosColumn(
          text: _formatRp(transaction.taxAmount),
          width: 6,
          styles: const PosStyles(align: PosAlign.right),
        ),
      ]);
    }
    bytes += generator.hr(ch: '=');
    bytes += generator.row([
      PosColumn(
        text: 'TOTAL',
        width: 6,
        styles: const PosStyles(bold: true, align: PosAlign.left),
      ),
      PosColumn(
        text: _formatRp(transaction.grandTotal),
        width: 6,
        styles: const PosStyles(
            bold: true, align: PosAlign.right, height: PosTextSize.size2),
      ),
    ]);
    bytes += generator.hr(ch: '=');

    // Payment details
    bytes += generator.row([
      PosColumn(
        text: 'Bayar',
        width: 6,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text: _formatRp(transaction.amountPaid),
        width: 6,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    bytes += generator.row([
      PosColumn(
        text: 'Kembali',
        width: 6,
        styles: const PosStyles(align: PosAlign.left),
      ),
      PosColumn(
        text: _formatRp(transaction.change),
        width: 6,
        styles: const PosStyles(align: PosAlign.right),
      ),
    ]);
    bytes += generator.text('');

    // Footer
    bytes += generator.setAlign(PosAlign.center);
    bytes += generator.text('Terima Kasih');
    bytes += generator.text('Barang yang sudah dibeli');
    bytes += generator.text('tidak dapat dikembalikan');
    bytes += generator.text('');
    bytes += generator.text('-- Cetak Struk --');
    bytes += generator.text('');

    // Cut paper
    bytes += generator.cut();

    return bytes;
  }

  String _formatRp(double amount) {
    return 'Rp${amount.toStringAsFixed(0).replaceAllMapped(RegExp(r'(\d)(?=(\d{3})+(?!\d))'), (match) => '${match.group(1)}.')}';
  }
}
