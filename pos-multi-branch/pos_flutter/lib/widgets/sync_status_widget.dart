import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/sync_provider.dart';
import '../services/sync_service.dart';
import 'conflict_resolution_dialog.dart';
import 'dead_letter_queue_widget.dart';
import '../utils/responsive.dart';

/// Sync status icon with badge showing pending count.
/// Placed in the AppBar to show sync state at a glance.
class SyncStatusIcon extends StatelessWidget {
  const SyncStatusIcon({super.key});

  @override
  Widget build(BuildContext context) {
    final syncProvider = context.watch<SyncProvider>();
    final isOnline = syncProvider.isOnline;
    final isSyncing = syncProvider.isSyncing;
    final pendingCount = syncProvider.pendingCount;
    final conflictCount = syncProvider.conflictCount;

    Widget icon;

    if (isSyncing) {
      icon = const SizedBox(
        width: 24,
        height: 24,
        child: CircularProgressIndicator(strokeWidth: 2),
      );
    } else if (!isOnline) {
      icon = const Icon(Icons.cloud_off, color: Colors.orange);
    } else if (conflictCount > 0) {
      icon = Badge(
        label: Text('$conflictCount', style: const TextStyle(fontSize: 10)),
        child: const Icon(Icons.cloud_sync, color: Colors.red),
      );
    } else if (pendingCount > 0) {
      icon = Badge(
        label: Text('$pendingCount', style: const TextStyle(fontSize: 10)),
        child: const Icon(Icons.cloud_upload_outlined, color: Colors.blueGrey),
      );
    } else {
      icon = const Icon(Icons.cloud_done, color: Colors.green);
    }

    return Padding(
      padding: const EdgeInsets.only(right: 8),
      child: icon,
    );
  }
}

/// Sync status card widget showing detailed sync information.
class SyncStatusCard extends StatelessWidget {
  const SyncStatusCard({super.key});

  @override
  Widget build(BuildContext context) {
    final syncProvider = context.watch<SyncProvider>();
    final isOnline = syncProvider.isOnline;
    final isSyncing = syncProvider.isSyncing;
    final pendingCount = syncProvider.pendingCount;
    final conflictCount = syncProvider.conflictCount;
    final lastResult = syncProvider.lastSyncResult;

    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Card(
      margin: const EdgeInsets.all(12),
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              children: [
                Icon(
                  isOnline ? Icons.wifi : Icons.wifi_off,
                  color: isOnline ? Colors.green : Colors.red,
                ),
                const SizedBox(width: 8),
                Text(
                  isOnline ? 'Online' : 'Offline',
                  style: TextStyle(
                    fontWeight: FontWeight.bold,
                    color: isOnline ? Colors.green : Colors.red,
                  ),
                ),
                const Spacer(),
                if (isSyncing)
                  const SizedBox(
                    width: 16,
                    height: 16,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  ),
              ],
            ),
            Divider(color: colorScheme.outlineVariant),
            _buildRow('Pending sync', '$pendingCount items', colorScheme),
            _buildRow('Conflicts', '$conflictCount', colorScheme),
            if (lastResult != null) ...[
              const SizedBox(height: 4),
              _buildRow('Last sync', _formatDuration(lastResult.completedAt), colorScheme),
              Text(
                lastResult.summary,
                style: TextStyle(
                  fontSize: 12,
                  color: lastResult.success ? Colors.green : colorScheme.error,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildRow(String label, String value, ColorScheme colorScheme) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label,
              style: TextStyle(color: colorScheme.onSurfaceVariant)),
          Text(value,
              style: const TextStyle(fontWeight: FontWeight.w500)),
        ],
      ),
    );
  }

  String _formatDuration(DateTime dateTime) {
    final diff = DateTime.now().difference(dateTime);
    if (diff.inSeconds < 60) return '${diff.inSeconds}s ago';
    if (diff.inMinutes < 60) return '${diff.inMinutes}m ago';
    return '${diff.inHours}h ago';
  }
}

/// Tile showing a conflict with detailed resolution dialog.
class SyncConflictTile extends StatelessWidget {
  final Map<String, dynamic> conflict;
  final VoidCallback onResolved;

  const SyncConflictTile({
    super.key,
    required this.conflict,
    required this.onResolved,
  });

