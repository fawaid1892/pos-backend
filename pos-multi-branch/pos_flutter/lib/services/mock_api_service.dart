import '../models/product.dart';
import '../models/transaction.dart';

/// Mock API service that simulates backend calls.
/// In production, replace with actual API client (dio/http + Supabase).
class MockApiService {
  static final MockApiService _instance = MockApiService._();
  factory MockApiService() => _instance;
  MockApiService._();

  final List<Product> _mockProducts = [
    Product(
      id: 'prod_001',
      name: 'Kopi Hitam',
      barcode: '8991001001001',
      price: 15000,
      category: 'Minuman',
      stock: 50,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_002',
      name: 'Kopi Susu',
      barcode: '8991001001002',
      price: 20000,
      category: 'Minuman',
      stock: 40,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_003',
      name: 'Nasi Goreng',
      barcode: '8991001001003',
      price: 35000,
      category: 'Makanan',
      stock: 25,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_004',
      name: 'Mie Goreng',
      barcode: '8991001001004',
      price: 25000,
      category: 'Makanan',
      stock: 30,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_005',
      name: 'Air Mineral',
      barcode: '8991001001005',
      price: 5000,
      category: 'Minuman',
      stock: 100,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_006',
      name: 'Teh Manis',
      barcode: '8991001001006',
      price: 8000,
      category: 'Minuman',
      stock: 60,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_007',
      name: 'Kentang Goreng',
      barcode: '8991001001007',
      price: 20000,
      category: 'Makanan',
      stock: 20,
      branchId: 'branch_001',
    ),
    Product(
      id: 'prod_008',
      name: 'Jus Jeruk',
      barcode: '8991001001008',
      price: 18000,
      category: 'Minuman',
      stock: 35,
      branchId: 'branch_001',
    ),
  ];

  List<Transaction> _transactions = [];

  Future<List<Product>> getProducts(String branchId) async {
    await Future.delayed(const Duration(milliseconds: 300));
    return _mockProducts.where((p) => p.branchId == branchId).toList();
  }

  Future<Product?> getProductByBarcode(String barcode, String branchId) async {
    await Future.delayed(const Duration(milliseconds: 200));
    return _mockProducts.firstWhere(
      (p) => p.barcode == barcode && p.branchId == branchId,
      orElse: () => _mockProducts.first,
    );
  }

  Future<List<Product>> searchProducts(
      String query, String branchId) async {
    await Future.delayed(const Duration(milliseconds: 200));
    final q = query.toLowerCase();
    return _mockProducts.where((p) {
      return p.branchId == branchId &&
          (p.name.toLowerCase().contains(q) ||
              p.barcode.contains(q));
    }).toList();
  }

  Future<Transaction> submitTransaction(Transaction transaction) async {
    await Future.delayed(const Duration(milliseconds: 500));
    _transactions.insert(0, transaction);
    return transaction;
  }

  Future<List<Transaction>> getTransactions(String branchId) async {
    await Future.delayed(const Duration(milliseconds: 300));
    return _transactions
        .where((t) => t.branchId == branchId)
        .toList();
  }
}
