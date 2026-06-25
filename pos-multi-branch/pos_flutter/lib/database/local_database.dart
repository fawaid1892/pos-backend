import 'package:sqflite/sqflite.dart';
import 'package:path/path.dart';

/// Singleton local SQLite database mirroring Supabase schema + sync fields.
class LocalDatabase {
  static final LocalDatabase _instance = LocalDatabase._internal();
  factory LocalDatabase() => _instance;
  LocalDatabase._internal();

  static Database? _database;

  Future<Database> get database async {
    if (_database != null) return _database!;
    _database = await _initDatabase();
    return _database!;
  }

  Future<Database> _initDatabase() async {
    final dbPath = await getDatabasesPath();
    final path = join(dbPath, 'pos_multi_branch.db');

    return await openDatabase(
      path,
      version: 2,
      onCreate: _onCreate,
      onUpgrade: _onUpgrade,
      onConfigure: _onConfigure,
    );
  }

  Future<void> _onConfigure(Database db) async {
    // Enable foreign keys
    await db.execute('PRAGMA foreign_keys = ON');
  }

  Future<void> _onCreate(Database db, int version) async {
    // ── Users ──
    await db.execute('''
      CREATE TABLE users (
        id TEXT PRIMARY KEY,
        email TEXT NOT NULL,
        name TEXT NOT NULL,
        role TEXT NOT NULL DEFAULT 'cashier',
        branch_id TEXT,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        updated_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending'
      )
    ''');

    // ── Branches ──
    await db.execute('''
      CREATE TABLE branches (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        address TEXT,
        phone TEXT,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        updated_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending'
      )
    ''');

    // ── Categories ──
    await db.execute('''
      CREATE TABLE categories (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        updated_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending'
      )
    ''');

    // ── Products ──
    await db.execute('''
      CREATE TABLE products (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        barcode TEXT NOT NULL,
        price REAL NOT NULL DEFAULT 0,
        category TEXT NOT NULL DEFAULT 'General',
        image_url TEXT,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        updated_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending'
      )
    ''');

    // ── Branch Products (inventory per branch) ──
    await db.execute('''
      CREATE TABLE branch_products (
        id TEXT PRIMARY KEY,
        branch_id TEXT NOT NULL,
        product_id TEXT NOT NULL,
        stock INTEGER NOT NULL DEFAULT 0,
        minimum_stock INTEGER NOT NULL DEFAULT 5,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        updated_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending',
        FOREIGN KEY (branch_id) REFERENCES branches(id),
        FOREIGN KEY (product_id) REFERENCES products(id)
      )
    ''');

    // ── Transactions ──
    await db.execute('''
      CREATE TABLE transactions (
        id TEXT PRIMARY KEY,
        branch_id TEXT NOT NULL,
        cashier_id TEXT NOT NULL,
        cashier_name TEXT NOT NULL,
        total REAL NOT NULL DEFAULT 0,
        discount_total REAL NOT NULL DEFAULT 0,
        tax_rate REAL NOT NULL DEFAULT 0,
        tax_amount REAL NOT NULL DEFAULT 0,
        grand_total REAL NOT NULL DEFAULT 0,
        payment_method TEXT NOT NULL DEFAULT 'Tunai',
        amount_paid REAL NOT NULL DEFAULT 0,
        change_amount REAL NOT NULL DEFAULT 0,
        receipt_number TEXT,
        created_at TEXT NOT NULL,
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending',
        FOREIGN KEY (branch_id) REFERENCES branches(id),
        FOREIGN KEY (cashier_id) REFERENCES users(id)
      )
    ''');

    // ── Transaction Items ──
    await db.execute('''
      CREATE TABLE transaction_items (
        id TEXT PRIMARY KEY,
        transaction_id TEXT NOT NULL,
        product_id TEXT NOT NULL,
        product_name TEXT NOT NULL,
        barcode TEXT NOT NULL,
        price REAL NOT NULL DEFAULT 0,
        quantity INTEGER NOT NULL DEFAULT 1,
        discount REAL NOT NULL DEFAULT 0,
        subtotal REAL NOT NULL DEFAULT 0,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending',
        FOREIGN KEY (transaction_id) REFERENCES transactions(id),
        FOREIGN KEY (product_id) REFERENCES products(id)
      )
    ''');

    // ── Stock Mutations (adjustments + transfers) ──
    await db.execute('''
      CREATE TABLE stock_mutations (
        id TEXT PRIMARY KEY,
        branch_id TEXT NOT NULL,
        product_id TEXT NOT NULL,
        product_name TEXT NOT NULL,
        quantity INTEGER NOT NULL,
        reason TEXT,
        type TEXT NOT NULL CHECK(type IN ('in','out','transfer_in','transfer_out')),
        reference_id TEXT,
        reference_type TEXT,
        source_branch_id TEXT,
        target_branch_id TEXT,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        pending_sync INTEGER NOT NULL DEFAULT 1,
        synced_at TEXT,
        sync_status TEXT NOT NULL DEFAULT 'pending',
        FOREIGN KEY (branch_id) REFERENCES branches(id),
        FOREIGN KEY (product_id) REFERENCES products(id)
      )
    ''');

    // ── Sync Queue ──
    await db.execute('''
      CREATE TABLE sync_queue (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        table_name TEXT NOT NULL,
        record_id TEXT NOT NULL,
        action TEXT NOT NULL CHECK(action IN ('insert','update','delete')),
        payload TEXT NOT NULL,
        status TEXT NOT NULL DEFAULT 'pending',
        error_message TEXT,
        created_at TEXT NOT NULL DEFAULT (datetime('now')),
        processed_at TEXT
      )
    ''');

    // ── Indexes ──
    await db.execute(
        'CREATE INDEX idx_branch_products_branch ON branch_products(branch_id)');
    await db.execute(
        'CREATE INDEX idx_branch_products_product ON branch_products(product_id)');
    await db.execute(
        'CREATE INDEX idx_transactions_branch ON transactions(branch_id)');
    await db.execute(
        'CREATE INDEX idx_transaction_items_transaction ON transaction_items(transaction_id)');
    await db.execute(
        'CREATE INDEX idx_stock_mutations_branch ON stock_mutations(branch_id)');
    await db.execute(
        'CREATE INDEX idx_sync_queue_status ON sync_queue(status)');
    await db.execute(
        'CREATE INDEX idx_sync_queue_table ON sync_queue(table_name)');
  }

