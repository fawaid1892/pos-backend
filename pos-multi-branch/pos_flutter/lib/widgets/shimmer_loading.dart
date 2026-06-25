import 'package:flutter/material.dart';

/// A shimmer / skeleton loading widget for composable loading placeholders.
///
/// Usage:
/// ```dart
/// ShimmerLoading(
///   child: Column(
///     children: [
///       ShimmerBox(width: 200, height: 20),
///       ShimmerBox(width: double.infinity, height: 48),
///     ],
///   ),
/// )
/// ```
class ShimmerLoading extends StatefulWidget {
  final Widget child;

  const ShimmerLoading({super.key, required this.child});

  @override
  State<ShimmerLoading> createState() => _ShimmerLoadingState();
}

class _ShimmerLoadingState extends State<ShimmerLoading>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return ShimmerMask(
          controller: _controller,
          child: child!,
        );
      },
      child: widget.child,
    );
  }
}

class ShimmerMask extends StatelessWidget {
  final Animation<double> controller;
  final Widget child;

  const ShimmerMask({
    super.key,
    required this.controller,
    required this.child,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final baseColor = isDark ? Colors.grey.shade800 : Colors.grey.shade300;
    final highlightColor =
        isDark ? Colors.grey.shade700 : Colors.grey.shade100;

    return ShaderMask(
      blendMode: BlendMode.srcATop,
      shaderCallback: (bounds) {
        return LinearGradient(
          begin: Alignment.centerLeft,
          end: Alignment.centerRight,
          colors: [
            baseColor,
            baseColor,
            highlightColor,
            baseColor,
            baseColor,
          ],
          stops: const [0.0, 0.35, 0.5, 0.65, 1.0],
          transform: _SlidingGradientTransform(
            slidePercent: controller.value,
          ),
        ).createShader(bounds);
      },
      child: child,
    );
  }
}

class _SlidingGradientTransform extends GradientTransform {
  final double slidePercent;

  _SlidingGradientTransform({required this.slidePercent});

  @override
  Matrix4? transform(Rect bounds, {TextDirection? textDirection}) {
    return Matrix4.translationValues(bounds.width * slidePercent, 0, 0);
  }
}

/// A single shimmer placeholder box.
class ShimmerBox extends StatelessWidget {
  final double width;
  final double height;
  final double borderRadius;

  const ShimmerBox({
    super.key,
    required this.width,
    required this.height,
    this.borderRadius = 8,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      width: width == double.infinity ? null : width,
      height: height,
      decoration: BoxDecoration(
        color: isDark ? Colors.grey.shade800 : Colors.grey.shade300,
        borderRadius: BorderRadius.circular(borderRadius),
      ),
    );
  }
}

/// A skeleton list tile for shimmer loading.
class ShimmerListTile extends StatelessWidget {
  final bool dense;

  const ShimmerListTile({super.key, this.dense = false});

  @override
  Widget build(BuildContext context) {
    final h = dense ? 48.0 : 72.0;
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 12, vertical: 4),
      child: SizedBox(
        height: h,
        child: Row(
          children: [
            const ShimmerBox(width: 40, height: 40, borderRadius: 20),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  ShimmerBox(
                    width: double.infinity,
                    height: dense ? 12 : 14,
                  ),
                  const SizedBox(height: 8),
                  ShimmerBox(
                    width: 120,
                    height: dense ? 10 : 12,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 12),
            ShimmerBox(width: 80, height: 32),
          ],
        ),
      ),
    );
  }
}

/// A shimmer loading page with multiple lines for generic screens.
class ShimmerPage extends StatelessWidget {
  final int itemCount;
  final bool dense;

  const ShimmerPage({super.key, this.itemCount = 6, this.dense = false});

  @override
  Widget build(BuildContext context) {
    return ShimmerLoading(
      child: ListView.builder(
        physics: const NeverScrollableScrollPhysics(),
        itemCount: itemCount,
        itemBuilder: (_, __) => ShimmerListTile(dense: dense),
      ),
    );
  }
}

/// A shimmer card with multiple lines for grid-style loading.
class ShimmerCard extends StatelessWidget {
  final double width;
  final double height;

  const ShimmerCard({
    super.key,
    this.width = double.infinity,
    this.height = 120,
  });

  @override
  Widget build(BuildContext context) {
    return ShimmerLoading(
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        child: Container(
          width: width,
          height: height,
          padding: const EdgeInsets.all(12),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(12),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              ShimmerBox(width: 140, height: 14),
              const SizedBox(height: 12),
              ShimmerBox(width: double.infinity, height: 12),
              const SizedBox(height: 8),
              ShimmerBox(width: 80, height: 12),
              const Spacer(),
              ShimmerBox(width: 100, height: 24, borderRadius: 12),
            ],
          ),
        ),
      ),
    );
  }
}
