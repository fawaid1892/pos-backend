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
    final colorScheme = theme.colorScheme;
    final isDark = theme.brightness == Brightness.dark;
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
                color: isDark
                    ? Colors.orange.shade900
                    : Colors.orange.shade50,
                borderRadius:
                    const BorderRadius.vertical(top: Radius.circular(0)),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.warning_amber_rounded,
                          color: isDark
                              ? Colors.orange.shade200
                              : Colors.orange.shade800,
                          size: 24),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          'Conflict Resolution',
                          style: theme.textTheme.titleMedium?.copyWith(
                            fontWeight: FontWeight.bold,
                            color: isDark
                                ? Colors.orange.shade200
                                : Colors.orange.shade900,
                          ),
                        ),
                      ),
                      IconButton(
                        icon: Icon(Icons.close, size: 20,
                            color: isDark
                                ? Colors.orange.shade200
                                : null),
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
                      color: isDark
                          ? Colors.orange.shade300
                          : Colors.grey.shade700,
                      fontFamily: 'monospace',
                    ),
                  ),
                  const SizedBox(height: 4),
                  Row(
                    children: [
                      _buildBadge(
                        '${_diffs.length} fields',
                        Colors.blueGrey,
                        isDark,
                      ),
                      const SizedBox(width: 6),
                      _buildBadge(
                        '$differCount different',
                        differCount > 0 ? Colors.orange : Colors.green,
                        isDark,
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
                          color: isDark
                              ? Colors.orange.shade300
                              : Colors.grey.shade700,
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
              color: isDark ? colorScheme.surfaceVariant : Colors.grey.shade100,
              child: Row(
                children: [
                  SizedBox(
                    width: 120,
                    child: Text('Field',
                        style: TextStyle(
                            fontWeight: FontWeight.bold, fontSize: 12,
                            color: colorScheme.onSurface)),
                  ),
                  Expanded(
                    child: Row(
                      children: [
                        Icon(Icons.devices, size: 14, color: Colors.blue),
                        const SizedBox(width: 4),
                        Text('Local Data',
                            style: TextStyle(
                                fontWeight: FontWeight.bold, fontSize: 12,
                                color: colorScheme.onSurface)),
                        const Spacer(),
                        Icon(Icons.cloud_download,
                            size: 14, color: Colors.green),
                        const SizedBox(width: 4),
                        Text('Server Data',
                            style: TextStyle(
                                fontWeight: FontWeight.bold, fontSize: 12,
                                color: colorScheme.onSurface)),
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
                      .map((diff) => _FieldRow(diff: diff, isDark: isDark))
                      .toList(),
                ),
              ),
            ),

            // Action buttons
            Container(
              padding: const EdgeInsets.fromLTRB(20, 12, 20, 16),
              decoration: BoxDecoration(
                color: isDark ? colorScheme.surfaceVariant : Colors.grey.shade50,
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

  Widget _buildBadge(String text, MaterialColor color, bool isDark) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: isDark ? color.shade800 : color.shade100,
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 10,
          color: isDark ? color.shade200 : color.shade800,
        ),
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
  final bool isDark;

  const _FieldRow({required this.diff, required this.isDark});

  @override
  Widget build(BuildContext context) {
    final bgColor = diff.isDifferent
        ? (isDark ? Colors.orange.shade900.withOpacity(0.3) : Colors.orange.shade50)
        : null;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 8),
      decoration: BoxDecoration(
        color: bgColor,
        border: Border(
          bottom: BorderSide(
            color: isDark ? Colors.grey.shade700 : Colors.grey.shade200,
            width: 0.5,
          ),
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              diff.key,
              style: TextStyle(
                fontSize: 11,
                fontFamily: 'monospace',
                fontWeight: diff.isDifferent ? FontWeight.bold : null,
                color: diff.isDifferent
                    ? (isDark ? Colors.orange.shade200 : Colors.orange.shade900)
                    : (isDark ? Colors.grey.shade400 : Colors.grey.shade700),
              ),
            ),
          ),
          Expanded(
            child: Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                color: diff.isDifferent
                    ? (isDark
                        ? Colors.orange.shade800.withOpacity(0.4)
                        : Colors.orange.shade100)
                    : Colors.transparent,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      diff.localValue,
                      style: TextStyle(
                        fontSize: 11,
                        fontFamily: 'monospace',
                        color: diff.isDifferent
                            ? Colors.blue.shade200
                            : (isDark ? Colors.white70 : Colors.black87),
                      ),
                    ),
                  ),
                  if (diff.isDifferent)
                    Icon(Icons.arrow_forward, size: 12,
                        color: isDark
                            ? Colors.orange.shade300
                            : Colors.orange.shade600),
                  Expanded(
                    child: Text(
                      diff.serverValue,
                      style: TextStyle(
                        fontSize: 11,
                        fontFamily: 'monospace',
                        color: diff.isDifferent
                            ? Colors.green.shade200
                            : (isDark ? Colors.white70 : Colors.black87),
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