  Future<void> _onUpgrade(Database db, int oldVersion, int newVersion) async {
    if (oldVersion < 2) {
      // Add tax columns to transactions
      await db.execute('ALTER TABLE transactions ADD COLUMN tax_rate REAL NOT NULL DEFAULT 0');
      await db.execute('ALTER TABLE transactions ADD COLUMN tax_amount REAL NOT NULL DEFAULT 0');
    }
  }

  /// Drop and recreate all tables (for testing/reset).
  Future<void> resetDatabase() async {
    final db = await database;
    await db.execute('DROP TABLE IF EXISTS sync_queue');
    await db.execute('DROP TABLE IF EXISTS stock_mutations');
    await db.execute('DROP TABLE IF EXISTS transaction_items');
    await db.execute('DROP TABLE IF EXISTS transactions');
    await db.execute('DROP TABLE IF EXISTS branch_products');
    await db.execute('DROP TABLE IF EXISTS products');
    await db.execute('DROP TABLE IF EXISTS categories');
    await db.execute('DROP TABLE IF EXISTS branches');
    await db.execute('DROP TABLE IF EXISTS users');
    await _onCreate(db, 1);
  }

  /// Close database connection.
  Future<void> close() async {
    final db = await database;
    await db.close();
    _database = null;
  }
}
