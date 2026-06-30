package repository

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/idgen"
	"pos-multi-branch/backend/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ─── Auth ───

func FindUserByUsername(username string) (*model.User, error) {
	u := &model.User{}
	err := database.DB.Where("username = ?", username).First(u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func FindUserByID(id int64) (*model.User, error) {
	u := &model.User{}
	err := database.DB.Where("id = ?", id).First(u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func VerifyPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

// ─── Branch ───

func ListBranches() ([]model.Branch, error) {
	var branches []model.Branch
	err := database.DB.Where("deleted_at IS NULL").Order("name").Find(&branches).Error
	return branches, err
}

func GetBranchByID(id int64) (*model.Branch, error) {
	b := &model.Branch{}
	err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func CreateBranch(req model.CreateBranchRequest) (*model.Branch, error) {
	b := &model.Branch{
		Name:         req.Name,
		Code:         req.Code,
		Address:      req.Address,
		Phone:        req.Phone,
		Province:     req.Province,
		ProvinceCode: req.ProvinceCode,
		City:         req.City,
		CityCode:     req.CityCode,
	}
	b.ID = idgen.Generate()
	err := database.DB.Create(b).Error
	if err != nil {
		return nil, err
	}
	return b, nil
}

func UpdateBranch(id int64, req model.UpdateBranchRequest) (*model.Branch, error) {
	b := &model.Branch{}
	err := database.DB.Model(b).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"name":          req.Name,
		"code":          req.Code,
		"address":       req.Address,
		"phone":         req.Phone,
		"province":      req.Province,
		"province_code": req.ProvinceCode,
		"city":          req.City,
		"city_code":     req.CityCode,
	}).First(b).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func SoftDeleteBranch(id int64) error {
	// Check if any users still reference this branch
	var count int64
	err := database.DB.Model(&model.User{}).Where("branch_id = ?", id).Limit(1).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete branch: there are users still assigned to this branch")
	}

	result := database.DB.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.Branch{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("branch not found")
	}
	return nil
}

// ─── Branch User Assignment ───

func ListUsersByBranch(branchID int64) ([]model.User, error) {
	var users []model.User
	err := database.DB.Where("branch_id = ? AND deleted_at IS NULL", branchID).Find(&users).Error
	return users, err
}

func AssignUserToBranch(userID int64, branchID *int64) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).UpdateColumn("branch_id", branchID).Error
}

// ─── Category ───

func ListCategories() ([]model.Category, error) {
	var cats []model.Category
	err := database.DB.Order("name").Find(&cats).Error
	return cats, err
}

func CreateCategory(name string) (*model.Category, error) {
	c := &model.Category{Name: name}
	c.ID = idgen.Generate()
	err := database.DB.Create(c).Error
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ─── Product ───

type ListProductsParams struct {
	Query      string
	Barcode    string
	CategoryID string
	SortBy     string
	SortOrder  string
	MinStock   *int
	Limit      int
	Offset     int
}

func CheckBarcodeExists(barcode string) (bool, error) {
	if barcode == "" {
		return false, nil
	}
	var count int64
	err := database.DB.Model(&model.Product{}).Where("barcode = ? AND deleted_at IS NULL", barcode).Limit(1).Count(&count).Error
	return count > 0, err
}

func CheckCategoryExists(id int64) (bool, error) {
	var count int64
	err := database.DB.Model(&model.Category{}).Where("id = ?", id).Limit(1).Count(&count).Error
	return count > 0, err
}

func ListProducts(p ListProductsParams) ([]model.Product, error) {
	query := database.DB.Table("products p").
		Select(`p.id, p.category_id, COALESCE(c.name,'') as category_name,
		        p.name, p.code, p.barcode, p.unit, p.price, COALESCE(p.cost_price, 0), p.stock,
		        p.created_at, p.updated_at, p.deleted_at`).
		Joins("LEFT JOIN categories c ON c.id = p.category_id").
		Where("p.deleted_at IS NULL")

	if p.Query != "" {
		query = query.Where("LOWER(p.name) LIKE LOWER(?)", "%"+p.Query+"%")
	}
	if p.Barcode != "" {
		query = query.Where("p.barcode = ?", p.Barcode)
	}
	if p.CategoryID != "" {
		query = query.Where("p.category_id = ?", p.CategoryID)
	}
	if p.MinStock != nil {
		query = query.Where("p.stock >= ?", *p.MinStock)
	}
	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}

	// Safe sort whitelist
	sortWhitelist := map[string]bool{
		"name":       true,
		"price":      true,
		"stock":      true,
		"created_at": true,
	}
	orderBy := "p.name"
	if sortWhitelist[p.SortBy] {
		orderBy = "p." + p.SortBy
	}
	orderDir := "ASC"
	if p.SortOrder == "desc" {
		orderDir = "DESC"
	}
	query = query.Order(orderBy + " " + orderDir).Limit(p.Limit).Offset(p.Offset)

	var products []model.Product
	err := query.Scan(&products).Error
	return products, err
}

func GetProductByID(id int64) (*model.Product, error) {
	p := &model.Product{}
	err := database.DB.Table("products p").
		Select(`p.id, p.category_id, COALESCE(c.name,'') as category_name,
		        p.name, p.code, p.barcode, p.unit, p.price, COALESCE(p.cost_price, 0), p.stock,
		        p.created_at, p.updated_at, p.deleted_at`).
		Joins("LEFT JOIN categories c ON c.id = p.category_id").
		Where("p.id = ? AND p.deleted_at IS NULL", id).
		Scan(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func CreateProduct(req model.CreateProductRequest) (*model.Product, error) {
	unit := req.Unit
	if unit == "" {
		unit = "PCS"
	}
	p := &model.Product{
		CategoryID: req.CategoryID,
		Name:       req.Name,
		Code:       req.Code,
		Barcode:    req.Barcode,
		Unit:       unit,
		Price:      req.Price,
		CostPrice:  req.CostPrice,
		Stock:      req.Stock,
	}
	p.ID = idgen.Generate()
	err := database.DB.Create(p).Error
	if err != nil {
		return nil, err
	}
	return p, nil
}

func UpdateProduct(id int64, req model.UpdateProductRequest) (*model.Product, error) {
	p := &model.Product{}
	err := database.DB.Model(p).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"category_id": req.CategoryID,
		"name":        req.Name,
		"code":        req.Code,
		"barcode":     req.Barcode,
		"unit":        req.Unit,
		"price":       req.Price,
		"cost_price":  req.CostPrice,
		"stock":       req.Stock,
	}).First(p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func SoftDeleteProduct(id int64) error {
	// Check if any transactions still reference this product
	var count int64
	err := database.DB.Model(&model.TransactionItem{}).Where("product_id = ?", id).Limit(1).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete product: it has transaction history")
	}

	result := database.DB.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.Product{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("product not found")
	}
	return nil
}

// ─── Transaction ───

func CreateTransaction(txData *model.Transaction) error {
	return database.DB.Create(txData).Error
}

func InsertTransactionItem(item *model.TransactionItem) error {
	return database.DB.Create(item).Error
}

func DeductProductStock(productID int64, qty int) error {
	result := database.DB.Model(&model.Product{}).
		Where("id = ? AND deleted_at IS NULL AND stock >= ?", productID, qty).
		UpdateColumn("stock", gorm.Expr("stock - ?", qty))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}

func DeductBranchProductStock(branchID, productID int64, qty float64) error {
	result := database.DB.Model(&model.BranchProduct{}).
		Where("branch_id = ? AND product_id = ? AND stock_qty >= ?", branchID, productID, qty).
		UpdateColumn("stock_qty", gorm.Expr("stock_qty - ?", qty))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient stock in this branch or product not found")
	}
	return nil
}

func ListTransactions(branchID *int64, limit, offset int) ([]model.Transaction, error) {
	query := database.DB.Model(&model.Transaction{})
	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	var txs []model.Transaction
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&txs).Error
	return txs, err
}

