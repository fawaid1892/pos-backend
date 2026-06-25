import 'package:flutter/material.dart';
import '../models/transaction.dart';

class CartItemTile extends StatelessWidget {
  final TransactionItem item;
  final VoidCallback onIncrement;
  final VoidCallback onDecrement;
  final VoidCallback onRemove;

  const CartItemTile({
    super.key,
    required this.item,
    required this.onIncrement,
    required this.onDecrement,
    required this.onRemove,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Row(
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.productName,
                    style: TextStyle(
                      fontWeight: FontWeight.w600,
                      fontSize: 15,
                      color: colorScheme.onSurface,
                    ),
                  ),
                  const SizedBox(height: 2),
                  Text(
                    'Rp ${_formatPrice(item.price)}',
                    style: TextStyle(
                      color: colorScheme.primary,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                  if (item.discount > 0)
                    Text(
                      'Diskon: Rp ${_formatPrice(item.discount)}',
                      style: TextStyle(
                        fontSize: 12,
                        color: colorScheme.error,
                      ),
                    ),
                ],
              ),
            ),
            Row(
              children: [
                IconButton(
                  icon: const Icon(Icons.remove_circle_outline),
                  onPressed: onDecrement,
                  iconSize: 28,
                ),
                SizedBox(
                  width: 32,
                  child: Text(
                    '${item.quantity}',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: colorScheme.onSurface,
                    ),
                  ),
                ),
                IconButton(
                  icon: const Icon(Icons.add_circle_outline),
                  onPressed: onIncrement,
                  iconSize: 28,
                ),
              ],
            ),
            SizedBox(
              width: 90,
              child: Text(
                'Rp ${_formatPrice(item.subtotal)}',
                textAlign: TextAlign.right,
                style: TextStyle(
                  fontWeight: FontWeight.bold,
                  fontSize: 15,
                  color: colorScheme.onSurface,
                ),
              ),
            ),
            IconButton(
              icon: const Icon(Icons.delete_outline, color: Colors.red),
              onPressed: onRemove,
              iconSize: 22,
            ),
          ],
        ),
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
