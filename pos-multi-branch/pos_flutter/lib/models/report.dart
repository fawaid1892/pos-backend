class SalesReport {
  final String branchId;
  final String branchName;
  final DateTime startDate;
  final DateTime endDate;
  final double totalSales;
  final int totalTransactions;
  final double averageTransactionValue;
  final int totalItemsSold;
  final List<SalesReportItem> items;

  SalesReport({
    required this.branchId,
    required this.branchName,
    required this.startDate,
    required this.endDate,
    required this.totalSales,
    required this.totalTransactions,
    required this.averageTransactionValue,
    required this.totalItemsSold,
    required this.items,
  });
}

class SalesReportItem {
  final String productId;
  final String productName;
  final int quantity;
  final double total;

  SalesReportItem({
    required this.productId,
    required this.productName,
    required this.quantity,
    required this.total,
  });
}

class StockReport {
  final String branchId;
  final String branchName;
  final int totalProducts;
  final int lowStockCount;
  final int outOfStockCount;
  final int totalStockValue;
  final List<ProductStockReport> products;

  StockReport({
    required this.branchId,
    required this.branchName,
    required this.totalProducts,
    required this.lowStockCount,
    required this.outOfStockCount,
    required this.totalStockValue,
    required this.products,
  });
}

class ProductStockReport {
  final String productId;
  final String productName;
  final String category;
  final int currentStock;
  final int minimumStock;
  final double price;
  final double stockValue;

  ProductStockReport({
    required this.productId,
    required this.productName,
    required this.category,
    required this.currentStock,
    this.minimumStock = 5,
    required this.price,
    required this.stockValue,
  });

  bool get isLowStock => currentStock <= minimumStock;
  bool get isOutOfStock => currentStock == 0;
}

class ProfitLossReport {
  final String branchId;
  final String branchName;
  final DateTime startDate;
  final DateTime endDate;
  final double totalRevenue;
  final double totalCost;
  final double grossProfit;
  final double totalExpenses;
  final double netProfit;
  final List<ProfitLossItem> items;

  ProfitLossReport({
    required this.branchId,
    required this.branchName,
    required this.startDate,
    required this.endDate,
    required this.totalRevenue,
    required this.totalCost,
    required this.grossProfit,
    required this.totalExpenses,
    required this.netProfit,
    required this.items,
  });
}

class ProfitLossItem {
  final String name;
  final String category; // 'revenue', 'cost', 'expense'
  final double amount;

  ProfitLossItem({
    required this.name,
    required this.category,
    required this.amount,
  });
}

class ExportOptions {
  final String reportType; // 'sales', 'stock', 'profit_loss'
  final String format; // 'pdf', 'xlsx'
  final DateTime startDate;
  final DateTime endDate;

  ExportOptions({
    required this.reportType,
    required this.format,
    required this.startDate,
    required this.endDate,
  });
}