func GetTransactionByID(id int64) (*model.Transaction, error) {
	tx := &model.Transaction{}
	err := database.DB.Preload("Items").Where("id = ?", id).First(tx).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return tx, nil
}

// ─── Branch Products ───

func UpsertBranchProduct(branchID, productID int64, qty float64) error {
	return database.DB.Exec(
		`INSERT INTO branch_products (branch_id, product_id, stock_qty)
		 VALUES (?, ?, ?)
		 ON CONFLICT (branch_id, product_id)
		 DO UPDATE SET stock_qty = branch_products.stock_qty + ?, updated_at = NOW()`,
		branchID, productID, qty, qty,
	).Error
}

func SetBranchProductStock(branchID, productID int64, qty float64) error {
	return database.DB.Exec(
		`INSERT INTO branch_products (branch_id, product_id, stock_qty)
		 VALUES (?, ?, ?)
		 ON CONFLICT (branch_id, product_id)
		 DO UPDATE SET stock_qty = ?, updated_at = NOW()`,
		branchID, productID, qty, qty,
	).Error
}

func GetBranchProduct(branchID, productID int64) (*model.BranchProduct, error) {
	bp := &model.BranchProduct{}
	err := database.DB.Table("branch_products bp").
		Select(`bp.branch_id, bp.product_id, bp.stock_qty, bp.created_at, bp.updated_at,
		        p.name, p.barcode, p.price, COALESCE(c.name, '') as category_name`).
		Joins("JOIN products p ON p.id = bp.product_id").
		Joins("LEFT JOIN categories c ON c.id = p.category_id").
		Where("bp.branch_id = ? AND bp.product_id = ?", branchID, productID).
		Scan(bp).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return bp, nil
}

