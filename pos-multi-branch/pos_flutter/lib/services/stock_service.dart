import '../database/local_database.dart';
import '../database/daos/stock_mutation_dao.dart';
import '../models/branch.dart';
import '../models/stock_adjustment.dart';

/// Service for stock operations backed by SQLite.
///
/// Replaces MockStockService for stock inventory, adjustments, and transfers.
class StockService {
  static final StockService _instance = StockService._internal();
  factory StockService() => _instance;
  StockService._internal();

  final LocalDatabase _db = LocalDatabase();
  final StockMutationDao _mutationDao = StockMutationDao();

  /// Get all branches from local DB.
  Future<List<Branch>> getBranches() async {
    final db = await _db.database;
    final maps = await db.query('branches', orderBy: 'name ASC');
    return maps.map((m) => Branch.fromJson(m)).toList();
  }

  /// Get inventory for a branch (from branch_products + products).
  Future<List<ProductStock>> getInventory(String branchId) async {
    final db = await _db.database;

    final maps = await db.rawQuery('''
      SELECT
        bp.product_id,
        p.name AS product_name,
        p.barcode,
        bp.stock AS current_stock,
        bp.minimum_stock,
        bp.branch_id
      FROM branch_products bp
      INNER JOIN products p ON p.id = bp.product_id
      WHERE bp.branch_id = ?
      ORDER BY p.name ASC
    ''', [branchId]);

    return maps.map((m) => ProductStock(
      productId: m['product_id'] as String,
      productName: m['product_name'] as String,
      barcode: m['barcode'] as String,
      currentStock: (m['current_stock'] as num).toInt(),
      minimumStock: (m['minimum_stock'] as num).toInt(),
      branchId: m['branch_id'] as String,
    )).toList();
  }

  /// Adjust stock (stock in / stock out).
  Future<StockAdjustment> adjustStock({
    required String branchId,
    required String productId,
    required int quantity,
    required String reason,
    required String type,
  }) async {
    final db = await _db.database;
    final adjustedQty = type == 'in' ? quantity : -quantity;

    // 1. Update branch_products stock count
    await db.rawUpdate('''
      UPDATE branch_products
      SET stock = stock + ?, pending_sync = 1, sync_status = 'pending'
      WHERE branch_id = ? AND product_id = ?
    ''', [adjustedQty, branchId, productId]);

    // 2. Get product name
    final productRows = await db.query('products',
      where: 'id = ?', whereArgs: [productId]);
    final productName = productRows.isNotEmpty
        ? (productRows.first['name'] as String)
        : 'Unknown';

    // 3. Create stock mutation record
    final mutation = StockAdjustment(
      id: 'adj_${DateTime.now().millisecondsSinceEpoch}',
      productId: productId,
      productName: productName,
      branchId: branchId,
      quantity: adjustedQty,
      reason: reason,
      type: type,
      createdAt: DateTime.now(),
    );

    await _mutationDao.insert(mutation.toJson());

    return mutation;
  }

