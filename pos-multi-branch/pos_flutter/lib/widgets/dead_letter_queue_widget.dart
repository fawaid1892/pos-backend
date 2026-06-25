import 'dart:convert';
import 'package:flutter/material.dart';
import '../services/sync_service.dart';

/// Dead Letter Queue (DLQ) panel showing sync items that failed repeatedly.
///
/// Users can:
/// - View all failed sync items with their error messages
/// - Retry individual items manually
/// - Dismiss individual items
/// - Clear all resolved/old failed items
class DeadLetterQueuePanel extends StatefulWidget {
  const DeadLetterQueuePanel({super.key});

  @override
  State<DeadLetterQueuePanel> createState() => _DeadLetterQueuePanelState();
}

class _DeadLetterQueuePanelState extends State<DeadLetterQueuePanel> {
  final SyncService _syncService = SyncService();

  List<Map<String, dynamic>> _failedItems = [];
  bool _isLoading = true;
  bool _isRetrying = false;

  @override
  void initState() {
    super.initState();
    _loadFailedItems();
  }

  Future<void> _loadFailedItems() async {
    setState(() => _isLoading = true);
    try {
      final items = await _syncService.getDeadLetterQueueItems();
      if (mounted) {
        setState(() {
          _failedItems = items;
          _isLoading = false;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  Future<void> _retryItem(int queueId) async {
    setState(() => _isRetrying = true);
    final success = await _syncService.retryDeadLetterItem(queueId);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success
              ? 'Item returned to pending queue for retry'
              : 'Failed to retry item'),
          backgroundColor: success ? Colors.green : Colors.red,
        ),
      );
      _loadFailedItems();
      setState(() => _isRetrying = false);
    }
  }

  Future<void> _dismissItem(int queueId) async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Dismiss Item'),
        content: const Text('Remove this item from the failed queue?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Dismiss'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    final success = await _syncService.dismissDeadLetterItem(queueId);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text(success ? 'Item dismissed' : 'Failed to dismiss item'),
          backgroundColor: success ? Colors.green : Colors.red,
        ),
      );
      _loadFailedItems();
    }
  }

  Future<void> _clearAll() async {
    final count = _failedItems.length;
    if (count == 0) return;

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear Dead Letter Queue'),
        content: Text('Remove all $count failed item(s) from the queue?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Clear All'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    final cleared = await _syncService.clearDeadLetterQueue();
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Cleared $cleared failed item(s)'),
          backgroundColor: Colors.green,
        ),
      );
      _loadFailedItems();
    }
  }

  Future<void> _clearOld() async {
    final confirmed = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Clear Old Items'),
        content: const Text('Remove failed items older than 7 days?'),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            style: TextButton.styleFrom(foregroundColor: Colors.red),
            child: const Text('Clear Old'),
          ),
        ],
      ),
    );
    if (confirmed != true) return;

    final cleared = await _syncService.clearDeadLetterQueue(olderThanDays: 7);
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Cleared $cleared old failed item(s)'),
          backgroundColor: Colors.green,
        ),
      );
      _loadFailedItems();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Header with actions
        Padding(
          padding: const EdgeInsets.fromLTRB(12, 16, 12, 4),
          child: Row(
            children: [
              const Icon(Icons.error_outline, size: 18, color: Colors.red),
              const SizedBox(width: 6),
              Text(
                'Dead Letter Queue',
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.bold,
                  color: Colors.red.shade800,
                ),
              ),
              if (!_isLoading && _failedItems.isNotEmpty)
                Container(
                  margin: const EdgeInsets.only(left: 6),
                  padding:
                      const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
                  decoration: BoxDecoration(
                    color: Colors.red.shade100,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    '${_failedItems.length}',
                    style: TextStyle(
                        fontSize: 11,
                        fontWeight: FontWeight.bold,
                        color: Colors.red.shade800),
                  ),
                ),
              const Spacer(),
              // Clear actions
              if (_failedItems.isNotEmpty) ...[
                IconButton(
                  icon: const Icon(Icons.clean_hands, size: 18),
                  tooltip: 'Clear old items (>7d)',
                  onPressed: _clearOld,
                  constraints: const BoxConstraints(),
                  padding: const EdgeInsets.all(6),
                ),
                const SizedBox(width: 4),
                IconButton(
                  icon: const Icon(Icons.delete_sweep, size: 18),
                  tooltip: 'Clear all failed items',
                  onPressed: _clearAll,
                  constraints: const BoxConstraints(),
                  padding: const EdgeInsets.all(6),
                  style: IconButton.styleFrom(foregroundColor: Colors.red),
                ),
              ],
              IconButton(
                icon: const Icon(Icons.refresh, size: 18),
                tooltip: 'Refresh',
                onPressed: _isLoading ? null : _loadFailedItems,
                constraints: const BoxConstraints(),
                padding: const EdgeInsets.all(6),
              ),
            ],
          ),
        ),

        // Content
        if (_isLoading)
          const Center(
            child: Padding(
              padding: EdgeInsets.all(16),
              child: CircularProgressIndicator(),
            ),
          )
        else if (_failedItems.isEmpty)
          const Padding(
            padding: EdgeInsets.symmetric(vertical: 24),
            child: Center(
              child: Column(
                children: [
                  Icon(Icons.check_circle_outline,
                      size: 40, color: Colors.green),
                  SizedBox(height: 8),
                  Text(
                    'No failed items',
                    style: TextStyle(color: Colors.grey, fontSize: 14),
                  ),
                  SizedBox(height: 4),
                  Text(
                    'All sync items processed successfully',
                    style: TextStyle(color: Colors.grey, fontSize: 12),
                  ),
                ],
              ),
            ),
          )
        else
          ..._failedItems.map((item) => _FailedItemTile(
                item: item,
                onRetry: () => _retryItem(item['id'] as int),
                onDismiss: () => _dismissItem(item['id'] as int),
              )),
      ],
    );
  }
}

