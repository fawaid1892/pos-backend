import 'dart:async';
import 'package:flutter/foundation.dart';
import '../services/sync_service.dart';
import '../services/connectivity_service.dart';

/// Provider that syncs SyncService + ConnectivityService state for UI.
class SyncProvider extends ChangeNotifier {
  final SyncService _syncService = SyncService();
  final ConnectivityService _connectivityService = ConnectivityService();

  bool _isSyncing = false;
  bool _isOnline = true;
  int _pendingCount = 0;
  int _conflictCount = 0;
  SyncResult? _lastSyncResult;
  StreamSubscription? _syncSubscription;

  SyncService get syncService => _syncService;
  ConnectivityService get connectivityService => _connectivityService;
  bool get isSyncing => _isSyncing;
  bool get isOnline => _isOnline;
  int get pendingCount => _pendingCount;
  int get conflictCount => _conflictCount;
  SyncResult? get lastSyncResult => _lastSyncResult;

  /// Initialize the provider.
  Future<void> init() async {
    // Listen to SyncService changes
    _syncService.addListener(_onSyncServiceChanged);

    // Start connectivity monitoring
    _connectivityService.startMonitoring(
      onReconnectedCallback: _onReconnected,
    );

    // Initialize the sync service
    await _syncService.init();

    // Sync initial state
    _isOnline = _connectivityService.isOnline;
    _pendingCount = _syncService.pendingCount;
    _conflictCount = _syncService.conflictCount;

    notifyListeners();
  }

  void _onSyncServiceChanged() {
    _isSyncing = _syncService.status == SyncStatus.syncing;
    _pendingCount = _syncService.pendingCount;
    _conflictCount = _syncService.conflictCount;
    _lastSyncResult = _syncService.lastSyncResult;
    notifyListeners();
  }

  void _onReconnected() {
    _isOnline = true;
    notifyListeners();

    // Auto-trigger sync when connection is restored
    triggerSync();
  }

  /// Trigger a full sync cycle.
  Future<SyncResult> triggerSync() async {
    _isSyncing = true;
    notifyListeners();

    final result = await _syncService.syncAll();

    _isSyncing = _syncService.status == SyncStatus.syncing;
    _pendingCount = _syncService.pendingCount;
    _conflictCount = _syncService.conflictCount;
    _lastSyncResult = result;
    notifyListeners();

    return result;
  }

  /// Check connectivity manually.
  Future<void> checkConnectivity() async {
    _isOnline = await _connectivityService.checkConnectivity();
    notifyListeners();
  }

  /// Refresh pending counts from the sync service.
  Future<void> refreshCounts() async {
    _pendingCount = _syncService.pendingCount;
    _conflictCount = _syncService.conflictCount;
    notifyListeners();
  }

  @override
  void dispose() {
    _syncService.removeListener(_onSyncServiceChanged);
    _connectivityService.dispose();
    super.dispose();
  }
}