func ListBranchProducts(branchID int64) ([]model.BranchProduct, error) {
	var items []model.BranchProduct
	err := database.DB.Table("branch_products bp").
		Select(`bp.branch_id, bp.product_id, bp.stock_qty, bp.created_at, bp.updated_at,
		        p.name, p.barcode, p.price, COALESCE(c.name, '') as category_name`).
		Joins("JOIN products p ON p.id = bp.product_id").
		Joins("LEFT JOIN categories c ON c.id = p.category_id").
		Where("bp.branch_id = ?", branchID).
		Order("p.name").
		Scan(&items).Error
	return items, err
}

// ─── Stock Mutations ───

func InsertStockMutation(m *model.StockMutation) error {
	m.ID = idgen.Generate()
	return database.DB.Create(m).Error
}

func ListStockMutations(branchID int64, limit, offset int) ([]model.StockMutation, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	var items []model.StockMutation
	err := database.DB.Table("stock_mutations sm").
		Select(`sm.id, sm.branch_id, sm.product_id, sm.type, sm.qty,
		        sm.reference_id, sm.notes, sm.created_at,
		        p.name, p.barcode`).
		Joins("JOIN products p ON p.id = sm.product_id").
		Where("sm.branch_id = ?", branchID).
		Order("sm.created_at DESC").
		Limit(limit).Offset(offset).
		Scan(&items).Error
	return items, err
}

// ─── Stock Transfer (atomic with DB txn) ───