  /// Transfer stock between branches.
  Future<StockTransfer> transferStock({
    required String sourceBranchId,
    required String targetBranchId,
    required String productId,
    required int quantity,
  }) async {
    final db = await _db.database;

    // 1. Deduct from source
    await db.rawUpdate('''
      UPDATE branch_products
      SET stock = stock - ?, pending_sync = 1, sync_status = 'pending'
      WHERE branch_id = ? AND product_id = ?
    ''', [quantity, sourceBranchId, productId]);

    // 2. Add to target (insert if not exists)
    final existingTarget = await db.query('branch_products',
      where: 'branch_id = ? AND product_id = ?',
      whereArgs: [targetBranchId, productId]);

    if (existingTarget.isNotEmpty) {
      await db.rawUpdate('''
        UPDATE branch_products
        SET stock = stock + ?, pending_sync = 1, sync_status = 'pending'
        WHERE branch_id = ? AND product_id = ?
      ''', [quantity, targetBranchId, productId]);
    } else {
      // Create new branch_product entry for target
      final id = '${targetBranchId}_$productId';
      await db.insert('branch_products', {
        'id': id,
        'branch_id': targetBranchId,
        'product_id': productId,
        'stock': quantity,
        'minimum_stock': 5,
        'pending_sync': 1,
        'sync_status': 'pending',
      });
    }

    // 3. Get names
    final productRows = await db.query('products',
      where: 'id = ?', whereArgs: [productId]);
    final productName = productRows.isNotEmpty
        ? (productRows.first['name'] as String)
        : 'Unknown';

    final sourceName = await _getBranchName(sourceBranchId);
    final targetName = await _getBranchName(targetBranchId);

    // 4. Create stock mutation records
    for (final entry in [
      {'branch_id': sourceBranchId, 'type': 'transfer_out'},
      {'branch_id': targetBranchId, 'type': 'transfer_in'},
    ]) {
      await _mutationDao.insert({
        'id': 'mut_${DateTime.now().millisecondsSinceEpoch}_${entry['type']}',
        'branch_id': entry['branch_id'],
        'product_id': productId,
        'product_name': productName,
        'quantity': entry['type'] == 'transfer_out' ? -quantity : quantity,
        'reason': 'Transfer antar cabang',
        'type': entry['type'],
        'created_at': DateTime.now().toIso8601String(),
        'pending_sync': 1,
        'sync_status': 'pending',
      });
    }

    return StockTransfer(
      id: 'trf_${DateTime.now().millisecondsSinceEpoch}',
      sourceBranchId: sourceBranchId,
      sourceBranchName: sourceName,
      targetBranchId: targetBranchId,
      targetBranchName: targetName,
      productId: productId,
      productName: productName,
      quantity: quantity,
      status: 'completed',
      createdAt: DateTime.now(),
    );
  }

  /// Get stock alerts (stock <= minimum_stock).
  Future<List<ProductStock>> getStockAlerts(String branchId,
      {int minimumStock = 5}) async {
    final db = await _db.database;

    final maps = await db.rawQuery('''
      SELECT
        bp.product_id,
        p.name AS product_name,
        p.barcode,
        bp.stock AS current_stock,
        bp.minimum_stock,
        bp.branch_id
      FROM branch_products bp
      INNER JOIN products p ON p.id = bp.product_id
      WHERE bp.branch_id = ? AND bp.stock <= ?
      ORDER BY bp.stock ASC
    ''', [branchId, minimumStock]);

    return maps.map((m) => ProductStock(
      productId: m['product_id'] as String,
      productName: m['product_name'] as String,
      barcode: m['barcode'] as String,
      currentStock: (m['current_stock'] as num).toInt(),
      minimumStock: (m['minimum_stock'] as num).toInt(),
      branchId: m['branch_id'] as String,
    )).toList();
  }

  /// Search inventory locally.
  Future<List<ProductStock>> searchInventory(
      String branchId, String query) async {
    final db = await _db.database;
    final q = '%$query%';

    final maps = await db.rawQuery('''
      SELECT
        bp.product_id,
        p.name AS product_name,
        p.barcode,
        bp.stock AS current_stock,
        bp.minimum_stock,
        bp.branch_id
      FROM branch_products bp
      INNER JOIN products p ON p.id = bp.product_id
      WHERE bp.branch_id = ?
        AND (p.name LIKE ? OR p.barcode LIKE ?)
      ORDER BY p.name ASC
    ''', [branchId, q, q]);

    return maps.map((m) => ProductStock(
      productId: m['product_id'] as String,
      productName: m['product_name'] as String,
      barcode: m['barcode'] as String,
      currentStock: (m['current_stock'] as num).toInt(),
      minimumStock: (m['minimum_stock'] as num).toInt(),
      branchId: m['branch_id'] as String,
    )).toList();
  }

  Future<String> _getBranchName(String branchId) async {
    final db = await _db.database;
    final rows = await db.query('branches',
      where: 'id = ?', whereArgs: [branchId]);
    if (rows.isEmpty) return 'Unknown';
    return rows.first['name'] as String;
  }
}
