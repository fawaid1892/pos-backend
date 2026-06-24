package repository

import (
	"context"
	"errors"
	"fmt"
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

// ─── Branch ───

func ListBranches(ctx context.Context) ([]model.Branch, error) {
	rows, err := database.Pool.Query(ctx,
		`SELECT id, name, address, phone, is_active, created_at, updated_at, deleted_at
		 FROM branches WHERE deleted_at IS NULL ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var branches []model.Branch
	for rows.Next() {
		var b model.Branch
		if err := rows.Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt); err != nil {
			return nil, err
		}
		branches = append(branches, b)
	}
	return branches, nil
}

func GetBranchByID(ctx context.Context, id uuid.UUID) (*model.Branch, error) {
	b := &model.Branch{}
	err := database.Pool.QueryRow(ctx,
		`SELECT id, name, address, phone, is_active, created_at, updated_at, deleted_at
		 FROM branches WHERE id = $1 AND deleted_at IS NULL`, id,
	).Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)
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
		`INSERT INTO branches (name, address, phone) VALUES ($1,$2,$3)
		 RETURNING id, name, address, phone, is_active, created_at, updated_at, deleted_at`,
		req.Name, req.Address, req.Phone,
	).Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)
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
		 RETURNING id, name, address, phone, is_active, created_at, updated_at, deleted_at`,
		req.Name, req.Address, req.Phone, id,
	).Scan(&b.ID, &b.Name, &b.Address, &b.Phone, &b.IsActive, &b.CreatedAt, &b.UpdatedAt, &b.DeletedAt)
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

type txQueries struct{}

func CreateTransaction(ctx context.Context, txData *model.Transaction) error {
	return database.Pool.QueryRow(ctx,
		`INSERT INTO transactions (branch_id, user_id, customer_name, subtotal,
		                           discount_percent, discount_amount, total,
		                           cash_amount, change_amount)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
		 RETURNING id, created_at`,
		txData.BranchID, txData.UserID, txData.CustomerName,
		txData.Subtotal, txData.DiscountPercent, txData.DiscountAmount,
		txData.Total, txData.CashAmount, txData.ChangeAmount,
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
		       t.total, t.cash_amount, t.change_amount, t.created_at
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
			&tx.Total, &tx.CashAmount, &tx.ChangeAmount, &tx.CreatedAt); err != nil {
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
		        total, cash_amount, change_amount, created_at
		 FROM transactions WHERE id = $1`, id,
	).Scan(&tx.ID, &tx.BranchID, &tx.UserID, &tx.CustomerName,
		&tx.Subtotal, &tx.DiscountPercent, &tx.DiscountAmount,
		&tx.Total, &tx.CashAmount, &tx.ChangeAmount, &tx.CreatedAt)
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
