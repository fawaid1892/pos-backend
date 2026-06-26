package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pos-multi-branch/backend/internal/database"
	"pos-multi-branch/backend/internal/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// ─── Auth ───

func FindUserByUsername(ctx context.Context, username string) (*model.User, error) {
	u := &model.User{}
	err := database.Pool.QueryRow(ctx,
		`SELECT id, username, password, full_name, role, branch_id, created_at, updated_at
		 FROM users WHERE username = $1`, username,
	).Scan(&u.ID, &u.Username, &u.Password, &u.FullName, &u.Role, &u.BranchID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func FindUserByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	u := &model.User{}
	err := database.Pool.QueryRow(ctx,
		`SELECT id, username, password, full_name, role, branch_id, created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(&u.ID, &u.Username, &u.Password, &u.FullName, &u.Role, &u.BranchID, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func VerifyPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}

func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func ListUsers(ctx context.Context, branchID *uuid.UUID) ([]model.User, error) {
	query := `SELECT id, username, full_name, role, branch_id, is_active, created_at, updated_at
			  FROM users WHERE deleted_at IS NULL`
	args := []interface{}{}
	argIdx := 1

	if branchID != nil {
		query += fmt.Sprintf(" AND (branch_id = $%d OR role = 'owner')", argIdx)
		args = append(args, *branchID)
		argIdx++
	}
	query += " ORDER BY created_at DESC"

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.BranchID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func CreateUser(ctx context.Context, req model.CreateUserRequest) (*model.User, error) {
	hashed, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	u := &model.User{}
	err = database.Pool.QueryRow(ctx,
		`INSERT INTO users (username, password, full_name, role, branch_id, is_active)
		 VALUES ($1,$2,$3,$4,$5,true)
		 RETURNING id, username, full_name, role, branch_id, is_active, created_at, updated_at`,
		req.Username, hashed, req.FullName, req.Role, req.BranchID,
	).Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &u.BranchID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func UpdateUser(ctx context.Context, id uuid.UUID, req model.UpdateUserRequest) (*model.User, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if req.FullName != nil {
		setClauses = append(setClauses, fmt.Sprintf("full_name = $%d", argIdx))
		args = append(args, *req.FullName)
		argIdx++
	}
	if req.Role != nil {
		setClauses = append(setClauses, fmt.Sprintf("role = $%d", argIdx))
		args = append(args, *req.Role)
		argIdx++
	}
	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argIdx))
		args = append(args, *req.IsActive)
		argIdx++
	}
	if req.BranchID != nil {
		if *req.BranchID == nil {
			setClauses = append(setClauses, fmt.Sprintf("branch_id = $%d", argIdx))
			args = append(args, nil)
			argIdx++
		} else {
			setClauses = append(setClauses, fmt.Sprintf("branch_id = $%d", argIdx))
			args = append(args, **req.BranchID)
			argIdx++
		}
	}

	if len(setClauses) == 0 {
		return nil, errors.New("no fields to update")
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))
	args = append(args, id)

	query := fmt.Sprintf(`UPDATE users SET %s
		WHERE id = $%d AND deleted_at IS NULL
		RETURNING id, username, full_name, role, branch_id, is_active, created_at, updated_at`,
		strings.Join(setClauses, ", "), argIdx)

	u := &model.User{}
	err := database.Pool.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.Username, &u.FullName, &u.Role, &u.BranchID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return u, nil
}

func SoftDeleteUser(ctx context.Context, id uuid.UUID) error {
	tag, err := database.Pool.Exec(ctx,
		`UPDATE users SET deleted_at=$1, is_active=false WHERE id=$2 AND deleted_at IS NULL`,
		time.Now(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("user not found")
	}
	return nil
}

// ─── Branch ───

func ListBranches(ctx context.Context) ([]model.Branch, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT id, name, address, phone, tax_rate, is_active, created_at, updated_at, deleted_at
		 FROM branches WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []model.Branch
	for rows.Next() {
		var b model.Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.TaxRate, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, nil
}

func GetBranchByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	b := &model.Branch{}
	err := database.Pool.QueryRow(ctx,
		`SELECT id, name, address, phone, tax_rate, is_active, created_at, updated_at, deleted_at
		 FROM branches WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.TaxRate, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func CreateBranch(ctx context.Context, req model.CreateBranchRequest) (*model.Branch, error) {
	b := &model.Branch{}
	err := database.Pool.QueryRow(ctx,
		`INSERT INTO branches (name, address, phone, tax_rate) VALUES ($1,$2,$3,$4)
		 RETURNING id, name, address, phone, tax_rate, is_active, created_at, updated_at, deleted_at`,
		req.Name, req.Address, req.Phone, 0,
	).Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.TaxRate, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func UpdateBranch(ctx context.Context, id uuid.UUID, req model.UpdateBranchRequest) (*model.Branch, error) {
	b := &model.Branch{}
	err := database.Pool.QueryRow(ctx,
		`UPDATE branches SET name=$1, address=$2, phone=$3, updated_at=NOW()
		 WHERE id=$4 AND deleted_at IS NULL
		 RETURNING id, name, address, phone, tax_rate, is_active, created_at, updated_at, deleted_at`,
		req.Name, req.Address, req.Phone, id,
	).Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.TaxRate, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return b, nil
}

func SoftDeleteBranch(ctx context.Context, id uuid.UUID) error {
	tag, err := database.Pool.Exec(ctx,
		`UPDATE branches SET deleted_at=$1, is_active=false WHERE id=$2 AND deleted_at IS NULL`,
		time.Now(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("branch not found")
	}
	return nil
}

// ─── Category ───

func ListCategories(ctx context.Context) ([]model.Category, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT id, name, created_at, updated_at FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func CreateCategory(ctx context.Context, name string) (*model.Category, error) {
	c := &model.Category{}
	err := database.Pool.QueryRow(ctx,
		`INSERT INTO categories (name) VALUES ($1) RETURNING id, name, created_at, updated_at`,
		name,
	).Scan(&c.ID, &c.Name, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

// ─── Product ───

type ListProductsParams struct {
	Query   string
	Barcode string
	Limit   int
	Offset  int
}

func ListProducts(ctx context.Context, p ListProductsParams) ([]model.Product, error) {
	where := "p.deleted_at IS NULL"
	args := []interface{}{}
	argIdx := 1

	if p.Query != "" {
		where += fmt.Sprintf(" AND LOWER(p.name) LIKE LOWER($%d)", argIdx)
		args = append(args, "%"+p.Query+"%")
		argIdx++
	}
	if p.Barcode != "" {
		where += fmt.Sprintf(" AND p.barcode = $%d", argIdx)
		args = append(args, p.Barcode)
		argIdx++
	}

	if p.Limit <= 0 || p.Limit > 100 {
		p.Limit = 20
	}

	query := fmt.Sprintf(`
		SELECT p.id, p.category_id, COALESCE(c.name,'') as category_name,
		       p.name, p.barcode, p.price, p.stock,
		       p.created_at, p.updated_at, p.deleted_at
		FROM products p
		LEFT JOIN categories c ON c.id = p.category_id
		WHERE %s
		ORDER BY p.name
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, p.Limit, p.Offset)

	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []model.Product
	for rows.Next() {
		var pr model.Product
		if err := rows.Scan(&pr.ID, &pr.CategoryID, &pr.CategoryName, &pr.Name, &pr.Barcode,
			&pr.Price, &pr.Stock, &pr.CreatedAt, &pr.UpdatedAt, &pr.DeletedAt); err != nil {
			return nil, err
		}
		products = append(products, pr)
	}
	return products, nil
}

func GetProductByID(ctx context.Context, id uuid.UUID) (*model.Product, error) {
	p := &model.Product{}
	err := database.Pool.QueryRow(ctx,
		`SELECT p.id, p.category_id, COALESCE(c.name,'') as category_name,
		        p.name, p.barcode, p.price, p.stock,
		        p.created_at, p.updated_at, p.deleted_at
		 FROM products p
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE p.id = $1 AND p.deleted_at IS NULL`, id,
	).Scan(&p.ID, &p.CategoryID, &p.CategoryName, &p.Name, &p.Barcode, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func CreateProduct(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	p := &model.Product{}
	err := database.Pool.QueryRow(ctx,
		`INSERT INTO products (category_id, name, barcode, price, stock)
		 VALUES ($1,$2,$3,$4,$5)
		 RETURNING id, category_id, name, barcode, price, stock, created_at, updated_at, deleted_at`,
		req.CategoryID, req.Name, req.Barcode, req.Price, req.Stock,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Barcode, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func UpdateProduct(ctx context.Context, id uuid.UUID, req model.UpdateProductRequest) (*model.Product, error) {
	p := &model.Product{}
	err := database.Pool.QueryRow(ctx,
		`UPDATE products SET category_id=$1, name=$2, barcode=$3, price=$4, stock=$5, updated_at=NOW()
		 WHERE id=$6 AND deleted_at IS NULL
		 RETURNING id, category_id, name, barcode, price, stock, created_at, updated_at, deleted_at`,
		req.CategoryID, req.Name, req.Barcode, req.Price, req.Stock, id,
	).Scan(&p.ID, &p.CategoryID, &p.Name, &p.Barcode, &p.Price, &p.Stock, &p.CreatedAt, &p.UpdatedAt, &p.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return p, nil
}

func SoftDeleteProduct(ctx context.Context, id uuid.UUID) error {
	tag, err := database.Pool.Exec(ctx,
		`UPDATE products SET deleted_at=$1 WHERE id=$2 AND deleted_at IS NULL`,
		time.Now(), id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("product not found")
	}
	return nil
}

// ─── Transaction ───

func CreateTransaction(ctx context.Context, txData *model.Transaction) error {
	return database.Pool.QueryRow(ctx,
		`INSERT INTO transactions (branch_id, user_id, customer_name, subtotal,
		                           discount_percent, discount_amount, tax_rate, tax_amount, total,
		                           cash_amount, change_amount, payment_method, payment_reference)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		 RETURNING id, created_at`,
		txData.BranchID, txData.UserID, txData.CustomerName,
		txData.Subtotal, txData.DiscountPercent, txData.DiscountAmount,
		txData.TaxRate, txData.TaxAmount,
		txData.Total, txData.CashAmount, txData.ChangeAmount,
		txData.PaymentMethod, txData.PaymentReference,
	).Scan(&txData.ID, &txData.CreatedAt)
}

func InsertTransactionItem(ctx context.Context, item *model.TransactionItem) error {
	return database.Pool.QueryRow(ctx,
		`INSERT INTO transaction_items (transaction_id, product_id, product_name, quantity, price, subtotal)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id`,
		item.TransactionID, item.ProductID, item.ProductName, item.Quantity, item.Price, item.Subtotal,
	).Scan(&item.ID)
}

func DeductProductStock(ctx context.Context, productID uuid.UUID, qty int) error {
	tag, err := database.Pool.Exec(ctx,
		`UPDATE products SET stock = stock - $1, updated_at = NOW()
		 WHERE id = $2 AND deleted_at IS NULL AND stock >= $1`,
		qty, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient stock or product not found")
	}
	return nil
}

func DeductBranchProductStock(ctx context.Context, branchID, productID uuid.UUID, qty float64) error {
	tag, err := database.Pool.Exec(ctx,
		`UPDATE branch_products SET stock_qty = stock_qty - $1, updated_at = NOW()
		 WHERE branch_id = $2 AND product_id = $3 AND stock_qty >= $1`,
		qty, branchID, productID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient stock in this branch or product not found")
	}
	return nil
}

func ListTransactions(ctx context.Context, branchID *uuid.UUID, limit, offset int) ([]model.Transaction, error) {
	where := ""
	args := []interface{}{}
	argIdx := 1

	if branchID != nil {
		where = fmt.Sprintf(" WHERE t.branch_id = $%d", argIdx)
		args = append(args, *branchID)
		argIdx++
	}

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT t.id, t.branch_id, t.user_id, t.customer_name,
		       t.subtotal, t.discount_percent, t.discount_amount,
		       t.tax_rate, t.tax_amount,
		       t.total, t.cash_amount, t.change_amount,
		       COALESCE(t.payment_method, 'cash'), COALESCE(t.payment_reference, ''),
		       t.created_at
		FROM transactions t%s
		ORDER BY t.created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, limit, offset)
	rows, err := database.Pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []model.Transaction
	for rows.Next() {
		var tx model.Transaction
		if err := rows.Scan(&tx.ID, &tx.BranchID, &tx.UserID, &tx.CustomerName,
			&tx.Subtotal, &tx.DiscountPercent, &tx.DiscountAmount,
			&tx.TaxRate, &tx.TaxAmount,
			&tx.Total, &tx.CashAmount, &tx.ChangeAmount,
			&tx.PaymentMethod, &tx.PaymentReference, &tx.CreatedAt); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	return txs, nil
}

func GetTransactionByID(ctx context.Context, id uuid.UUID) (*model.Transaction, error) {
	tx := &model.Transaction{}
	err := database.Pool.QueryRow(ctx,
		`SELECT id, branch_id, user_id, customer_name,
		        subtotal, discount_percent, discount_amount,
		        tax_rate, tax_amount,
		        total, cash_amount, change_amount,
		        COALESCE(payment_method, 'cash'), COALESCE(payment_reference, ''),
		        created_at
		 FROM transactions WHERE id = $1`, id,
	).Scan(&tx.ID, &tx.BranchID, &tx.UserID, &tx.CustomerName,
		&tx.Subtotal, &tx.DiscountPercent, &tx.DiscountAmount,
		&tx.TaxRate, &tx.TaxAmount,
		&tx.Total, &tx.CashAmount, &tx.ChangeAmount,
		&tx.PaymentMethod, &tx.PaymentReference, &tx.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// load items
	rows, err := database.Pool.Query(ctx,
		`SELECT id, transaction_id, product_id, product_name, quantity, price, subtotal
		 FROM transaction_items WHERE transaction_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item model.TransactionItem
		if err := rows.Scan(&item.ID, &item.TransactionID, &item.ProductID,
			&item.ProductName, &item.Quantity, &item.Price, &item.Subtotal); err != nil {
			return nil, err
		}
		tx.Items = append(tx.Items, item)
	}
	return tx, nil
}

// ─── Branch Products ───

func UpsertBranchProduct(ctx context.Context, branchID, productID uuid.UUID, qty float64) error {
	_, err := database.Pool.Exec(ctx,
		`INSERT INTO branch_products (branch_id, product_id, stock_qty)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (branch_id, product_id)
		 DO UPDATE SET stock_qty = branch_products.stock_qty + $3, updated_at = NOW()`,
		branchID, productID, qty)
	return err
}

func SetBranchProductStock(ctx context.Context, branchID, productID uuid.UUID, qty float64) error {
	_, err := database.Pool.Exec(ctx,
		`INSERT INTO branch_products (branch_id, product_id, stock_qty)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (branch_id, product_id)
		 DO UPDATE SET stock_qty = $3, updated_at = NOW()`,
		branchID, productID, qty)
	return err
}

func GetBranchProduct(ctx context.Context, branchID, productID uuid.UUID) (*model.BranchProduct, error) {
	bp := &model.BranchProduct{}
	err := database.Pool.QueryRow(ctx,
		`SELECT bp.branch_id, bp.product_id, bp.stock_qty, bp.created_at, bp.updated_at,
		        p.name, p.barcode, p.price, COALESCE(c.name, '') as category_name
		 FROM branch_products bp
		 JOIN products p ON p.id = bp.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE bp.branch_id = $1 AND bp.product_id = $2`,
		branchID, productID,
	).Scan(&bp.BranchID, &bp.ProductID, &bp.StockQty, &bp.CreatedAt, &bp.UpdatedAt,
		&bp.ProductName, &bp.Barcode, &bp.Price, &bp.CategoryName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return bp, nil
}

func ListBranchProducts(ctx context.Context, branchID uuid.UUID) ([]model.BranchProduct, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT bp.branch_id, bp.product_id, bp.stock_qty, bp.created_at, bp.updated_at,
		        p.name, p.barcode, p.price, COALESCE(c.name, '') as category_name
		 FROM branch_products bp
		 JOIN products p ON p.id = bp.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE bp.branch_id = $1
		 ORDER BY p.name`,
		branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.BranchProduct
	for rows.Next() {
		var bp model.BranchProduct
		if err := rows.Scan(&bp.BranchID, &bp.ProductID, &bp.StockQty, &bp.CreatedAt, &bp.UpdatedAt,
			&bp.ProductName, &bp.Barcode, &bp.Price, &bp.CategoryName); err != nil {
			return nil, err
		}
		items = append(items, bp)
	}
	return items, nil
}

// ─── Stock Mutations ───

func InsertStockMutation(ctx context.Context, m *model.StockMutation) error {
	return database.Pool.QueryRow(ctx,
		`INSERT INTO stock_mutations (branch_id, product_id, type, qty, reference_id, notes)
		 VALUES ($1,$2,$3,$4,$5,$6)
		 RETURNING id, created_at`,
		m.BranchID, m.ProductID, m.Type, m.Qty, m.ReferenceID, m.Notes,
	).Scan(&m.ID, &m.CreatedAt)
}

func ListStockMutations(ctx context.Context, branchID uuid.UUID, limit, offset int) ([]model.StockMutation, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	rows, err := database.Pool.Query(ctx,
		`SELECT sm.id, sm.branch_id, sm.product_id, sm.type, sm.qty,
		        sm.reference_id, sm.notes, sm.created_at,
		        p.name, p.barcode
		 FROM stock_mutations sm
		 JOIN products p ON p.id = sm.product_id
		 WHERE sm.branch_id = $1
		 ORDER BY sm.created_at DESC
		 LIMIT $2 OFFSET $3`,
		branchID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.StockMutation
	for rows.Next() {
		var m model.StockMutation
		if err := rows.Scan(&m.ID, &m.BranchID, &m.ProductID, &m.Type, &m.Qty,
			&m.ReferenceID, &m.Notes, &m.CreatedAt, &m.ProductName, &m.Barcode); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, nil
}

// ─── Stock Transfer (atomic with DB txn) ───

func TransferStock(ctx context.Context, sourceBranchID, targetBranchID, productID uuid.UUID, qty float64, notes string) error {
	tx, err := database.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Decrease source branch stock
	tag, err := tx.Exec(ctx,
		`UPDATE branch_products SET stock_qty = stock_qty - $1, updated_at = NOW()
		 WHERE branch_id = $2 AND product_id = $3 AND stock_qty >= $1`,
		qty, sourceBranchID, productID)
	if err != nil {
		return fmt.Errorf("decrease source: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("insufficient stock in source branch or product not found")
	}

	// Increase target branch stock
	_, err = tx.Exec(ctx,
		`INSERT INTO branch_products (branch_id, product_id, stock_qty)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (branch_id, product_id)
		 DO UPDATE SET stock_qty = branch_products.stock_qty + $3, updated_at = NOW()`,
		targetBranchID, productID, qty)
	if err != nil {
		return fmt.Errorf("increase target: %w", err)
	}

	// Insert transfer_out mutation for source
	_, err = tx.Exec(ctx,
		`INSERT INTO stock_mutations (branch_id, product_id, type, qty, notes)
		 VALUES ($1, $2, 'transfer_out', $3, $4)`,
		sourceBranchID, productID, qty, notes)
	if err != nil {
		return fmt.Errorf("source mutation: %w", err)
	}

	// Insert transfer_in mutation for target
	_, err = tx.Exec(ctx,
		`INSERT INTO stock_mutations (branch_id, product_id, type, qty, notes)
		 VALUES ($1, $2, 'transfer_in', $3, $4)`,
		targetBranchID, productID, qty, notes)
	if err != nil {
		return fmt.Errorf("target mutation: %w", err)
	}

	return tx.Commit(ctx)
}

// ─── Reports ───

func GetSalesReport(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.SalesReportRow, float64, float64, float64, int, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT DATE(created_at)::TEXT as day,
		        COUNT(*)::INT as tx_count,
		        COALESCE(SUM(subtotal), 0) as total_subtotal,
		        COALESCE(SUM(discount_amount), 0) as total_discount,
		        COALESCE(SUM(total), 0) as total_net
		 FROM transactions
		 WHERE branch_id = $1 AND created_at >= $2 AND created_at < $3
		 GROUP BY DATE(created_at)
		 ORDER BY day`,
		branchID, start, end)
	if err != nil {
		return nil, 0, 0, 0, 0, err
	}
	defer rows.Close()

	var report []model.SalesReportRow
	var totalSales, totalDiscount, totalNet float64
	var totalTx int

	for rows.Next() {
		var r model.SalesReportRow
		if err := rows.Scan(&r.Date, &r.TransactionCount, &r.Subtotal, &r.Discount, &r.Total); err != nil {
			return nil, 0, 0, 0, 0, err
		}
		report = append(report, r)
		totalSales += r.Subtotal
		totalDiscount += r.Discount
		totalNet += r.Total
		totalTx += r.TransactionCount
	}
	return report, totalSales, totalDiscount, totalNet, totalTx, nil
}

func GetStockReport(ctx context.Context, branchID uuid.UUID) ([]model.StockReportRow, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT bp.product_id, p.name, p.barcode,
		        COALESCE(c.name, '') as category_name,
		        bp.stock_qty, COALESCE(bp.min_stock, 0),
		        (SELECT MAX(sm.created_at) FROM stock_mutations sm WHERE sm.branch_id = bp.branch_id AND sm.product_id = bp.product_id) as last_mutation
		 FROM branch_products bp
		 JOIN products p ON p.id = bp.product_id
		 LEFT JOIN categories c ON c.id = p.category_id
		 WHERE bp.branch_id = $1
		 ORDER BY p.name`,
		branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.StockReportRow
	for rows.Next() {
		var r model.StockReportRow
		if err := rows.Scan(&r.ProductID, &r.ProductName, &r.Barcode,
			&r.CategoryName, &r.CurrentStock, &r.MinStock, &r.LastMutation); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, nil
}

func GetProfitLossReport(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.ProfitLossRow, model.ProfitLossSummary, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT ti.product_id, p.name,
		        SUM(ti.quantity)::INT as qty_sold,
		        COALESCE(SUM(ti.subtotal), 0) as revenue,
		        COALESCE(SUM(ti.quantity * p.price * 0.7), 0) as cost,
		        COALESCE(SUM(ti.subtotal - ti.quantity * p.price * 0.7), 0) as profit
		 FROM transaction_items ti
		 JOIN transactions t ON t.id = ti.transaction_id
		 JOIN products p ON p.id = ti.product_id
		 WHERE t.branch_id = $1 AND t.created_at >= $2 AND t.created_at < $3
		 GROUP BY ti.product_id, p.name
		 ORDER BY profit DESC`,
		branchID, start, end)
	if err != nil {
		return nil, model.ProfitLossSummary{}, err
	}
	defer rows.Close()

	var items []model.ProfitLossRow
	var summary model.ProfitLossSummary

	for rows.Next() {
		var r model.ProfitLossRow
		if err := rows.Scan(&r.ProductID, &r.ProductName, &r.QtySold, &r.Revenue, &r.Cost, &r.Profit); err != nil {
			return nil, model.ProfitLossSummary{}, err
		}
		items = append(items, r)
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

func GetSalesExportData(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]SalesExportRow, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT t.created_at::TEXT, t.customer_name,
		        (SELECT STRING_AGG(ti.product_name || ' x' || ti.quantity::TEXT, ', ')
		         FROM transaction_items ti WHERE ti.transaction_id = t.id) as items,
		        t.subtotal, t.discount_amount, t.tax_amount, t.total, t.cash_amount, t.change_amount
		 FROM transactions t
		 WHERE t.branch_id = $1 AND t.created_at >= $2 AND t.created_at < $3
		 ORDER BY t.created_at DESC`,
		branchID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var data []SalesExportRow
	for rows.Next() {
		var r SalesExportRow
		if err := rows.Scan(&r.Date, &r.CustomerName, &r.Items, &r.Subtotal, &r.Discount, &r.TaxAmount, &r.Total, &r.Cash, &r.Change); err != nil {
			return nil, err
		}
		data = append(data, r)
	}
	return data, nil
}

// ─── Low Stock ───

func GetLowStockProducts(ctx context.Context, branchID uuid.UUID, threshold float64) ([]model.LowStockItem, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT p.name, b.name, bp.stock_qty, COALESCE(bp.min_stock, 0)
		 FROM branch_products bp
		 JOIN products p ON p.id = bp.product_id
		 JOIN branches b ON b.id = bp.branch_id
		 WHERE bp.branch_id = $1 AND bp.stock_qty <= $2
		 ORDER BY bp.stock_qty ASC`,
		branchID, threshold)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.LowStockItem
	for rows.Next() {
		var r model.LowStockItem
		if err := rows.Scan(&r.ProductName, &r.BranchName, &r.StockQty, &r.MinStock); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, nil
}

func GetLowStockProductsByMinStock(ctx context.Context, branchID uuid.UUID) ([]model.LowStockItem, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT p.name, b.name, bp.stock_qty, COALESCE(bp.min_stock, 0)
		 FROM branch_products bp
		 JOIN products p ON p.id = bp.product_id
		 JOIN branches b ON b.id = bp.branch_id
		 WHERE bp.branch_id = $1 AND bp.min_stock > 0 AND bp.stock_qty <= bp.min_stock
		 ORDER BY bp.stock_qty ASC`,
		branchID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []model.LowStockItem
	for rows.Next() {
		var r model.LowStockItem
		if err := rows.Scan(&r.ProductName, &r.BranchName, &r.StockQty, &r.MinStock); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, nil
}

// ─── PDF Sales Data ───

func GetSalesPDFData(ctx context.Context, branchID uuid.UUID, start, end time.Time) ([]model.SalesPDFRow, error) {
	pgRows, err := database.Pool.Query(ctx,
		`SELECT t.created_at::TEXT as tanggal,
		        ti.product_name, ti.quantity, ti.price,
		        ti.subtotal, t.tax_amount, t.total
		 FROM transactions t
		 JOIN transaction_items ti ON ti.transaction_id = t.id
		 WHERE t.branch_id = $1 AND t.created_at >= $2 AND t.created_at < $3
		 ORDER BY t.created_at DESC, ti.product_name`,
		branchID, start, end)
	if err != nil {
		return nil, err
	}
	defer pgRows.Close()

	var items []model.SalesPDFRow
	for pgRows.Next() {
		var r model.SalesPDFRow
		if err := pgRows.Scan(&r.Date, &r.ProductName, &r.Quantity, &r.Price, &r.Subtotal, &r.TaxAmount, &r.Total); err != nil {
			return nil, err
		}
		items = append(items, r)
	}
	return items, nil
}