  @override
  Widget build(BuildContext context) {
    final tableName = conflict['table'] as String? ?? '?';
    final recordId = conflict['id'] as String? ?? '?';
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      color: isDark ? Colors.orange.shade900 : Colors.orange.shade50,
      child: ListTile(
        leading: CircleAvatar(
          radius: 16,
          backgroundColor:
              isDark ? Colors.orange.shade800 : Colors.orange.shade100,
          child: Icon(Icons.warning_amber_rounded,
              size: 20, color: Colors.orange.shade700),
        ),
        title: Text(
          '$tableName/$recordId',
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: isDark ? Colors.orange.shade200 : null,
          ),
          overflow: TextOverflow.ellipsis,
        ),
        subtitle: Text(
          'Conflict detected • Tap to compare versions',
          style: TextStyle(
            fontSize: 12,
            color: isDark ? Colors.orange.shade300 : Colors.grey.shade600,
          ),
        ),
        trailing: const Icon(Icons.chevron_right, color: Colors.grey),
        onTap: () => _openResolutionDialog(context),
      ),
    );
  }

  Future<void> _openResolutionDialog(BuildContext context) async {
    final syncService = context.read<SyncProvider>().syncService;
    final tableName = conflict['table'] as String;
    final recordId = conflict['id'] as String;

    // Show loading indicator
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: const Row(
          children: [
            SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            SizedBox(width: 12),
            Text('Loading conflict details...'),
          ],
        ),
        duration: const Duration(seconds: 30),
      ),
    );

    try {
      final detail = await syncService.getConflictDetail(
        tableName: tableName,
        recordId: recordId,
      );

      if (!context.mounted) return;
      ScaffoldMessenger.of(context).hideCurrentSnackBar();

      final result = await showConflictResolutionDialog(
        context: context,
        tableName: tableName,
        recordId: recordId,
        localData: detail['local'] ?? {},
        serverData: detail['server'] ?? {},
      );

      if (result != null && context.mounted) {
        final useLocal = result['useLocal'] as bool;
        final success = await syncService.resolveConflict(
          tableName: tableName,
          recordId: recordId,
          useLocal: useLocal,
        );

        if (context.mounted) {
          ScaffoldMessenger.of(context).showSnackBar(
            SnackBar(
              content: Text(success
                  ? 'Conflict resolved with ${useLocal ? "local" : "server"} version'
                  : 'Failed to resolve conflict'),
              backgroundColor: success ? Colors.green : Colors.red,
            ),
          );
          onResolved();
          context.read<SyncProvider>().refreshCounts();
        }
      }
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).hideCurrentSnackBar();
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text('Error loading conflict details: $e'),
            backgroundColor: Colors.red,
          ),
        );
      }
    }
  }
}

/// Helper to batch-resolve all conflicts with a strategy.
class _BatchResolveDialog extends StatefulWidget {
  final int conflictCount;

  const _BatchResolveDialog({required this.conflictCount});

  @override
  State<_BatchResolveDialog> createState() => _BatchResolveDialogState();
}

class _BatchResolveDialogState extends State<_BatchResolveDialog> {
  bool _useLocal = true;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('Batch Resolve Conflicts'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            'Resolve all ${widget.conflictCount} conflict(s) at once?',
            style: const TextStyle(fontSize: 14),
          ),
          const SizedBox(height: 16),
          SegmentedButton<bool>(
            segments: const [
              ButtonSegment(
                value: true,
                label: Text('Use Local'),
                icon: Icon(Icons.devices),
              ),
              ButtonSegment(
                value: false,
                label: Text('Use Server'),
                icon: Icon(Icons.cloud_download),
              ),
            ],
            selected: {_useLocal},
            onSelectionChanged: (set) =>
                setState(() => _useLocal = set.first),
          ),
          const SizedBox(height: 8),
          Text(
            _useLocal
                ? 'Keep all local versions and re-push'
                : 'Replace all local data with server versions',
            style: TextStyle(
              fontSize: 12,
              color: Theme.of(context).colorScheme.onSurfaceVariant,
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: () =>
              Navigator.pop(context, {'useLocal': _useLocal}),
          style: ElevatedButton.styleFrom(
            backgroundColor: Colors.orange,
            foregroundColor: Colors.white,
          ),
          child: const Text('Resolve All'),
        ),
      ],
    );
  }
}

/// Full sync status screen with integrated conflict resolution and DLQ.
class SyncStatusScreen extends StatefulWidget {
  const SyncStatusScreen({super.key});

  @override
  State<SyncStatusScreen> createState() => _SyncStatusScreenState();
}

class _SyncStatusScreenState extends State<SyncStatusScreen> {
  List<Map<String, dynamic>> _conflicts = [];
  bool _isLoadingConflicts = false;
  int _deadLetterCount = 0;

  @override
  void initState() {
    super.initState();
    _loadConflicts();
    _loadDeadLetterCount();
  }

