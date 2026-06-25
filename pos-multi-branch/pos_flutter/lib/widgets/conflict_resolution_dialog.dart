import 'dart:convert';
import 'package:flutter/material.dart';

/// A dialog that displays local vs server data side-by-side for a sync conflict.
///
/// User can inspect differences and choose which version to keep.
class ConflictResolutionDialog extends StatefulWidget {
  final String tableName;
  final String recordId;
  final Map<String, dynamic> localData;
  final Map<String, dynamic> serverData;

  const ConflictResolutionDialog({
    super.key,
    required this.tableName,
    required this.recordId,
    required this.localData,
    required this.serverData,
  });

  @override
  State<ConflictResolutionDialog> createState() =>
      _ConflictResolutionDialogState();
}

class _ConflictResolutionDialogState extends State<ConflictResolutionDialog> {
  bool _showOnlyDifferences = true;
  late List<FieldDiff> _diffs;

  @override
  void initState() {
    super.initState();
    _diffs = _computeDiffs(widget.localData, widget.serverData);
  }

  List<FieldDiff> _computeDiffs(
    Map<String, dynamic> local,
    Map<String, dynamic> server,
  ) {
    final allKeys = <String>{...local.keys, ...server.keys};
    final diffs = <FieldDiff>[];

    // Priority fields to show first
    const priorityFields = [
      'name', 'barcode', 'price', 'stock', 'quantity', 'total',
      'grand_total', 'amount_paid', 'status', 'role', 'email',
      'address', 'phone', 'reason', 'type',
    ];

    final sortedKeys = allKeys.toList()
      ..sort((a, b) {
        final pa = priorityFields.indexOf(a);
        final pb = priorityFields.indexOf(b);
        if (pa != -1 && pb != -1) return pa.compareTo(pb);
        if (pa != -1) return -1;
        if (pb != -1) return 1;
        return a.compareTo(b);
      });

    // Fields to skip (internal sync fields)
    const skipFields = {
      'pending_sync', 'synced_at', 'sync_status',
    };

    for (final key in sortedKeys) {
      if (skipFields.contains(key)) continue;
      final localVal = local.containsKey(key) ? _formatValue(local[key]) : null;
      final serverVal =
          server.containsKey(key) ? _formatValue(server[key]) : null;
      final isDifferent = localVal != serverVal;
      diffs.add(FieldDiff(
        key: key,
        localValue: localVal ?? '—',
        serverValue: serverVal ?? '—',
        isDifferent: isDifferent,
      ));
    }
    return diffs;
  }

