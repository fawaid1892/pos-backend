import 'package:flutter/foundation.dart';
import '../models/transaction.dart';
import '../models/product.dart';

class CartProvider extends ChangeNotifier {
  List<TransactionItem> _items = [];
  double _discountTotal = 0.0;
  double _taxRate = 0.0; // e.g. 0.11 for 11% PPN
  String _selectedPaymentMethod = 'Tunai';

  List<TransactionItem> get items => List.unmodifiable(_items);
  double get discountTotal => _discountTotal;
  double get taxRate => _taxRate;
  String get selectedPaymentMethod => _selectedPaymentMethod;

  int get itemCount => _items.fold(0, (sum, item) => sum + item.quantity);

  double get total => _items.fold(0.0, (sum, item) => sum + item.subtotal);

  double get taxableAmount => (total - _discountTotal).clamp(0, double.infinity);

  double get taxAmount => taxableAmount * _taxRate;

  double get grandTotal => taxableAmount + taxAmount;

  bool get isEmpty => _items.isEmpty;

  void addProduct(Product product, {int quantity = 1}) {
    final existingIndex =
        _items.indexWhere((item) => item.productId == product.id);

    if (existingIndex >= 0) {
      _items[existingIndex].quantity += quantity;
    } else {
      _items.add(TransactionItem(
        productId: product.id,
        productName: product.name,
        barcode: product.barcode,
        price: product.price,
        quantity: quantity,
      ));
    }

    notifyListeners();
  }

  void updateQuantity(String productId, int newQuantity) {
    final index = _items.indexWhere((item) => item.productId == productId);
    if (index >= 0) {
      if (newQuantity <= 0) {
        _items.removeAt(index);
      } else {
        _items[index].quantity = newQuantity;
      }
      notifyListeners();
    }
  }

  void removeItem(String productId) {
    _items.removeWhere((item) => item.productId == productId);
    notifyListeners();
  }

  void setItemDiscount(String productId, double discount) {
    final index = _items.indexWhere((item) => item.productId == productId);
    if (index >= 0) {
      _items[index].discount = discount;
      notifyListeners();
    }
  }

  void setCartDiscount(double discount) {
    _discountTotal = discount;
    notifyListeners();
  }

  void setTaxRate(double rate) {
    _taxRate = rate;
    notifyListeners();
  }

  void setPaymentMethod(String method) {
    _selectedPaymentMethod = method;
    notifyListeners();
  }

  void clearCart() {
    _items.clear();
    _discountTotal = 0.0;
    _taxRate = 0.0;
    _selectedPaymentMethod = 'Tunai';
    notifyListeners();
  }

  /// Build a Transaction from current cart state.
  Transaction buildTransaction({
    required String id,
    required String branchId,
    required String cashierId,
    required String cashierName,
    required double amountPaid,
    required double change,
    String? receiptNumber,
  }) {
    return Transaction(
      id: id,
      branchId: branchId,
      cashierId: cashierId,
      cashierName: cashierName,
      items: _items.map((item) {
        return TransactionItem(
          productId: item.productId,
          productName: item.productName,
          barcode: item.barcode,
          price: item.price,
          quantity: item.quantity,
          discount: item.discount,
        );
      }).toList(),
      total: total,
      discountTotal: _discountTotal,
      taxRate: _taxRate,
      taxAmount: taxAmount,
      grandTotal: grandTotal,
      paymentMethod: _selectedPaymentMethod,
      amountPaid: amountPaid,
      change: change,
      createdAt: DateTime.now(),
      receiptNumber: receiptNumber,
    );
  }
}