  Future<void> _loadConflicts() async {
    setState(() => _isLoadingConflicts = true);
    final syncService = context.read<SyncProvider>().syncService;
    final conflicts = await syncService.getConflicts();
    if (mounted) {
      setState(() {
        _conflicts = conflicts;
        _isLoadingConflicts = false;
      });
    }
  }

  Future<void> _loadDeadLetterCount() async {
    final syncService = context.read<SyncProvider>().syncService;
    final count = await syncService.getDeadLetterCount();
    if (mounted) {
      setState(() => _deadLetterCount = count);
    }
  }

  Future<void> _batchResolve() async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) =>
          _BatchResolveDialog(conflictCount: _conflicts.length),
    );
    if (result == null || !mounted) return;

    final useLocal = result['useLocal'] as bool;
    final syncService = context.read<SyncProvider>().syncService;
    int resolved = 0;

    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(
        content: Row(
          children: [
            const SizedBox(
              width: 16,
              height: 16,
              child: CircularProgressIndicator(strokeWidth: 2),
            ),
            const SizedBox(width: 12),
            Text('Resolving ${_conflicts.length} conflict(s)...'),
          ],
        ),
        duration: const Duration(seconds: 30),
      ),
    );

    for (final conflict in _conflicts) {
      final success = await syncService.resolveConflict(
        tableName: conflict['table'] as String,
        recordId: conflict['id'] as String,
        useLocal: useLocal,
      );
      if (success) resolved++;
    }

    if (mounted) {
      ScaffoldMessenger.of(context).hideCurrentSnackBar();
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('Resolved $resolved / ${_conflicts.length} conflicts'),
          backgroundColor: resolved > 0 ? Colors.green : Colors.red,
        ),
      );
      _loadConflicts();
      context.read<SyncProvider>().refreshCounts();
    }
  }

  @override
  Widget build(BuildContext context) {
    final syncProvider = context.watch<SyncProvider>();
    final theme = Theme.of(context);
    final colorScheme = theme.colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Sync Status'),
        actions: [
          PopupMenuButton<String>(
            onSelected: (value) async {
              switch (value) {
                case 'dlq':
                  await _loadDeadLetterCount();
                  break;
                case 'batch_resolve':
                  if (_conflicts.isNotEmpty) _batchResolve();
                  break;
              }
            },
            itemBuilder: (context) => [
              PopupMenuItem(
                value: 'dlq',
                child: Row(
                  children: [
                    const Icon(Icons.error_outline, size: 18),
                    const SizedBox(width: 8),
                    Text('Dead Letter Queue ($_deadLetterCount)'),
                  ],
                ),
              ),
              if (_conflicts.isNotEmpty)
                PopupMenuItem(
                  value: 'batch_resolve',
                  child: Row(
                    children: [
                      const Icon(Icons.done_all, size: 18),
                      const SizedBox(width: 8),
                      Text('Resolve All (${_conflicts.length})'),
                    ],
                  ),
                ),
            ],
          ),
          IconButton(
            icon: syncProvider.isSyncing
                ? const SizedBox(
                    width: 20,
                    height: 20,
                    child: CircularProgressIndicator(strokeWidth: 2),
                  )
                : const Icon(Icons.refresh),
            onPressed: syncProvider.isSyncing
                ? null
                : () async {
                    final result = await syncProvider.triggerSync();
                    if (context.mounted) {
                      ScaffoldMessenger.of(context).showSnackBar(
                        SnackBar(
                          content: Text(result.summary),
                          backgroundColor:
                              result.success ? Colors.green : Colors.red,
                        ),
                      );
                      _loadConflicts();
                      _loadDeadLetterCount();
                    }
                  },
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: () async {
          await syncProvider.triggerSync();
          await _loadConflicts();
          await _loadDeadLetterCount();
        },
        child: ListView(
          children: [
            const SyncStatusCard(),

            // ── Quick Stats Row ──
            Padding(
              padding:
                  const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
              child: Row(
                children: [
                  _StatChip(
                    icon: Icons.cloud_upload,
                    label: '${syncProvider.pendingCount} Pending',
                    color: Colors.blueGrey,
                  ),
                  const SizedBox(width: 8),
                  _StatChip(
                    icon: Icons.warning_amber,
                    label: '${syncProvider.conflictCount} Conflicts',
                    color: Colors.orange,
                  ),
                  const SizedBox(width: 8),
                  _StatChip(
                    icon: Icons.error_outline,
                    label: '$_deadLetterCount Failed',
                    color: Colors.red,
                  ),
                ],
              ),
            ),

            Divider(color: colorScheme.outlineVariant),

            // ── Conflicts Section ──
            Padding(
              padding: const EdgeInsets.fromLTRB(12, 8, 12, 4),
              child: Row(
                children: [
                  Icon(Icons.warning_amber_rounded,
                      size: 18, color: Colors.orange.shade700),
                  const SizedBox(width: 6),
                  Text(
                    'Conflicts',
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: Colors.orange.shade800,
                    ),
                  ),
                  if (!_isLoadingConflicts && _conflicts.isNotEmpty)
                    Container(
                      margin: const EdgeInsets.only(left: 6),
                      padding: const EdgeInsets.symmetric(
                          horizontal: 6, vertical: 1),
                      decoration: BoxDecoration(
                        color: Colors.orange.shade100,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        '${_conflicts.length}',
                        style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                            color: Colors.orange.shade800),
                      ),
                    ),
                  const Spacer(),
                  if (_conflicts.length > 1)
                    TextButton.icon(
                      icon: const Icon(Icons.done_all, size: 16),
                      label: const Text('Resolve All',
                          style: TextStyle(fontSize: 12)),
                      onPressed: _batchResolve,
                    ),
                ],
              ),
            ),

            if (_isLoadingConflicts)
              const Center(
                child: Padding(
                  padding: EdgeInsets.all(16),
                  child: CircularProgressIndicator(),
                ),
              )
            else if (_conflicts.isEmpty)
              Padding(
                padding: const EdgeInsets.all(12),
                child: Text('No conflicts',
                    textAlign: TextAlign.center,
                    style: TextStyle(color: colorScheme.onSurfaceVariant)),
              )
            else
              ..._conflicts.map((conflict) => SyncConflictTile(
                    conflict: conflict,
                    onResolved: () {
                      _loadConflicts();
                      _loadDeadLetterCount();
                    },
                  )),

            Divider(color: colorScheme.outlineVariant),

            // ── Dead Letter Queue Section ──
            const DeadLetterQueuePanel(),

            Divider(color: colorScheme.outlineVariant),

            // ── Pending Sync Queue Section ──
            _buildSectionHeader('Pending Sync Queue', colorScheme),
            _buildPendingList(colorScheme),
          ],
        ),
      ),
    );
  }

  Widget _buildSectionHeader(String title, ColorScheme colorScheme) {
    return Padding(
      padding: const EdgeInsets.fromLTRB(12, 16, 12, 4),
      child: Row(
        children: [
          Icon(Icons.cloud_upload, size: 18, color: Colors.blueGrey),
          const SizedBox(width: 6),
          Text(
            title,
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: colorScheme.onSurface,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPendingList(ColorScheme colorScheme) {
    return FutureBuilder<List<Map<String, dynamic>>>(
      future: context
          .read<SyncProvider>()
          .syncService
          .getPendingQueueItems(),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(
            child: Padding(
              padding: EdgeInsets.all(8),
              child: CircularProgressIndicator(),
            ),
          );
        }

        final items = snapshot.data ?? [];
        if (items.isEmpty) {
          return Padding(
            padding: const EdgeInsets.all(12),
            child: Text('No pending items',
                textAlign: TextAlign.center,
                style:
                    TextStyle(color: colorScheme.onSurfaceVariant)),
          );
        }

        return ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: items.length,
          itemBuilder: (context, index) {
            final item = items[index];
            return ListTile(
              dense: true,
              leading: Icon(
                item['action'] == 'delete'
                    ? Icons.delete_outline
                    : Icons.cloud_upload,
                size: 20,
                color: Colors.blueGrey,
              ),
              title: Text(
                '${item['table_name']}/${item['record_id']}',
                style: const TextStyle(fontSize: 13),
              ),
              subtitle: Text(
                'Action: ${item['action']} • ${item['created_at']}',
                style: const TextStyle(fontSize: 11),
              ),
            );
          },
        );
      },
    );
  }
}

/// A small stat chip for the quick stats row.
class _StatChip extends StatelessWidget {
  final IconData icon;
  final String label;
  final MaterialColor color;
  final VoidCallback? onTap;

  const _StatChip({
    required this.icon,
    required this.label,
    required this.color,
    this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return GestureDetector(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
        decoration: BoxDecoration(
          color: isDark ? color.shade900 : color.shade50,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(color: isDark ? color.shade700 : color.shade200),
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(icon, size: 14, color: isDark ? color.shade200 : color.shade700),
            const SizedBox(width: 4),
            Text(
              label,
              style: TextStyle(
                fontSize: 11,
                fontWeight: FontWeight.w600,
                color: isDark ? color.shade200 : color.shade800,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
