class Branch {
  final String id;
  final String name;
  final String? address;
  final String? phone;

  Branch({
    required this.id,
    required this.name,
    this.address,
    this.phone,
  });

  Map<String, dynamic> toJson() => {
        'id': id,
        'name': name,
        'address': address,
        'phone': phone,
      };

  factory Branch.fromJson(Map<String, dynamic> json) => Branch(
        id: json['id'] as String,
        name: json['name'] as String,
        address: json['address'] as String?,
        phone: json['phone'] as String?,
      );
}