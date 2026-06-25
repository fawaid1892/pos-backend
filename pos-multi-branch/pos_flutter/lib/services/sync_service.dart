import 'dart:convert';
import 'package:flutter/foundation.dart';
import '../database/local_database.dart';
import '../database/daos/daos.dart';
import 'connectivity_service.dart';

/// Sync status for the entire sync engine.
enum SyncStatus { idle, syncing, success, error }

/// Result of a single sync operation.
class SyncResult {
  final bool success;
  final int pushedCount;
  final int pulledCount;
  final List<String> conflicts;
  final String? error;
  final DateTime completedAt;

  SyncResult({
    required this.success,
    this.pushedCount = 0,
    this.pulledCount = 0,
    this.conflicts = const [],
    this.error,
    DateTime? completedAt,
  }) : completedAt = completedAt ?? DateTime.now();

  String get summary {
    if (success) {
      final parts = <String>[];
      if (pushedCount > 0) parts.add('$pushedCount pushed');
      if (pulledCount > 0) parts.add('$pulledCount pulled');
      if (conflicts.isNotEmpty) parts.add('${conflicts.length} conflicts');
      return parts.isEmpty ? 'No changes' : '${parts.join(', ')}';
    }
    return 'Failed: ${error ?? "Unknown error"}';
  }
}

/// Sync engine service for managing offline-first data synchronization.
///
/// Handles:
/// - Queue management for pending local changes
/// - Push changes to server (POST /api/v1/sync/push)
/// - Pull master data from server (GET /api/v1/sync/pull)
/// - Conflict detection and resolution
class SyncService extends ChangeNotifier {
  static final SyncService _instance = SyncService._internal();
  factory SyncService() => _instance;
  SyncService._internal();

  final LocalDatabase _db = LocalDatabase();
  final SyncQueueDao _syncQueueDao = SyncQueueDao();
  final ConnectivityService _connectivity = ConnectivityService();

  SyncStatus _status = SyncStatus.idle;
  String? _lastSyncError;

  // Counts for pending sync across all tables
  int _pendingCount = 0;
  int _conflictCount = 0;

  // Last sync result for UI display
  SyncResult? _lastSyncResult;

  SyncStatus get status => _status;
  String? get lastSyncError => _lastSyncError;
  int get pendingCount => _pendingCount;
  int get conflictCount => _conflictCount;
  SyncResult? get lastSyncResult => _lastSyncResult;

  /// Initialize the sync service and compute initial pending counts.
  Future<void> init() async {
    await _refreshPendingCounts();
  }

  /// Refresh pending sync counts from all tables.
  Future<void> _refreshPendingCounts() async {
    try {
      int total = 0;
      int conflicts = 0;

      final tables = [
        'users',
        'branches',
        'categories',
        'products',
        'branch_products',
        'transactions',
        'transaction_items',
        'stock_mutations',
      ];

      final db = await _db.database;
      for (final table in tables) {
        final pendingResult = await db.rawQuery(
          'SELECT COUNT(*) as count FROM $table WHERE pending_sync = 1',
        );
        total += (Sqflite.firstIntValue(pendingResult) ?? 0);

        final conflictResult = await db.rawQuery(
          "SELECT COUNT(*) as count FROM $table WHERE sync_status = 'conflict'",
        );
        conflicts += (Sqflite.firstIntValue(conflictResult) ?? 0);
      }

      // Also count sync_queue pending entries
      total += await _syncQueueDao.countPending();

      _pendingCount = total;
      _conflictCount = conflicts;
      notifyListeners();
    } catch (e) {
      debugPrint('Error refreshing pending counts: $e');
    }
  }

  /// Full sync cycle: push pending changes, then pull master data.
  Future<SyncResult> syncAll() async {
    if (_status == SyncStatus.syncing) {
      return SyncResult(success: false, error: 'Sync already in progress');
    }

    _status = SyncStatus.syncing;
    _lastSyncError = null;
    notifyListeners();

    try {
      // Step 1: Push pending local changes
      final pushResult = await _pushPendingChanges();

      // Step 2: Pull master data from server
      final pullResult = await _pullMasterData();

      // Step 3: Refresh counts
      await _refreshPendingCounts();

      final hasConflicts = pushResult.conflicts.isNotEmpty;
      _lastSyncResult = SyncResult(
        success: true,
        pushedCount: pushResult.pushedCount,
        pulledCount: pullResult.pulledCount,
        conflicts: pushResult.conflicts,
      );

      _status = SyncStatus.success;
      if (hasConflicts) {
        _lastSyncError = '${pushResult.conflicts.length} conflict(s) detected';
      }
      notifyListeners();
      return _lastSyncResult!;
    } catch (e) {
      _lastSyncError = e.toString();
      _lastSyncResult = SyncResult(success: false, error: _lastSyncError);
      _status = SyncStatus.error;
      notifyListeners();

      debugPrint('Sync error: $e');
      return _lastSyncResult!;
    }
  }

