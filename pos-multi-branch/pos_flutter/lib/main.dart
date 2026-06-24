import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'providers/auth_provider.dart';
import 'providers/cart_provider.dart';
import 'providers/stock_provider.dart';
import 'providers/report_provider.dart';
import 'screens/login_screen.dart';
import 'screens/pos_screen.dart';
import 'screens/checkout_screen.dart';
import 'screens/receipt_screen.dart';
import 'screens/stock_screen.dart';
import 'screens/stock_adjustment_screen.dart';
import 'screens/stock_transfer_screen.dart';
import 'screens/report_screen.dart';
import 'screens/export_screen.dart';
import 'models/stock_adjustment.dart';

void main() {
  runApp(
    MultiProvider(
      providers: [
        ChangeNotifierProvider(create: (_) => AuthProvider()),
        ChangeNotifierProvider(create: (_) => CartProvider()),
        ChangeNotifierProvider(create: (_) => StockProvider()),
        ChangeNotifierProvider(create: (_) => ReportProvider()),
      ],
      child: const PosApp(),
    ),
  );
}

class PosApp extends StatelessWidget {
  const PosApp({super.key});

  @override
  Widget build(BuildContext context) {
    return MaterialApp(
      title: 'POS Multi Branch',
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        colorSchemeSeed: const Color(0xFF1565C0),
        useMaterial3: true,
        brightness: Brightness.light,
        appBarTheme: const AppBarTheme(
          centerTitle: true,
          elevation: 1,
        ),
      ),
      initialRoute: '/login',
      routes: {
        '/login': (context) => const LoginScreen(),
        '/pos': (context) => const PosScreen(),
        '/checkout': (context) => const CheckoutScreen(),
        '/receipt': (context) => const ReceiptScreen(),
        '/stock': (context) => const StockScreen(),
        '/stock-adjustment': (context) {
          final args = ModalRoute.of(context)?.settings.arguments;
          if (args is ProductStock) {
            return StockAdjustmentScreen(initialProduct: args);
          }
          return const StockAdjustmentScreen();
        },
        '/stock-transfer': (context) {
          final args = ModalRoute.of(context)?.settings.arguments;
          if (args is ProductStock) {
            return StockTransferScreen(initialProduct: args);
          }
          return const StockTransferScreen();
        },
        '/reports': (context) => const ReportScreen(),
        '/export-report': (context) => const ExportScreen(),
      },
    );
  }
}