func TransferStock(sourceBranchID, targetBranchID, productID int64, qty float64, notes string) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin tx: %w", tx.Error)
	}
	defer tx.Rollback()

	// Decrease source branch stock
	result := tx.Model(&model.BranchProduct{}).
		Where("branch_id = ? AND product_id = ? AND stock_qty >= ?", sourceBranchID, productID, qty).
		UpdateColumn("stock_qty", gorm.Expr("stock_qty - ?", qty))
	if result.Error != nil {
		return fmt.Errorf("decrease source: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("insufficient stock in source branch or product not found")
	}

	// Increase target branch stock
	err := tx.Exec(
		`INSERT INTO branch_products (branch_id, product_id, stock_qty)
		 VALUES (?, ?, ?)
		 ON CONFLICT (branch_id, product_id)
		 DO UPDATE SET stock_qty = branch_products.stock_qty + ?, updated_at = NOW()`,
		targetBranchID, productID, qty, qty,
	).Error
	if err != nil {
		return fmt.Errorf("increase target: %w", err)
	}

	// Insert transfer_out mutation for source
	err = tx.Exec(
		`INSERT INTO stock_mutations (branch_id, product_id, type, qty, notes)
		 VALUES (?, ?, 'transfer_out', ?, ?)`,
		sourceBranchID, productID, qty, notes,
	).Error
	if err != nil {
		return fmt.Errorf("source mutation: %w", err)
	}

	// Insert transfer_in mutation for target
	err = tx.Exec(
		`INSERT INTO stock_mutations (branch_id, product_id, type, qty, notes)
		 VALUES (?, ?, 'transfer_in', ?, ?)`,
		targetBranchID, productID, qty, notes,
	).Error
	if err != nil {
		return fmt.Errorf("target mutation: %w", err)
	}

	return tx.Commit().Error
}

// ─── Reports ───

func GetSalesReport(branchID int64, start, end time.Time) ([]model.SalesReportRow, float64, float64, float64, int, error) {
	type salesRow struct {
		Day       string
		TxCount   int
		Subtotal  float64
		Discount  float64
		TotalNet  float64
	}
	var rows []salesRow
	err := database.DB.Table("transactions").
		Select(`DATE(created_at)::TEXT as day,
		        COUNT(*)::INT as tx_count,
		        COALESCE(SUM(subtotal), 0) as subtotal,
		        COALESCE(SUM(discount_amount), 0) as discount,
		        COALESCE(SUM(total), 0) as total_net`).
		Where("branch_id = ? AND created_at >= ? AND created_at < ?", branchID, start, end).
		Group("DATE(created_at)").
		Order("day").
		Scan(&rows).Error
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}

	var report []model.SalesReportRow
	var totalSales, totalDiscount, totalNet float64
	var totalTx int
	for _, r := range rows {
		report = append(report, model.SalesReportRow{
			Date:              r.Day,
			TransactionCount:  r.TxCount,
			Subtotal:          r.Subtotal,
			Discount:          r.Discount,
			Total:             r.TotalNet,
		})
		totalSales += r.Subtotal
		totalDiscount += r.Discount
		totalNet += r.TotalNet
		totalTx += r.TxCount
	}
	return report, totalSales, totalDiscount, totalNet, totalTx, nil
}

func GetStockReport(branchID int64) ([]model.StockReportRow, error) {
	var items []model.StockReportRow
	err := database.DB.Table("branch_products bp").
		Select(`bp.product_id, p.name, p.barcode,
		        COALESCE(c.name, '') as category_name,
		        bp.stock_qty, COALESCE(bp.min_stock, 0),
		        (SELECT MAX(sm.created_at) FROM stock_mutations sm WHERE sm.branch_id = bp.branch_id AND sm.product_id = bp.product_id) as last_mutation`).
		Joins("JOIN products p ON p.id = bp.product_id").
		Joins("LEFT JOIN categories c ON c.id = p.category_id").
		Where("bp.branch_id = ?", branchID).
		Order("p.name").
		Scan(&items).Error
	return items, err
}

