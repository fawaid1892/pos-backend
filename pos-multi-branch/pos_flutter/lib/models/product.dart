class Product {
  final String id;
  final String name;
  final String barcode;
  final double price;
  final String category;
  final String? imageUrl;
  final int stock;
  final String branchId;

  Product({
    required this.id,
    required this.name,
    required this.barcode,
    required this.price,
    required this.category,
    this.imageUrl,
    this.stock = 0,
    required this.branchId,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'barcode': barcode,
        'price': price,
        'category': category,
        'imageUrl': imageUrl,
        'stock': stock,
        'branchId': branchId,
      };

  factory Product.fromJson(Map<String, dynamic> json) => Product(
        id: json['id'] as String,
        name: json['name'] as String,
        barcode: json['barcode'] as String,
        price: (json['price'] as num).toDouble(),
        category: json['category'] as String,
        imageUrl: json['imageUrl'] as String?,
        stock: json['stock'] as int? ?? 0,
        branchId: json['branchId'] as String,
      );

  Product copyWith({
    String? id,
    String? name,
    String? barcode,
    double? price,
    String? category,
    String? imageUrl,
    int? stock,
    String? branchId,
  }) =>
      Product(
        id: id ?? this.id,
        name: name ?? this.name,
        barcode: barcode ?? this.barcode,
        price: price ?? this.price,
        category: category ?? this.category,
        imageUrl: imageUrl ?? this.imageUrl,
        stock: stock ?? this.stock,
        branchId: branchId ?? this.branchId,
      );
}
