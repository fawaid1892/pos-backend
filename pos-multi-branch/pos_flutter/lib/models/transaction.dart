class TransactionItem {
  final String productId;
  final String productName;
  final String barcode;
  final double price;
  int quantity;
  double discount;

  TransactionItem({
    required this.productId,
    required this.productName,
    required this.barcode,
    required this.price,
    this.quantity = 1,
    this.discount = 0.0,
  });

  double get subtotal => (price * quantity) - discount;

  Map<String, dynamic> toJson() => {
        'productId': productId,
        'productName': productName,
        'barcode': barcode,
        'price': price,
        'quantity': quantity,
        'discount': discount,
        'subtotal': subtotal,
      };

  factory TransactionItem.fromJson(Map<String, dynamic> json) =>
      TransactionItem(
        productId: json['productId'] as String,
        productName: json['productName'] as String,
        barcode: json['barcode'] as String,
        price: (json['price'] as num).toDouble(),
        quantity: json['quantity'] as int? ?? 1,
        discount: (json['discount'] as num?)?.toDouble() ?? 0.0,
      );
}

class Transaction {
  final String id;
  final String branchId;
  final String cashierId;
  final String cashierName;
  final List<TransactionItem> items;
  final double total;
  final double discountTotal;
  final double grandTotal;
  final String paymentMethod;
  final double amountPaid;
  final double change;
  final DateTime createdAt;
  final String? receiptNumber;

  Transaction({
    required this.id,
    required this.branchId,
    required this.cashierId,
    required this.cashierName,
    required this.items,
    required this.total,
    this.discountTotal = 0.0,
    required this.grandTotal,
    required this.paymentMethod,
    required this.amountPaid,
    required this.change,
    required this.createdAt,
    this.receiptNumber,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'branchId': branchId,
        'cashierId': cashierId,
        'cashierName': cashierName,
        'items': items.map((i) => i.toJson()).toList(),
        'total': total,
        'discountTotal': discountTotal,
        'grandTotal': grandTotal,
        'paymentMethod': paymentMethod,
        'amountPaid': amountPaid,
        'change': change,
        'createdAt': createdAt.toIso8601String(),
        'receiptNumber': receiptNumber,
      };

  factory Transaction.fromJson(Map<String, dynamic> json) => Transaction(
        id: json['id'] as String,
        branchId: json['branchId'] as String,
        cashierId: json['cashierId'] as String,
        cashierName: json['cashierName'] as String,
        items: (json['items'] as List)
            .map((i) => TransactionItem.fromJson(i as Map<String, dynamic>))
            .toList(),
        total: (json['total'] as num).toDouble(),
        discountTotal: (json['discountTotal'] as num?)?.toDouble() ?? 0.0,
        grandTotal: (json['grandTotal'] as num).toDouble(),
        paymentMethod: json['paymentMethod'] as String,
        amountPaid: (json['amountPaid'] as num).toDouble(),
        change: (json['change'] as num).toDouble(),
        createdAt: DateTime.parse(json['createdAt'] as String),
        receiptNumber: json['receiptNumber'] as String?,
      );
}
