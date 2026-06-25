import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import '../database/local_database.dart';
import '../database/daos/daos.dart';
import 'api_config.dart';
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

  /// HTTP client — can be injected for testing.
  http.Client _httpClient = http.Client();

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

  // ── Testing Support ──

  /// Replace the default HTTP client (e.g. with a mock for testing).
  void setHttpClient(http.Client client) {
    _httpClient = client;
  }

  // ── Initialization ──

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

      total += await _syncQueueDao.countPending();

      _pendingCount = total;
      _conflictCount = conflicts;
      notifyListeners();
    } catch (e) {
      debugPrint('Error refreshing pending counts: $e');
    }
  }

  // ── Full Sync Cycle ──

  /// Full sync cycle: push pending changes, then pull master data.
  Future<SyncResult> syncAll() async {
    if (_status == SyncStatus.syncing) {
      return SyncResult(success: false, error: 'Sync already in progress');
    }

    _status = SyncStatus.syncing;
    _lastSyncError = null;
    notifyListeners();

    try {
      final pushResult = await _pushPendingChanges();
      final pullResult = await _pullMasterData();

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

  // ── Push ──

  /// Push pending changes from local DB to server.
  Future<SyncResult> _pushPendingChanges() async {
    final db = await _db.database;
    final conflicts = <String>[];
    int pushedCount = 0;

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

      final payload = {
        'table': table,
        'records': pendingRecords,
        'device_id': 'flutter_pos_app',
        'timestamp': DateTime.now().toIso8601String(),
      };

      final response = await _httpPushSync(payload);

      if (response['status'] == 'success') {
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

  // ── Pull ──

  /// Pull master data from server.
  Future<SyncResult> _pullMasterData() async {
    int pulledCount = 0;

    final pullTables = ['products', 'branches', 'categories', 'users'];

    for (final table in pullTables) {
      final response = await _httpPullSync(table);

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

  // ── Real HTTP Implementations ──

  /// Push pending data to server via POST /api/v1/sync/push.
  Future<Map<String, dynamic>> _httpPushSync(
      Map<String, dynamic> payload) async {
    try {
      if (!_connectivity.isOnline) {
        return {'status': 'offline', 'message': 'No internet connection'};
      }

      final uri = Uri.parse(ApiConfig.syncPushUrl);
      final response = await _httpClient
          .post(
            uri,
            headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json',
            },
            body: jsonEncode(payload),
          )
          .timeout(ApiConfig.requestTimeout);

      if (response.statusCode == 200 || response.statusCode == 201) {
        final body = jsonDecode(response.body) as Map<String, dynamic>;
        return body;
      } else if (response.statusCode == 409) {
        // Conflict response from server
        final body = jsonDecode(response.body) as Map<String, dynamic>;
        return {
          'status': 'conflict',
          'conflict_ids': List<String>.from(body['conflict_ids'] ?? []),
          'message': body['message'] ?? 'Conflict detected',
        };
      } else {
        return {
          'status': 'error',
          'message':
              'Server returned ${response.statusCode}: ${response.reasonPhrase}',
        };
      }
    } on http.ClientException catch (e) {
      debugPrint('HTTP push client error: $e');
      return {'status': 'error', 'message': 'Connection failed: ${e.message}'};
    } on Exception catch (e) {
      debugPrint('HTTP push error: $e');
      return {'status': 'error', 'message': e.toString()};
    }
  }

  /// Pull master data from server via GET /api/v1/sync/pull?since=.
  Future<Map<String, dynamic>> _httpPullSync(String table) async {
    try {
      if (!_connectivity.isOnline) {
        return {'status': 'offline', 'data': [], 'message': 'No internet connection'};
      }

      // Use last successful sync timestamp; fallback to 7 days ago.
      final db = await _db.database;
      final lastSyncResult = await db.rawQuery(
        'SELECT MAX(synced_at) as last_sync FROM $table WHERE sync_status = ?',
        ['synced'],
      );
      final lastSync = Sqflite.firstIntValue(lastSyncResult) != null
          ? (lastSyncResult.first['last_sync'] as String)
          : DateTime.now().subtract(const Duration(days: 7)).toIso8601String();

      final uri = Uri.parse(ApiConfig.syncPullUrlWithSince(lastSync))
          .replace(queryParameters: {
        'table': table,
        'since': lastSync,
      });

      final response = await _httpClient
          .get(
            uri,
            headers: {
              'Accept': 'application/json',
            },
          )
          .timeout(ApiConfig.requestTimeout);

      if (response.statusCode == 200) {
        final body = jsonDecode(response.body) as Map<String, dynamic>;
        return body;
      } else {
        return {
          'status': 'error',
          'data': [],
          'message':
              'Pull failed: ${response.statusCode} ${response.reasonPhrase}',
        };
      }
    } on http.ClientException catch (e) {
      debugPrint('HTTP pull client error: $e');
      return {'status': 'error', 'data': [], 'message': 'Connection failed: ${e.message}'};
    } on Exception catch (e) {
      debugPrint('HTTP pull error: $e');
      return {'status': 'error', 'data': [], 'message': e.toString()};
    }
  }

  /// Send conflict resolution choice to server via POST /api/v1/sync/resolve.
  Future<bool> _httpResolveConflict({
    required String tableName,
    required String recordId,
    required bool useLocal,
    Map<String, dynamic>? serverData,
  }) async {
    try {
      final uri = Uri.parse(ApiConfig.syncResolveUrl);
      final response = await _httpClient
          .post(
            uri,
            headers: {
              'Content-Type': 'application/json',
              'Accept': 'application/json',
            },
            body: jsonEncode({
              'table': tableName,
              'record_id': recordId,
              'use_local': useLocal,
              'server_data': serverData,
            }),
          )
          .timeout(ApiConfig.requestTimeout);

      return response.statusCode == 200 || response.statusCode == 201;
    } catch (e) {
      debugPrint('HTTP resolve conflict error: $e');
      // If sending the resolution to server fails, still resolve locally.
      return true;
    }
  }

  /// Fetch server version of a conflicted record via GET /api/v1/sync/conflict/:table/:id.
  Future<Map<String, dynamic>?> _httpFetchServerVersion(
    String tableName,
    String recordId,
  ) async {
    try {
      final uri = Uri.parse(ApiConfig.conflictDetailUrl(tableName, recordId));
      final response = await _httpClient
          .get(
            uri,
            headers: {'Accept': 'application/json'},
          )
          .timeout(ApiConfig.requestTimeout);

      if (response.statusCode == 200) {
        final body = jsonDecode(response.body) as Map<String, dynamic>;
        // Mark server version as synced
        body['sync_status'] = 'synced';
        body['pending_sync'] = 0;
        return body;
      }
      return null;
    } catch (e) {
      debugPrint('HTTP fetch server version error: $e');
      return null;
    }
  }

  // ── Conflict Resolution ──

  /// Manually resolve a conflict by choosing local or server version.
  Future<bool> resolveConflict({
    required String tableName,
    required String recordId,
    required bool useLocal,
    Map<String, dynamic>? serverData,
  }) async {
    try {
      // Notify server about the resolution choice
      await _httpResolveConflict(
        tableName: tableName,
        recordId: recordId,
        useLocal: useLocal,
        serverData: serverData,
      );

      final db = await _db.database;

      if (useLocal) {
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

  // ── Conflict Detail Helpers ──

  /// Fetch detailed local and server versions of a conflicted record.
  Future<Map<String, Map<String, dynamic>>> getConflictDetail({
    required String tableName,
    required String recordId,
  }) async {
    final db = await _db.database;

    final localRows = await db.query(
      tableName,
      where: 'id = ?',
      whereArgs: [recordId],
    );
    final localData = localRows.isNotEmpty
        ? Map<String, dynamic>.from(localRows.first)
        : <String, dynamic>{};

    // Fetch server version via real HTTP
    final serverData = await _httpFetchServerVersion(tableName, recordId) ??
        _buildFallbackServerVersion(localData);

    return {
      'local': localData,
      'server': serverData,
    };
  }

  /// Fallback: build a simulated server version when HTTP fetch fails or returns null.
  Map<String, dynamic> _buildFallbackServerVersion(
      Map<String, dynamic> localData) {
    if (localData.isEmpty) return <String, dynamic>{};

    final server = Map<String, dynamic>.from(localData);
    if (server.containsKey('updated_at')) {
      server['updated_at'] = DateTime.now().toIso8601String();
    }
    server['sync_status'] = 'synced';
    server['pending_sync'] = 0;
    return server;
  }

  /// Get all pending sync queue entries for UI display.
  Future<List<Map<String, dynamic>>> getPendingQueueItems() async {
    return await _syncQueueDao.getPending();
  }

  // ── Dead Letter Queue (DLQ) Helpers ──

  Future<List<Map<String, dynamic>>> getDeadLetterQueueItems() async {
    return await _syncQueueDao.getFailed();
  }

  Future<int> getDeadLetterCount() async {
    return await _syncQueueDao.countFailed();
  }

  Future<bool> retryDeadLetterItem(int queueId) async {
    try {
      return await _syncQueueDao.retryFailedItem(queueId);
    } catch (e) {
      debugPrint('Error retrying DLQ item $queueId: $e');
      return false;
    }
  }

  Future<bool> dismissDeadLetterItem(int queueId) async {
    try {
      return await _syncQueueDao.deleteFailedItem(queueId);
    } catch (e) {
      debugPrint('Error dismissing DLQ item $queueId: $e');
      return false;
    }
  }

  Future<int> clearDeadLetterQueue({int olderThanDays = 0}) async {
    try {
      return await _syncQueueDao.clearResolved(olderThanDays: olderThanDays);
    } catch (e) {
      debugPrint('Error clearing DLQ: $e');
      return 0;
    }
  }
}