  /// Push pending changes from local DB to server.
  Future<SyncResult> _pushPendingChanges() async {
    final db = await _db.database;
    final conflicts = <String>[];
    int pushedCount = 0;

    // Collect pending records from each table
    final tables = [
      'transactions',
      'transaction_items',
      'stock_mutations',
    ];

    for (final table in tables) {
      final pendingRecords = await db.query(
        table,
        where: 'pending_sync = 1',
      );

      if (pendingRecords.isEmpty) continue;

      // Build push payload
      final payload = {
        'table': table,
        'records': pendingRecords,
        'device_id': 'flutter_pos_app',
        'timestamp': DateTime.now().toIso8601String(),
      };

      // Simulate HTTP POST /api/v1/sync/push
      final response = await _mockPushSync(payload);

      if (response['status'] == 'success') {
        // Mark records as synced
        for (final record in pendingRecords) {
          await db.update(
            table,
            {
              'pending_sync': 0,
              'synced_at': DateTime.now().toIso8601String(),
              'sync_status': 'synced',
            },
            where: 'id = ?',
            whereArgs: [record['id']],
          );
        }
        pushedCount += pendingRecords.length;
      } else if (response['status'] == 'conflict') {
        // Mark conflicting records
        final conflictIds = List<String>.from(response['conflict_ids'] ?? []);
        for (final id in conflictIds) {
          await db.update(
            table,
            {'sync_status': 'conflict'},
            where: 'id = ?',
            whereArgs: [id],
          );
          conflicts.add('$table/$id');
        }
        // Mark non-conflicted as synced
        for (final record in pendingRecords) {
          if (!conflictIds.contains(record['id'])) {
            await db.update(
              table,
              {
                'pending_sync': 0,
                'synced_at': DateTime.now().toIso8601String(),
                'sync_status': 'synced',
              },
              where: 'id = ?',
              whereArgs: [record['id']],
            );
            pushedCount++;
          }
        }
      }
    }

    return SyncResult(
      success: true,
      pushedCount: pushedCount,
      conflicts: conflicts,
    );
  }

  /// Pull master data from server.
  Future<SyncResult> _pullMasterData() async {
    int pulledCount = 0;

    // Tables to pull master data for
    final pullTables = ['products', 'branches', 'categories', 'users'];

    for (final table in pullTables) {
      // Simulate GET /api/v1/sync/pull?table=$table&since=
      final response = await _mockPullSync(table);

      if (response['status'] == 'success') {
        final records = List<Map<String, dynamic>>.from(response['data']);
        final db = await _db.database;

        for (final record in records) {
          await db.insert(
            table,
            {
              ...record,
              'pending_sync': 0,
              'synced_at': DateTime.now().toIso8601String(),
              'sync_status': 'synced',
            },
            conflictAlgorithm: ConflictAlgorithm.replace,
          );
          pulledCount++;
        }
      }
    }

    return SyncResult(success: true, pulledCount: pulledCount);
  }

  /// Manually resolve a conflict by choosing local or server version.
  Future<bool> resolveConflict({
    required String tableName,
    required String recordId,
    required bool useLocal,
    Map<String, dynamic>? serverData,
  }) async {
    try {
      final db = await _db.database;

      if (useLocal) {
        // Keep local version, mark as pending for re-push
        await db.update(
          tableName,
          {
            'sync_status': 'pending',
            'pending_sync': 1,
          },
          where: 'id = ?',
          whereArgs: [recordId],
        );
      } else if (serverData != null) {
        // Replace with server version
        await db.update(
          tableName,
          {
            ...serverData,
            'pending_sync': 0,
            'synced_at': DateTime.now().toIso8601String(),
            'sync_status': 'synced',
          },
          where: 'id = ?',
          whereArgs: [recordId],
        );
      }

      await _refreshPendingCounts();
      return true;
    } catch (e) {
      debugPrint('Error resolving conflict: $e');
      return false;
    }
  }

  /// Get conflicting records across all tables.
  Future<List<Map<String, dynamic>>> getConflicts() async {
    final db = await _db.database;
    final conflicts = <Map<String, dynamic>>[];

    final tables = [
      'transactions',
      'transaction_items',
      'stock_mutations',
      'branch_products',
    ];

    for (final table in tables) {
      final records = await db.query(
        table,
        where: "sync_status = 'conflict'",
      );
      for (final record in records) {
        conflicts.add({
          'table': table,
          ...record,
        });
      }
    }

    return conflicts;
  }

  // ──────────────────────────────────────────────────────────
  // Mock implementations for simulating server communication
  // ──────────────────────────────────────────────────────────