/// Tile for a single failed sync queue item.
class _FailedItemTile extends StatelessWidget {
  final Map<String, dynamic> item;
  final VoidCallback onRetry;
  final VoidCallback onDismiss;

  const _FailedItemTile({
    required this.item,
    required this.onRetry,
    required this.onDismiss,
  });

  @override
  Widget build(BuildContext context) {
    final createdDate = item['created_at'] as String? ?? '';
    final errorMsg = item['error_message'] as String? ?? 'Unknown error';
    final tableName = item['table_name'] as String? ?? '?';
    final recordId = item['record_id'] as String? ?? '?';
    final action = item['action'] as String? ?? '?';

    // Parse payload for preview
    String payloadPreview = '';
    try {
      final payloadStr = item['payload'] as String? ?? '';
      if (payloadStr.isNotEmpty && payloadStr.startsWith('{')) {
        final payload = jsonDecode(payloadStr) as Map<String, dynamic>;
        final previewKeys = ['name', 'barcode', 'product_name', 'total', 'price', 'quantity'];
        for (final k in previewKeys) {
          if (payload.containsKey(k)) {
            payloadPreview = '${payload[k]}';
            break;
          }
        }
        if (payloadPreview.length > 40) {
          payloadPreview = '${payloadPreview.substring(0, 40)}...';
        }
      }
    } catch (_) {}

    final theme = Theme.of(context);

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      color: Colors.red.shade50,
      child: ListTile(
        contentPadding: const EdgeInsets.fromLTRB(12, 4, 4, 4),
        leading: CircleAvatar(
          radius: 16,
          backgroundColor: Colors.red.shade100,
          child: Icon(
            action == 'delete' ? Icons.delete_outline : Icons.cloud_off,
            size: 18,
            color: Colors.red.shade700,
          ),
        ),
        title: Row(
          children: [
            Flexible(
              child: Text(
                '$tableName/$recordId',
                style: const TextStyle(
                    fontSize: 13, fontWeight: FontWeight.w600),
                overflow: TextOverflow.ellipsis,
              ),
            ),
            const SizedBox(width: 4),
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
              decoration: BoxDecoration(
                color: Colors.red.shade200,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                action,
                style: TextStyle(
                  fontSize: 9,
                  fontWeight: FontWeight.bold,
                  color: Colors.red.shade900,
                ),
              ),
            ),
          ],
        ),
        subtitle: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            if (payloadPreview.isNotEmpty)
              Text(
                payloadPreview,
                style: TextStyle(fontSize: 11, color: Colors.grey.shade600),
                overflow: TextOverflow.ellipsis,
              ),
            const SizedBox(height: 2),
            Row(
              children: [
                Icon(Icons.error_outline, size: 11, color: Colors.red.shade400),
                const SizedBox(width: 3),
                Expanded(
                  child: Text(
                    errorMsg,
                    style: TextStyle(
                      fontSize: 11,
                      color: Colors.red.shade600,
                      fontStyle: FontStyle.italic,
                    ),
                    overflow: TextOverflow.ellipsis,
                    maxLines: 2,
                  ),
                ),
              ],
            ),
            const SizedBox(height: 2),
            Text(
              createdDate,
              style: TextStyle(fontSize: 10, color: Colors.grey.shade400),
            ),
          ],
        ),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Retry button
            IconButton(
              icon: Icon(Icons.replay, size: 20, color: Colors.orange.shade700),
              tooltip: 'Retry sync',
              onPressed: onRetry,
            ),
            // Dismiss button
            IconButton(
              icon: Icon(Icons.close, size: 18, color: Colors.grey.shade500),
              tooltip: 'Dismiss',
              onPressed: onDismiss,
            ),
          ],
        ),
      ),
    );
  }
}

/// Standalone full-screen DLQ page (can be navigated to directly).
class DeadLetterQueueScreen extends StatelessWidget {
  const DeadLetterQueueScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Dead Letter Queue'),
      ),
      body: const SingleChildScrollView(
        child: DeadLetterQueuePanel(),
      ),
    );
  }
}