package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"pos-multi-branch/backend/internal/model"
	"pos-multi-branch/backend/internal/repository"

	"github.com/google/uuid"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// List godoc
// @Summary      List products
// @Description  Get paginated list of products with search and filter options
// @Tags         Products
// @Produce      json
// @Param q           query     string  false  "Search by name"
// @Param barcode     query     string  false  "Filter by barcode"
// @Param category_id query     string  false  "Filter by category UUID"
// @Param sort_by     query     string  false  "Sort field (name, price, stock, created_at)"
// @Param sort_order  query     string  false  "Sort direction (asc, desc)"
// @Param min_stock   query     int     false  "Minimum stock threshold"
// @Param limit       query     int     false  "Items per page (max 100)"
// @Param offset      query     int     false  "Number of items to skip"
// @Success      200  {array}   model.Product
// @Security     BearerAuth
// @Router       /products [get]
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	limit, _ := strconv.Atoi(q.Get("limit"))
	offset, _ := strconv.Atoi(q.Get("offset"))
	minStock, _ := strconv.Atoi(q.Get("min_stock"))

	var categoryID *uuid.UUID
	if cid := q.Get("category_id"); cid != "" {
		if parsed, err := uuid.Parse(cid); err == nil {
			categoryID = &parsed
		}
	}

	products, err := repository.ListProducts(r.Context(), repository.ListProductsParams{
		Query:      q.Get("q"),
		Barcode:    q.Get("barcode"),
		CategoryID: categoryID,
		SortBy:     q.Get("sort_by"),
		SortOrder:  q.Get("sort_order"),
		MinStock:   minStock,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if products == nil {
		products = []model.Product{}
	}
	writeJSON(w, http.StatusOK, products)
}

// GetByID godoc
// @Summary      Get product by ID
// @Description  Get a single product by its UUID
// @Tags         Products
// @Produce      json
// @Param id   path      string  true  "Product UUID"
// @Success      200  {object}  model.Product
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /products/{id} [get]
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	product, err := repository.GetProductByID(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if product == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	writeJSON(w, http.StatusOK, product)
}

// Create godoc
// @Summary      Create a product
// @Description  Create a new product
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param request body model.CreateProductRequest true "Product data"
// @Success      201  {object}  model.Product
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /products [post]
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req model.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if req.Name == "" || req.Barcode == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name and barcode are required"})
		return
	}
	product, err := repository.CreateProduct(r.Context(), req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, product)
}

// Update godoc
// @Summary      Update a product
// @Description  Update product details by UUID
// @Tags         Products
// @Accept       json
// @Produce      json
// @Param id       path      string                    true  "Product UUID"
// @Param request  body      model.UpdateProductRequest true  "Updated product data"
// @Success      200  {object}  model.Product
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /products/{id} [put]
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	var req model.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	product, err := repository.UpdateProduct(r.Context(), id, req)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if product == nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "product not found"})
		return
	}
	writeJSON(w, http.StatusOK, product)
}

// Delete godoc
// @Summary      Soft-delete a product
// @Description  Soft-delete a product by UUID
// @Tags         Products
// @Produce      json
// @Param id   path      string  true  "Product UUID"
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  map[string]string
// @Failure      404  {object}  map[string]string
// @Security     BearerAuth
// @Router       /products/{id} [delete]
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid id"})
		return
	}
	if err := repository.SoftDeleteProduct(r.Context(), id); err != nil {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"message": "deleted"})
}

// ─── Category Sub-handler ───

// ListCategories godoc
// @Summary      List categories
// @Description  Get all product categories
// @Tags         Categories
// @Produce      json
// @Success      200  {array}   model.Category
// @Security     BearerAuth
// @Router       /categories [get]
func (h *ProductHandler) ListCategories(w http.ResponseWriter, r *http.Request) {
	cats, err := repository.ListCategories(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	if cats == nil {
		cats = []model.Category{}
	}
	writeJSON(w, http.StatusOK, cats)
}

// CreateCategory godoc
// @Summary      Create a category
// @Description  Create a new product category
// @Tags         Categories
// @Accept       json
// @Produce      json
// @Param request body object{name=string} true "Category name"
// @Success      201  {object}  model.Category
// @Failure      400  {object}  map[string]string
// @Security     BearerAuth
// @Router       /categories [post]
func (h *ProductHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid request body"})
		return
	}
	if body.Name == "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "name is required"})
		return
	}
	cat, err := repository.CreateCategory(r.Context(), body.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, cat)
}