func GetProfitLossReport(branchID int64, start, end time.Time) ([]model.ProfitLossRow, model.ProfitLossSummary, error) {
	type profitRow struct {
		ProductID   int64
		Name        string
		QtySold     int
		Revenue     float64
		Cost        float64
		Profit      float64
	}
	var rows []profitRow
	err := database.DB.Table("transaction_items ti").
		Select(`ti.product_id, p.name,
		        SUM(ti.quantity)::INT as qty_sold,
		        COALESCE(SUM(ti.subtotal), 0) as revenue,
		        COALESCE(SUM(ti.quantity * COALESCE(NULLIF(p.cost_price, 0), p.price * 0.7)), 0) as cost,
		        COALESCE(SUM(ti.subtotal - ti.quantity * COALESCE(NULLIF(p.cost_price, 0), p.price * 0.7)), 0) as profit`).
		Joins("JOIN transactions t ON t.id = ti.transaction_id").
		Joins("JOIN products p ON p.id = ti.product_id").
		Where("t.branch_id = ? AND t.created_at >= ? AND t.created_at < ?", branchID, start, end).
		Group("ti.product_id, p.name").
		Order("profit DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, model.ProfitLossSummary{}, err
	}

	var items []model.ProfitLossRow
	var summary model.ProfitLossSummary
	for _, r := range rows {
		items = append(items, model.ProfitLossRow{
			ProductID:   r.ProductID,
			ProductName: r.Name,
			QtySold:     r.QtySold,
			Revenue:     r.Revenue,
			Cost:        r.Cost,
			Profit:      r.Profit,
		})
		summary.TotalRevenue += r.Revenue
		summary.TotalCost += r.Cost
		summary.TotalProfit += r.Profit
	}
	return items, summary, nil
}

// ─── Export ───

type SalesExportRow struct {
	Date         string  `json:"date"`
	CustomerName string  `json:"customer_name"`
	Items        string  `json:"items"`
	Subtotal     float64 `json:"subtotal"`
	Discount     float64 `json:"discount"`
	TaxAmount    float64 `json:"tax_amount"`
	Total        float64 `json:"total"`
	Cash         float64 `json:"cash"`
	Change       float64 `json:"change"`
}

func GetSalesExportData(branchID int64, start, end time.Time) ([]SalesExportRow, error) {
	var data []SalesExportRow
	err := database.DB.Table("transactions t").
		Select(`t.created_at::TEXT, t.customer_name,
		        (SELECT STRING_AGG(ti.product_name || ' x' || ti.quantity::TEXT, ', ')
		         FROM transaction_items ti WHERE ti.transaction_id = t.id) as items,
		        t.subtotal, t.discount_amount, t.tax_amount, t.total, t.cash_amount, t.change_amount`).
		Where("t.branch_id = ? AND t.created_at >= ? AND t.created_at < ?", branchID, start, end).
		Order("t.created_at DESC").
		Scan(&data).Error
	return data, err
}

// ─── Low Stock ───

func GetLowStockProducts(branchID int64, threshold float64) ([]model.LowStockItem, error) {
	var items []model.LowStockItem
	err := database.DB.Table("branch_products bp").
		Select("p.name, b.name, bp.stock_qty, COALESCE(bp.min_stock, 0)").
		Joins("JOIN products p ON p.id = bp.product_id").
		Joins("JOIN branches b ON b.id = bp.branch_id").
		Where("bp.branch_id = ? AND bp.stock_qty <= ?", branchID, threshold).
		Order("bp.stock_qty ASC").
		Scan(&items).Error
	return items, err
}

func GetLowStockProductsByMinStock(branchID int64) ([]model.LowStockItem, error) {
	var items []model.LowStockItem
	err := database.DB.Table("branch_products bp").
		Select("p.name, b.name, bp.stock_qty, COALESCE(bp.min_stock, 0)").
		Joins("JOIN products p ON p.id = bp.product_id").
		Joins("JOIN branches b ON b.id = bp.branch_id").
		Where("bp.branch_id = ? AND bp.min_stock > 0 AND bp.stock_qty <= bp.min_stock", branchID).
		Order("bp.stock_qty ASC").
		Scan(&items).Error
	return items, err
}

// ─── PDF Sales Data ───

