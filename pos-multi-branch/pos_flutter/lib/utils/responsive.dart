import 'package:flutter/material.dart';

/// Breakpoints for responsive layout.
class Breakpoints {
  static const double mobile = 600;
  static const double tablet = 900;
  static const double desktop = 1200;
}

/// Responsive helper extension on BuildContext.
extension ResponsiveContext on BuildContext {
  /// Whether the screen width is at least [Breakpoints.tablet].
  bool get isTablet => MediaQuery.of(this).size.shortestSide >= Breakpoints.mobile;

  /// Whether the screen width is at least [Breakpoints.desktop].
  bool get isDesktop =>
      MediaQuery.of(this).size.shortestSide >= Breakpoints.desktop;

  /// Returns the current device type label.
  DeviceType get deviceType {
    final w = MediaQuery.of(this).size.shortestSide;
    if (w >= Breakpoints.desktop) return DeviceType.desktop;
    if (w >= Breakpoints.tablet) return DeviceType.tablet;
    return DeviceType.mobile;
  }
}

enum DeviceType { mobile, tablet, desktop }

/// Responsive builder widget that builds different layouts per breakpoint.
class ResponsiveLayout extends StatelessWidget {
  final Widget mobile;
  final Widget? tablet;
  final Widget? desktop;

  const ResponsiveLayout({
    super.key,
    required this.mobile,
    this.tablet,
    this.desktop,
  });

  @override
  Widget build(BuildContext context) {
    final device = context.deviceType;

    switch (device) {
      case DeviceType.desktop:
        return desktop ?? tablet ?? mobile;
      case DeviceType.tablet:
        return tablet ?? mobile;
      case DeviceType.mobile:
        return mobile;
    }
  }
}

/// Calculates responsive values based on screen width.
class ResponsiveValue<T> {
  final T mobile;
  final T? tablet;
  final T? desktop;

  const ResponsiveValue({
    required this.mobile,
    this.tablet,
    this.desktop,
  });

  T resolve(BuildContext context) {
    final device = context.deviceType;
    switch (device) {
      case DeviceType.desktop:
        return desktop ?? tablet ?? mobile;
      case DeviceType.tablet:
        return tablet ?? mobile;
      case DeviceType.mobile:
        return mobile;
    }
  }
}

/// Responsive padding/margin values.
class ResponsivePadding {
  static EdgeInsets of(BuildContext context) {
    final isTablet = context.isTablet;
    return EdgeInsets.all(isTablet ? 24.0 : 16.0);
  }

  static EdgeInsets horizontal(BuildContext context) {
    final isTablet = context.isTablet;
    return EdgeInsets.symmetric(horizontal: isTablet ? 24.0 : 16.0);
  }

  static double gridSpacing(BuildContext context) {
    return context.isTablet ? 16.0 : 8.0;
  }

  static int gridColumns(BuildContext context) {
    if (context.isDesktop) return 4;
    if (context.isTablet) return 3;
    return 2;
  }
}
