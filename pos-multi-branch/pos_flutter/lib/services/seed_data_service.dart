import '../database/local_database.dart';

/// Seeds initial data into local SQLite on first run.
///
/// Called once in main.dart after database initialization.
class SeedDataService {
  static final SeedDataService _instance = SeedDataService._internal();
  factory SeedDataService() => _instance;
  SeedDataService._internal();

  final LocalDatabase _db = LocalDatabase();

  /// Seed all initial data if the database is empty.
  Future<void> seedIfEmpty() async {
    final db = await _db.database;

    // Check if data already exists
    final count = Sqflite.firstIntValue(
      await db.rawQuery('SELECT COUNT(*) FROM branches'),
    );
    if (count != null && count > 0) return;

    // ── Branches ──
    await db.insert('branches', {
      'id': 'branch_001', 'name': 'Cabang Utama',
      'address': 'Jl. Merdeka No.1', 'phone': '021-1234567',
      'pending_sync': 0, 'sync_status': 'synced',
    });
    await db.insert('branches', {
      'id': 'branch_002', 'name': 'Cabang Kedua',
      'address': 'Jl. Sudirman No.45', 'phone': '021-7654321',
      'pending_sync': 0, 'sync_status': 'synced',
    });
    await db.insert('branches', {
      'id': 'branch_003', 'name': 'Cabang Ketiga',
      'address': 'Jl. Gatot Subroto No.88', 'phone': '021-5551234',
      'pending_sync': 0, 'sync_status': 'synced',
    });

    // ── Users ──
    await db.insert('users', {
      'id': 'user_001', 'email': 'owner@example.com',
      'name': 'Owner', 'role': 'owner',
      'pending_sync': 0, 'sync_status': 'synced',
    });
    await db.insert('users', {
      'id': 'user_002', 'email': 'kasir@example.com',
      'name': 'Kasir', 'role': 'cashier', 'branch_id': 'branch_001',
      'pending_sync': 0, 'sync_status': 'synced',
    });

    // ── Categories ──
    await db.insert('categories', {
      'id': 'cat_001', 'name': 'Minuman',
      'pending_sync': 0, 'sync_status': 'synced',
    });
    await db.insert('categories', {
      'id': 'cat_002', 'name': 'Makanan',
      'pending_sync': 0, 'sync_status': 'synced',
    });

    // ── Products ──
    final products = [
      {'id': 'prod_001', 'name': 'Kopi Hitam', 'barcode': '8991001001001', 'price': 15000, 'category': 'Minuman'},
      {'id': 'prod_002', 'name': 'Kopi Susu', 'barcode': '8991001001002', 'price': 20000, 'category': 'Minuman'},
      {'id': 'prod_003', 'name': 'Nasi Goreng', 'barcode': '8991001001003', 'price': 35000, 'category': 'Makanan'},
      {'id': 'prod_004', 'name': 'Mie Goreng', 'barcode': '8991001001004', 'price': 25000, 'category': 'Makanan'},
      {'id': 'prod_005', 'name': 'Air Mineral', 'barcode': '8991001001005', 'price': 5000, 'category': 'Minuman'},
      {'id': 'prod_006', 'name': 'Teh Manis', 'barcode': '8991001001006', 'price': 8000, 'category': 'Minuman'},
      {'id': 'prod_007', 'name': 'Kentang Goreng', 'barcode': '8991001001007', 'price': 20000, 'category': 'Makanan'},
      {'id': 'prod_008', 'name': 'Jus Jeruk', 'barcode': '8991001001008', 'price': 18000, 'category': 'Minuman'},
    ];

    for (final p in products) {
      await db.insert('products', {
        ...p,
        'pending_sync': 0, 'sync_status': 'synced',
      });
    }

    // ── Branch Products (inventory per branch) ──
    final branchStocks = {
      'branch_001': {'prod_001': 50, 'prod_002': 40, 'prod_003': 25, 'prod_004': 30, 'prod_005': 100, 'prod_006': 60, 'prod_007': 20, 'prod_008': 3},
      'branch_002': {'prod_001': 12, 'prod_002': 8, 'prod_003': 5, 'prod_004': 0, 'prod_005': 45, 'prod_006': 22, 'prod_007': 2, 'prod_008': 15},
      'branch_003': {'prod_001': 30, 'prod_002': 20, 'prod_003': 0, 'prod_004': 10, 'prod_005': 80, 'prod_006': 15, 'prod_007': 7, 'prod_008': 25},
    };

    for (final branchEntry in branchStocks.entries) {
      for (final stockEntry in branchEntry.value.entries) {
        await db.insert('branch_products', {
          'id': '${branchEntry.key}_${stockEntry.key}',
          'branch_id': branchEntry.key,
          'product_id': stockEntry.key,
          'stock': stockEntry.value,
          'minimum_stock': 5,
          'pending_sync': 0, 'sync_status': 'synced',
        });
      }
    }
  }
}
