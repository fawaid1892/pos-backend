import '../database/local_database.dart';
import '../models/transaction.dart';

/// Service for transaction CRUD operations backed by SQLite.
///
/// Replaces MockApiService for transaction operations.
class TransactionService {
  static final TransactionService _instance = TransactionService._internal();
  factory TransactionService() => _instance;
  TransactionService._internal();

  final LocalDatabase _db = LocalDatabase();

  /// Submit a new transaction (saves to local DB with pending sync).
  Future<Transaction> submitTransaction(Transaction transaction) async {
    final db = await _db.database;

    // 1. Insert transaction record
    await db.insert('transactions', {
      'id': transaction.id,
      'branch_id': transaction.branchId,
      'cashier_id': transaction.cashierId,
      'cashier_name': transaction.cashierName,
      'total': transaction.total,
      'discount_total': transaction.discountTotal,
      'tax_rate': transaction.taxRate,
      'tax_amount': transaction.taxAmount,
      'grand_total': transaction.grandTotal,
      'payment_method': transaction.paymentMethod,
      'amount_paid': transaction.amountPaid,
      'change_amount': transaction.change,
      'receipt_number': transaction.receiptNumber,
      'created_at': transaction.createdAt.toIso8601String(),
      'pending_sync': 1,
      'sync_status': 'pending',
    });

    // 2. Insert transaction items
    for (final item in transaction.items) {
      final itemId = '${transaction.id}_${item.productId}';
      await db.insert('transaction_items', {
        'id': itemId,
        'transaction_id': transaction.id,
        'product_id': item.productId,
        'product_name': item.productName,
        'barcode': item.barcode,
        'price': item.price,
        'quantity': item.quantity,
        'discount': item.discount,
        'subtotal': item.subtotal,
        'pending_sync': 1,
        'sync_status': 'pending',
      });

      // 3. Update stock (reduce from branch_products)
      await db.rawUpdate('''
        UPDATE branch_products
        SET stock = stock - ?, pending_sync = 1, sync_status = 'pending'
        WHERE branch_id = ? AND product_id = ?
      ''', [item.quantity, transaction.branchId, item.productId]);
    }

    return transaction;
  }

  /// Get transactions for a branch.
  Future<List<Transaction>> getTransactions(String branchId,
      {int limit = 50, int offset = 0}) async {
    final db = await _db.database;

    final transactionMaps = await db.query('transactions',
      where: 'branch_id = ?',
      whereArgs: [branchId],
      orderBy: 'created_at DESC',
      limit: limit,
      offset: offset,
    );

    final transactions = <Transaction>[];
    for (final t in transactionMaps) {
      final itemMaps = await db.query('transaction_items',
        where: 'transaction_id = ?',
        whereArgs: [t['id']],
      );

      final items = itemMaps.map((i) => TransactionItem(
        productId: i['product_id'] as String,
        productName: i['product_name'] as String,
        barcode: i['barcode'] as String,
        price: (i['price'] as num).toDouble(),
        quantity: (i['quantity'] as num).toInt(),
        discount: (i['discount'] as num?)?.toDouble() ?? 0,
      )).toList();

      transactions.add(Transaction(
        id: t['id'] as String,
        branchId: t['branch_id'] as String,
        cashierId: t['cashier_id'] as String,
        cashierName: t['cashier_name'] as String,
        items: items,
        total: (t['total'] as num).toDouble(),
        discountTotal: (t['discount_total'] as num?)?.toDouble() ?? 0,
        taxRate: (t['tax_rate'] as num?)?.toDouble() ?? 0.0,
        taxAmount: (t['tax_amount'] as num?)?.toDouble() ?? 0.0,
        grandTotal: (t['grand_total'] as num).toDouble(),
        paymentMethod: t['payment_method'] as String,
        amountPaid: (t['amount_paid'] as num).toDouble(),
        change: (t['change_amount'] as num).toDouble(),
        createdAt: DateTime.parse(t['created_at'] as String),
        receiptNumber: t['receipt_number'] as String?,
      ));
    }

    return transactions;
  }
}
