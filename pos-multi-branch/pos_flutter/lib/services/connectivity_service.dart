import 'dart:async';
import 'package:connectivity_plus/connectivity_plus.dart';
import 'package:flutter/foundation.dart';

/// Monitors internet connectivity and streams status across the app.
///
/// Auto-triggers sync when connection is restored from offline state.
class ConnectivityService extends ChangeNotifier {
  final Connectivity _connectivity = Connectivity();
  StreamSubscription<List<ConnectivityResult>>? _subscription;

  bool _isOnline = true;
  bool _wasOffline = false;

  bool get isOnline => _isOnline;
  bool get wasOffline => _wasOffline;

  /// The callback to invoke when connectivity is restored.
  VoidCallback? onReconnected;

  /// Start listening to connectivity changes.
  void startMonitoring({VoidCallback? onReconnectedCallback}) {
    onReconnected = onReconnectedCallback;

    // Check initial connectivity status
    _checkInitialConnectivity();

    // Listen for changes
    _subscription = _connectivity.onConnectivityChanged.listen(_onConnectivityChanged);
  }

  Future<void> _checkInitialConnectivity() async {
    try {
      final results = await _connectivity.checkConnectivity();
      _updateStatus(results);
    } catch (e) {
      debugPrint('Connectivity check error: $e');
      _isOnline = true; // Assume online on error
    }
  }

  void _onConnectivityChanged(List<ConnectivityResult> results) {
    final wasOnline = _isOnline;
    _updateStatus(results);

    // If we just came back online after being offline
    if (!wasOnline && _isOnline) {
      _wasOffline = true;
      debugPrint('Connection restored! Triggering reconnection callback...');
      onReconnected?.call();
      notifyListeners();

      // Reset flag after a short delay
      Future.delayed(const Duration(seconds: 2), () {
        _wasOffline = false;
        notifyListeners();
      });
    }
  }

  void _updateStatus(List<ConnectivityResult> results) {
    // The device is online if any result is not 'none'
    _isOnline = results.any((r) => r != ConnectivityResult.none);
    notifyListeners();
  }

  /// Manually check current connectivity status.
  Future<bool> checkConnectivity() async {
    try {
      final results = await _connectivity.checkConnectivity();
      _updateStatus(results);
      return _isOnline;
    } catch (e) {
      debugPrint('Connectivity check error: $e');
      return true; // Assume online
    }
  }

  /// Trigger reconnection callback manually (e.g., from a retry button).
  void triggerReconnected() {
    _wasOffline = true;
    onReconnected?.call();
    Future.delayed(const Duration(seconds: 2), () {
      _wasOffline = false;
      notifyListeners();
    });
  }

  @override
  void dispose() {
    _subscription?.cancel();
    super.dispose();
  }
}