  String _formatValue(dynamic value) {
    if (value == null) return '—';
    if (value is String) return value;
    if (value is num) return value.toString();
    if (value is bool) return value ? 'true' : 'false';
    return jsonEncode(value);
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final differCount = _diffs.where((d) => d.isDifferent).length;
    final isDesktop = MediaQuery.of(context).size.width > 600;

    return Dialog(
      insetPadding: const EdgeInsets.all(16),
      child: ConstrainedBox(
        constraints: BoxConstraints(
          maxWidth: isDesktop ? 800 : 500,
          maxHeight: MediaQuery.of(context).size.height * 0.85,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Header
            Container(
              padding: const EdgeInsets.fromLTRB(20, 16, 8, 8),
              decoration: BoxDecoration(
                color: Colors.orange.shade50,
                borderRadius:
                    const BorderRadius.vertical(top: Radius.circular(0)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.warning_amber_rounded,
                          color: Colors.orange.shade800, size: 24),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'Conflict Resolution',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: Colors.orange.shade900,
                          ),
                        ),
                      ),
                      IconButton(
                        icon: const Icon(Icons.close, size: 20),
                        onPressed: () => Navigator.of(context).pop(),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(),
                      ),
                    ],
                  ),
                  const SizedBox(height: 4),
                  Text(
                    '${widget.tableName}/${widget.recordId}',
                    style: TextStyle(
                      fontSize: 12,
                      color: Colors.grey.shade700,
                      fontFamily: 'monospace',
                    ),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      _buildBadge(
                        '${_diffs.length} fields',
                        Colors.blueGrey,
                      ),
                      const SizedBox(width: 6),
                      _buildBadge(
                        '$differCount different',
                        differCount > 0 ? Colors.orange : Colors.green,
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  // Toggle show only differences
                  Row(
                    children: [
                      Text(
                        'Show only differences',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade700,
                        ),
                      ),
                      const SizedBox(width: 4),
                      SizedBox(
                        height: 24,
                        child: Switch.adaptive(
                          value: _showOnlyDifferences,
                          onChanged: (v) =>
                              setState(() => _showOnlyDifferences = v),
                          materialTapTargetSize:
                              MaterialTapTargetSize.shrinkWrap,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),

            // Column headers
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
              color: Colors.grey.shade100,
              child: Row(
                children: [
                  const SizedBox(
                    width: 120,
                    child: Text('Field',
                        style: TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 12)),
                  ),
                  Expanded(
                    child: Row(
                      children: [
                        Icon(Icons.devices, size: 14, color: Colors.blue),
                        const SizedBox(width: 4),
                        const Text('Local Data',
                            style: TextStyle(
                                fontWeight: FontWeight.bold, fontSize: 12)),
                        const Spacer(),
                        Icon(Icons.cloud_download,
                            size: 14, color: Colors.green),
                        const SizedBox(width: 4),
                        const Text('Server Data',
                            style: TextStyle(
                                fontWeight: FontWeight.bold, fontSize: 12)),
                      ],
                    ),
                  ),
                ],
              ),
            ),

            // Scrollable field list
            Flexible(
              child: SingleChildScrollView(
                child: Column(
                  children: (_showOnlyDifferences
                          ? _diffs.where((d) => d.isDifferent)
                          : _diffs)
                      .map((diff) => _FieldRow(diff: diff))
                      .toList(),
                ),
              ),
            ),

            // Action buttons
            Container(
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 16),
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius:
                    const BorderRadius.vertical(bottom: Radius.circular(0)),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  OutlinedButton.icon(
                    icon: const Icon(Icons.replay, size: 18),
                    label: const Text('Cancel'),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                  Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // Use Local
                      _ActionButton(
                        icon: Icons.devices,
                        label: 'Gunakan Lokal',
                        color: Colors.blue,
                        onPressed: () =>
                            Navigator.of(context).pop(<String, dynamic>{
                          'useLocal': true,
                        }),
                      ),
                      const SizedBox(width: 8),
                      // Use Server
                      _ActionButton(
                        icon: Icons.cloud_download,
                        label: 'Gunakan Server',
                        color: Colors.green,
                        onPressed: () =>
                            Navigator.of(context).pop(<String, dynamic>{
                          'useLocal': false,
                        }),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildBadge(String text, MaterialColor color) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.shade100,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        text,
        style: TextStyle(fontSize: 10, color: color.shade800),
      ),
    );
  }
}

/// Data class holding a single field comparison.
class FieldDiff {
  final String key;
  final String localValue;
  final String serverValue;
  final bool isDifferent;

  FieldDiff({
    required this.key,
    required this.localValue,
    required this.serverValue,
    required this.isDifferent,
  });
}

/// A single field row in the diff comparison table.
class _FieldRow extends StatelessWidget {
  final FieldDiff diff;

  const _FieldRow({required this.diff});

  @override
  Widget build(BuildContext context) {
    final bgColor = diff.isDifferent ? Colors.orange.shade50 : null;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      decoration: BoxDecoration(
        color: bgColor,
        border: Border(
          bottom: BorderSide(color: Colors.grey.shade200, width: 0.5),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Field name
          SizedBox(
            width: 120,
            child: Text(
              diff.key,
              style: TextStyle(
                fontSize: 11,
                fontFamily: 'monospace',
                fontWeight: diff.isDifferent ? FontWeight.bold : null,
                color: diff.isDifferent ? Colors.orange.shade900 : Colors.grey.shade700,
              ),
            ),
          ),
          // Values
          Expanded(
            child: Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                color: diff.isDifferent ? Colors.orange.shade100 : Colors.transparent,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  // Local value
                  Expanded(
                    child: Text(
                      diff.localValue,
                      style: TextStyle(
                        fontSize: 11,
                        fontFamily: 'monospace',
                        color: diff.isDifferent ? Colors.blue.shade800 : Colors.black87,
                      ),
                    ),
                  ),
                  if (diff.isDifferent)
                    Icon(Icons.arrow_forward, size: 12, color: Colors.orange.shade600),
                  // Server value
                  Expanded(
                    child: Text(
                      diff.serverValue,
                      style: TextStyle(
                        fontSize: 11,
                        fontFamily: 'monospace',
                        color: diff.isDifferent ? Colors.green.shade800 : Colors.black87,
                      ),
                      textAlign: TextAlign.right,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}

/// Styled action button used in the dialog footer.
class _ActionButton extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onPressed;

  const _ActionButton({
    required this.icon,
    required this.label,
    required this.color,
    required this.onPressed,
  });

  @override
  Widget build(BuildContext context) {
    return ElevatedButton.icon(
      icon: Icon(icon, size: 18),
      label: Text(label),
      style: ElevatedButton.styleFrom(
        backgroundColor: color,
        foregroundColor: Colors.white,
        padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
        textStyle: const TextStyle(fontSize: 13, fontWeight: FontWeight.w600),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      ),
      onPressed: onPressed,
    );
  }
}

/// Static helper to show the conflict resolution dialog.
///
/// Returns `{'useLocal': true/false}` or null if cancelled.
Future<Map<String, dynamic>?> showConflictResolutionDialog({
  required BuildContext context,
  required String tableName,
  required String recordId,
  required Map<String, dynamic> localData,
  required Map<String, dynamic> serverData,
}) {
  return showDialog<Map<String, dynamic>>(
    context: context,
    builder: (context) => ConflictResolutionDialog(
      tableName: tableName,
      recordId: recordId,
      localData: localData,
      serverData: serverData,
    ),
  );
}
