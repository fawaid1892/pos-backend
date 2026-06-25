import 'package:flutter/material.dart';

/// Generic error state widget with retry button.
///
/// Shows an error icon, message, and optional retry action.
class ErrorStateWidget extends StatelessWidget {
  final String message;
  final String? title;
  final IconData icon;
  final VoidCallback? onRetry;
  final String? retryLabel;

  const ErrorStateWidget({
    super.key,
    required this.message,
    this.title,
    this.icon = Icons.error_outline,
    this.onRetry,
    this.retryLabel,
  });

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              size: 64,
              color: theme.colorScheme.error,
            ),
            const SizedBox(height: 16),
            if (title != null) ...[
              Text(
                title!,
                style: theme.textTheme.titleMedium?.copyWith(
                  fontWeight: FontWeight.bold,
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
            ],
            Text(
              message,
              style: theme.textTheme.bodyMedium?.copyWith(
                color: theme.colorScheme.onSurfaceVariant,
              ),
              textAlign: TextAlign.center,
            ),
            if (onRetry != null) ...[
              const SizedBox(height: 24),
              FilledButton.tonalIcon(
                onPressed: onRetry,
                icon: const Icon(Icons.refresh),
                label: Text(retryLabel ?? 'Coba Lagi'),
              ),
            ],
          ],
        ),
      ),
    );
  }
}

/// Convenience widget that wraps a Future/AsyncSnapshot and shows
/// loading / error / data states automatically.
class AsyncSnapshotWidget<T> extends StatelessWidget {
  final AsyncSnapshot<T> snapshot;
  final Widget Function(T data) dataBuilder;
  final Widget? loadingWidget;
  final String? errorMessage;
  final VoidCallback? onRetry;

  const AsyncSnapshotWidget({
    super.key,
    required this.snapshot,
    required this.dataBuilder,
    this.loadingWidget,
    this.errorMessage,
    this.onRetry,
  });

  @override
  Widget build(BuildContext context) {
    if (snapshot.connectionState == ConnectionState.waiting) {
      return loadingWidget ??
          const Center(child: CircularProgressIndicator());
    }

    if (snapshot.hasError) {
      return ErrorStateWidget(
        message: errorMessage ??
            'Terjadi kesalahan: ${snapshot.error}',
        onRetry: onRetry,
      );
    }

    if (!snapshot.hasData || snapshot.data == null) {
      return ErrorStateWidget(
        message: 'Tidak ada data yang tersedia',
        onRetry: onRetry,
      );
    }

    return dataBuilder(snapshot.data as T);
  }
}
