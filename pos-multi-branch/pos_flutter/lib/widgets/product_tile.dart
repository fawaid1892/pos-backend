import 'package:flutter/material.dart';
import '../models/product.dart';
import '../utils/responsive.dart';

class ProductTile extends StatelessWidget {
  final Product product;
  final VoidCallback onAdd;

  const ProductTile({
    super.key,
    required this.product,
    required this.onAdd,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;
    final isTablet = context.isTablet;

    return Card(
      elevation: 1,
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: colorScheme.primaryContainer,
          child: Text(
            product.name[0].toUpperCase(),
            style: TextStyle(
              color: colorScheme.onPrimaryContainer,
              fontWeight: FontWeight.bold,
            ),
          ),
        ),
        title: Text(
          product.name,
          style: const TextStyle(fontWeight: FontWeight.w600),
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Barcode: ${product.barcode}',
              style: TextStyle(color: colorScheme.onSurfaceVariant),
            ),
            Text(
              'Stok: ${product.stock}',
              style: TextStyle(color: colorScheme.onSurfaceVariant),
            ),
          ],
        ),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              'Rp ${_formatPrice(product.price)}',
              style: TextStyle(
                fontWeight: FontWeight.bold,
                fontSize: isTablet ? 18 : 16,
                color: colorScheme.primary,
              ),
            ),
            if (product.stock <= 5)
              Text(
                'Stok menipis!',
                style: TextStyle(
                  fontSize: 11,
                  color: colorScheme.error,
                ),
              ),
          ],
        ),
        onTap: onAdd,
      ),
    );
  }

  String _formatPrice(double amount) {
    return amount.toStringAsFixed(0).replaceAllMapped(
          RegExp(r'(\d)(?=(\d{3})+(?!\d))'),
          (match) => '${match.group(1)}.',
        );
  }
}
