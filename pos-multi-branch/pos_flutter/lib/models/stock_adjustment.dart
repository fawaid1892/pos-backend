class StockAdjustment {
  final String id;
  final String productId;
  final String productName;
  final String branchId;
  final int quantity; // positive = stock in, negative = stock out
  final String reason;
  final String type; // 'in' | 'out'
  final DateTime createdAt;

  StockAdjustment({
    required this.id,
    required this.productId,
    required this.productName,
    required this.branchId,
    required this.quantity,
    required this.reason,
    required this.type,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'productId': productId,
        'productName': productName,
        'branchId': branchId,
        'quantity': quantity,
        'reason': reason,
        'type': type,
        'createdAt': createdAt.toIso8601String(),
      };

  factory StockAdjustment.fromJson(Map<String, dynamic> json) =>
      StockAdjustment(
        id: json['id'] as String,
        productId: json['productId'] as String,
        productName: json['productName'] as String,
        branchId: json['branchId'] as String,
        quantity: json['quantity'] as int,
        reason: json['reason'] as String,
        type: json['type'] as String,
        createdAt: DateTime.parse(json['createdAt'] as String),
      );
}

class StockTransfer {
  final String id;
  final String sourceBranchId;
  final String sourceBranchName;
  final String targetBranchId;
  final String targetBranchName;
  final String productId;
  final String productName;
  final int quantity;
  final String status; // 'pending', 'completed', 'cancelled'
  final DateTime createdAt;

  StockTransfer({
    required this.id,
    required this.sourceBranchId,
    required this.sourceBranchName,
    required this.targetBranchId,
    required this.targetBranchName,
    required this.productId,
    required this.productName,
    required this.quantity,
    required this.status,
    required this.createdAt,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'sourceBranchId': sourceBranchId,
        'sourceBranchName': sourceBranchName,
        'targetBranchId': targetBranchId,
        'targetBranchName': targetBranchName,
        'productId': productId,
        'productName': productName,
        'quantity': quantity,
        'status': status,
        'createdAt': createdAt.toIso8601String(),
      };

  factory StockTransfer.fromJson(Map<String, dynamic> json) => StockTransfer(
        id: json['id'] as String,
        sourceBranchId: json['sourceBranchId'] as String,
        sourceBranchName: json['sourceBranchName'] as String,
        targetBranchId: json['targetBranchId'] as String,
        targetBranchName: json['targetBranchName'] as String,
        productId: json['productId'] as String,
        productName: json['productName'] as String,
        quantity: json['quantity'] as int,
        status: json['status'] as String,
        createdAt: DateTime.parse(json['createdAt'] as String),
      );
}

class StockAlert {
  final ProductStock product;
  final int minimumStock;
  final bool isLow;

  StockAlert({
    required this.product,
    required this.minimumStock,
    required this.isLow,
  });
}

class ProductStock {
  final String productId;
  final String productName;
  final String barcode;
  final int currentStock;
  final int minimumStock;
  final String branchId;

  ProductStock({
    required this.productId,
    required this.productName,
    required this.barcode,
    required this.currentStock,
    this.minimumStock = 5,
    required this.branchId,
  });

  bool get isLowStock => currentStock <= minimumStock;
}