func GetSalesPDFData(branchID int64, start, end time.Time) ([]model.SalesPDFRow, error) {
	var items []model.SalesPDFRow
	err := database.DB.Table("transactions t").
		Select(`t.created_at::TEXT as date,
		        ti.product_name, ti.quantity, ti.price,
		        ti.subtotal, t.tax_amount, t.total`).
		Joins("JOIN transaction_items ti ON ti.transaction_id = t.id").
		Where("t.branch_id = ? AND t.created_at >= ? AND t.created_at < ?", branchID, start, end).
		Order("t.created_at DESC, ti.product_name").
		Scan(&items).Error
	return items, err
}

// ─── Dashboard ───

func GetDashboardStats(branchID int64) (*model.DashboardStatsResponse, error) {
	resp := &model.DashboardStatsResponse{}

	err := database.DB.Table("transactions").
		Select("COALESCE(SUM(total), 0)").
		Where("created_at >= CURRENT_DATE AND branch_id = ?", branchID).
		Scan(&resp.TodayRevenue).Error
	if err != nil {
		return nil, err
	}

	err = database.DB.Table("transactions").
		Select("COUNT(*)").
		Where("created_at >= CURRENT_DATE AND branch_id = ?", branchID).
		Scan(&resp.TotalTransactions).Error
	if err != nil {
		return nil, err
	}

	err = database.DB.Model(&model.Branch{}).
		Where("deleted_at IS NULL AND is_active = true").
		Select("COUNT(*)").
		Scan(&resp.ActiveBranches).Error
	if err != nil {
		return nil, err
	}

	err = database.DB.Table("branch_products").
		Where("stock_qty > 0 AND stock_qty <= min_stock AND branch_id = ?", branchID).
		Select("COUNT(*)").
		Scan(&resp.LowStockItems).Error
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func GetSalesChartData(branchID int64, start, end time.Time) (*model.SalesChartResponse, error) {
	resp := &model.SalesChartResponse{}
	resp.Period.Start = start.Format("2006-01-02")
	resp.Period.End = end.Format("2006-01-02")

	type chartRow struct {
		Day   string
		Total float64
		Count int
	}
	var rows []chartRow
	err := database.DB.Table("transactions").
		Select(`DATE(created_at)::TEXT as day,
		        COALESCE(SUM(total), 0) as total,
		        COUNT(*)::INT as count`).
		Where("branch_id = ? AND created_at >= ? AND created_at < ?", branchID, start, end).
		Group("DATE(created_at)").
		Order("day").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, r := range rows {
		resp.Rows = append(resp.Rows, model.SalesChartRow{
			Date:  r.Day,
			Total: r.Total,
			Count: r.Count,
		})
	}
	if resp.Rows == nil {
		resp.Rows = []model.SalesChartRow{}
	}

	return resp, nil
}

// ─── Refresh Tokens ───

func GenerateRefreshTokenString() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func CreateRefreshToken(userID int64, token string, expiresAt time.Time) error {
	rt := &model.RefreshToken{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	rt.ID = idgen.Generate()
	return database.DB.Create(rt).Error
}

func FindRefreshToken(token string) (*model.RefreshToken, error) {
	rt := &model.RefreshToken{}
	err := database.DB.Where("token = ? AND revoked_at IS NULL AND expires_at > NOW()", token).First(rt).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return rt, nil
}

func RevokeRefreshToken(token string) error {
	now := time.Now()
	return database.DB.Model(&model.RefreshToken{}).Where("token = ?", token).UpdateColumn("revoked_at", &now).Error
}

func RevokeAllUserRefreshTokens(userID int64) error {
	now := time.Now()
	return database.DB.Model(&model.RefreshToken{}).Where("user_id = ? AND revoked_at IS NULL", userID).UpdateColumn("revoked_at", &now).Error
}

// ─── User Management ───

type ListUsersParams struct {
	Page     int
	Limit    int
	Role     string
	BranchID *int64
}

func ListUsers(p ListUsersParams) ([]model.User, int, error) {
	query := database.DB.Model(&model.User{}).Where("deleted_at IS NULL")

	if p.Role != "" {
		query = query.Where("role = ?", p.Role)
	}
	if p.BranchID != nil {
		query = query.Where("branch_id = ?", *p.BranchID)
	}

	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}
	if p.Page <= 0 {
		p.Page = 1
	}
	offset := (p.Page - 1) * p.Limit

	var total int64
	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	var users []model.User
	err = query.Order("full_name").Limit(p.Limit).Offset(offset).Find(&users).Error
	return users, int(total), err
}

func CreateUser(req model.CreateUserRequest) (*model.User, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	u := &model.User{
		Username: req.Username,
		Password: string(hashed),
		FullName: req.FullName,
		Role:     req.Role,
		BranchID: req.BranchID,
	}
	u.ID = idgen.Generate()
	err = database.DB.Create(u).Error
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUser(id int64, req model.UpdateUserRequest) (*model.User, error) {
	updates := map[string]interface{}{}

	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("hash password: %w", err)
		}
		updates["password"] = string(hashed)
	}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.BranchID != nil {
		updates["branch_id"] = *req.BranchID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if len(updates) == 0 {
		return nil, errors.New("no fields to update")
	}

	u := &model.User{}
	err := database.DB.Model(u).Where("id = ? AND deleted_at IS NULL", id).Updates(updates).First(u).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func SoftDeleteUser(id int64) error {
	result := database.DB.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.User{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ─── RBAC ───

func RoleHasPermission(roleID int64, permissionName string) (bool, error) {
	var count int64
	err := database.DB.Table("role_permissions rp").
		Joins("JOIN permissions p ON p.id = rp.permission_id").
		Where("rp.role_id = ? AND p.name = ?", roleID, permissionName).
		Limit(1).
		Count(&count).Error
	return count > 0, err
}

func ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := database.DB.Where("deleted_at IS NULL").Order("name").Find(&roles).Error
	return roles, err
}

func GetRoleByID(id int64) (*model.Role, error) {
	var role model.Role
	err := database.DB.Where("id = ? AND deleted_at IS NULL", id).First(&role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &role, nil
}

func ListPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := database.DB.Order("group, name").Find(&perms).Error
	return perms, err
}

func GetRolePermissions(roleID int64) ([]model.Permission, error) {
	var perms []model.Permission
	err := database.DB.Table("permissions p").
		Joins("JOIN role_permissions rp ON rp.permission_id = p.id").
		Where("rp.role_id = ?", roleID).
		Order("p.group, p.name").
		Find(&perms).Error
	return perms, err
}

func SetRolePermissions(roleID int64, permissionIDs []int64) error {
	tx := database.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}
	defer tx.Rollback()

	// Delete existing permissions for this role
	if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return err
	}

	// Insert new permissions
	for _, pid := range permissionIDs {
		if err := tx.Create(&model.RolePermission{RoleID: roleID, PermissionID: pid}).Error; err != nil {
			return err
		}
	}

	return tx.Commit().Error
}

func CreateRole(name, description string) (*model.Role, error) {
	role := &model.Role{
		Name:        name,
		Description: description,
		IsSystem:    false,
	}
	role.ID = idgen.Generate()
	err := database.DB.Create(role).Error
	if err != nil {
		return nil, err
	}
	return role, nil
}

func UpdateRole(id int64, name, description string) (*model.Role, error) {
	role := &model.Role{}
	err := database.DB.Model(role).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).First(role).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return role, nil
}

func SoftDeleteRole(id int64) error {
	// Check if any users reference this role
	var count int64
	err := database.DB.Model(&model.User{}).Where("role_id = ?", id).Limit(1).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return errors.New("cannot delete role: users are assigned to this role")
	}

	result := database.DB.Where("id = ? AND deleted_at IS NULL", id).Delete(&model.Role{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errors.New("role not found")
	}
	return nil
}