  Future<Map<String, dynamic>> _mockPushSync(Map<String, dynamic> payload) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 500));

    // Simulate 90% success, 10% conflict rate
    final hasConflict = DateTime.now().millisecondsSinceEpoch % 10 == 0;

    if (hasConflict) {
      final records = List<Map<String, dynamic>>.from(payload['records']);
      final conflictIds = records
          .where((r) => r['id']?.toString().contains('conflict') ?? false)
          .map((r) => r['id'] as String)
          .toList();

      return {
        'status': 'conflict',
        'conflict_ids': conflictIds,
        'message': '${conflictIds.length} conflict(s) detected',
      };
    }

    return {
      'status': 'success',
      'processed': (payload['records'] as List).length,
      'message': 'Sync successful',
    };
  }

  Future<Map<String, dynamic>> _mockPullSync(String table) async {
    // Simulate network delay
    await Future.delayed(const Duration(milliseconds: 400));

    // Return empty data for mock — in production this would return server data
    return {
      'status': 'success',
      'table': table,
      'since': DateTime.now().subtract(const Duration(days: 7)).toIso8601String(),
      'data': [],
      'message': 'Pull successful',
    };
  }

  /// Get all pending sync queue entries for UI display.
  Future<List<Map<String, dynamic>>> getPendingQueueItems() async {
    return await _syncQueueDao.getPending();
  }

  // ──────────────────────────────────────────────────────────
  // Conflict Detail helpers
  // ──────────────────────────────────────────────────────────

  /// Fetch detailed local and server versions of a conflicted record.
  /// Returns a map with 'local' and 'server' keys.
  Future<Map<String, Map<String, dynamic>>> getConflictDetail({
    required String tableName,
    required String recordId,
  }) async {
    final db = await _db.database;

    // Get local version from table
    final localRows = await db.query(
      tableName,
      where: 'id = ?',
      whereArgs: [recordId],
    );
    final localData = localRows.isNotEmpty ? Map<String, dynamic>.from(localRows.first) : <String, dynamic>{};

    // Simulate fetching server version
    // In production this would call GET /api/v1/sync/conflict/:table/:id
    final serverData = await _mockFetchServerVersion(tableName, recordId, localData);

    return {
      'local': localData,
      'server': serverData,
    };
  }

  Future<Map<String, dynamic>> _mockFetchServerVersion(
    String tableName,
    String recordId,
    Map<String, dynamic> localData,
  ) async {
    await Future.delayed(const Duration(milliseconds: 200));

    // Simulate slight server-side differences for demo purposes
    if (localData.isEmpty) return <String, dynamic>{};

    final server = Map<String, dynamic>.from(localData);
    // Tweak a few fields to simulate server having different values
    if (server.containsKey('updated_at')) {
      server['updated_at'] = DateTime.now().toIso8601String();
    }
    if (server.containsKey('name') && localData['id']?.toString().contains('conflict') == true) {
      server['name'] = '${server['name']} (Server v2)';
    }
    if (server.containsKey('price') && localData['price'] is num) {
      // Simulate server having slightly different price
      server['price'] = (localData['price'] as num).toDouble() * 1.05;
    }
    if (server.containsKey('stock') && localData['stock'] is int) {
      server['stock'] = (localData['stock'] as int) + 3;
    }
    // Mark server version as synced
    server['sync_status'] = 'synced';
    server['pending_sync'] = 0;

    return server;
  }

  // ──────────────────────────────────────────────────────────
  // Dead Letter Queue (DLQ) helpers
  // ──────────────────────────────────────────────────────────

  /// Get all failed sync queue items (DLQ).
  Future<List<Map<String, dynamic>>> getDeadLetterQueueItems() async {
    return await _syncQueueDao.getFailed();
  }

  /// Get count of failed sync queue items.
  Future<int> getDeadLetterCount() async {
    return await _syncQueueDao.countFailed();
  }

  /// Retry a single failed queue item by resetting it to pending.
  Future<bool> retryDeadLetterItem(int queueId) async {
    try {
      return await _syncQueueDao.retryFailedItem(queueId);
    } catch (e) {
      debugPrint('Error retrying DLQ item $queueId: $e');
      return false;
    }
  }

  /// Delete a specific failed queue item from the DLQ.
  Future<bool> dismissDeadLetterItem(int queueId) async {
    try {
      return await _syncQueueDao.deleteFailedItem(queueId);
    } catch (e) {
      debugPrint('Error dismissing DLQ item $queueId: $e');
      return false;
    }
  }

  /// Clear all failed items or only those older than [olderThanDays].
  Future<int> clearDeadLetterQueue({int olderThanDays = 0}) async {
    try {
      return await _syncQueueDao.clearResolved(olderThanDays: olderThanDays);
    } catch (e) {
      debugPrint('Error clearing DLQ: $e');
      return 0;
    }
  }
}
