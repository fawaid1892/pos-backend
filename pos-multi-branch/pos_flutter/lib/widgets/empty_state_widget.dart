import 'package:flutter/material.dart';

/// Generic empty state widget with illustration icon and message.
///
/// Shows when a list, search result, or data view has no items.
class EmptyStateWidget extends StatelessWidget {
  final IconData icon;
  final String title;
  final String? subtitle;
  final Widget? action;
  final double iconSize;

  const EmptyStateWidget({
    super.key,
    this.icon = Icons.inbox_outlined,
    required this.title,
    this.subtitle,
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.search({
    super.key,
    this.icon = Icons.search_off,
    this.title = 'Produk tidak ditemukan',
    this.subtitle = 'Coba ubah kata kunci pencarian',
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.cart({
    super.key,
    this.icon = Icons.shopping_cart_outlined,
    this.title = 'Keranjang kosong',
    this.subtitle = 'Tambahkan produk dari daftar produk',
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.stock({
    super.key,
    this.icon = Icons.inventory_2_outlined,
    this.title = 'Tidak ada data inventory',
    this.subtitle = 'Belum ada produk di cabang ini',
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.report({
    super.key,
    this.icon = Icons.bar_chart_outlined,
    this.title = 'Belum ada data laporan',
    this.subtitle = 'Pilih rentang tanggal untuk melihat laporan',
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.transaction({
    super.key,
    this.icon = Icons.receipt_long_outlined,
    this.title = 'Belum ada transaksi',
    this.subtitle = 'Lakukan transaksi penjualan terlebih dahulu',
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.alert({
    super.key,
    this.icon = Icons.check_circle_outline,
    this.title = 'Semua stok aman',
    this.subtitle = 'Tidak ada produk yang perlu di-adjust',
    this.action,
    this.iconSize = 80,
  });

  const EmptyStateWidget.sync({
    super.key,
    this.icon = Icons.cloud_done,
    this.title = 'Semua tersinkronisasi',
    this.subtitle = 'Tidak ada item yang perlu disinkronkan',
    this.action,
    this.iconSize = 80,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(32),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              size: iconSize,
              color: theme.colorScheme.onSurfaceVariant.withOpacity(0.4),
            ),
            const SizedBox(height: 16),
            Text(
              title,
              style: theme.textTheme.titleMedium?.copyWith(
                fontWeight: FontWeight.w600,
                color: theme.colorScheme.onSurfaceVariant,
              ),
              textAlign: TextAlign.center,
            ),
            if (subtitle != null) ...[
              const SizedBox(height: 8),
              Text(
                subtitle!,
                style: theme.textTheme.bodyMedium?.copyWith(
                  color: theme.colorScheme.onSurfaceVariant
                      .withOpacity(0.7),
                ),
                textAlign: TextAlign.center,
              ),
            ],
            if (action != null) ...[
              const SizedBox(height: 24),
              action!,
            ],
          ],
        ),
      ),
    );
  }
}
