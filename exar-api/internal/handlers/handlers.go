package handlers

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/armin/expenses/backend/internal/auth"
	"github.com/armin/expenses/backend/internal/database"
	"github.com/armin/expenses/backend/internal/jalali"
	"github.com/armin/expenses/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgconn"
)

type Handler struct {
	db *database.DB
}

func New(db *database.DB) *Handler {
	return &Handler{db: db}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return strings.Contains(err.Error(), "duplicate key")
}

func (h *Handler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, personID, err := auth.Authenticate(req.Username, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, models.LoginResponse{
		Token:    token,
		Username: req.Username,
		PersonID: personID,
	})
}

func (h *Handler) ListPersons(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name FROM persons ORDER BY id`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list persons"})
		return
	}
	defer rows.Close()

	var persons []models.Person
	for rows.Next() {
		var p models.Person
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read persons"})
			return
		}
		persons = append(persons, p)
	}

	c.JSON(http.StatusOK, persons)
}

func (h *Handler) ListShops(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name FROM shops ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list shops"})
		return
	}
	defer rows.Close()

	var shops []models.Shop
	for rows.Next() {
		var s models.Shop
		if err := rows.Scan(&s.ID, &s.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read shops"})
			return
		}
		shops = append(shops, s)
	}

	c.JSON(http.StatusOK, shops)
}

func (h *Handler) CreateShop(c *gin.Context) {
	var req models.CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	var id int
	err := h.db.QueryRow(`INSERT INTO shops (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "shop already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create shop"})
		return
	}

	c.JSON(http.StatusCreated, models.Shop{ID: id, Name: name})
}

func (h *Handler) DeleteShop(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop id"})
		return
	}

	var inUse int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM items WHERE shop_id = $1`, id).Scan(&inUse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check shop usage"})
		return
	}
	if inUse > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "shop is used by items"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM shops WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete shop"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListProducts(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name FROM products ORDER BY name`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list products"})
		return
	}
	defer rows.Close()

	var products []models.Product
	for rows.Next() {
		var p models.Product
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read products"})
			return
		}
		products = append(products, p)
	}

	if products == nil {
		products = []models.Product{}
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var req models.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	var id int
	err := h.db.QueryRow(`INSERT INTO products (name) VALUES ($1) RETURNING id`, name).Scan(&id)
	if err != nil {
		if isUniqueViolation(err) {
			c.JSON(http.StatusConflict, gin.H{"error": "product already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, models.Product{ID: id, Name: name})
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product id"})
		return
	}

	var inUse int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM items WHERE product_id = $1`, id).Scan(&inUse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check product usage"})
		return
	}
	if inUse > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "product is used by expenses"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete product"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) ListItems(c *gin.Context) {
	query := `
		SELECT i.id, i.shop_id, s.name,
		       COALESCE(i.product_id, 0), COALESCE(p.name, ''),
		       COALESCE(i.person_id, 0), COALESCE(per.name, ''),
		       i.amount, i.date
		FROM items i
		JOIN shops s ON s.id = i.shop_id
		LEFT JOIN products p ON p.id = i.product_id
		LEFT JOIN persons per ON per.id = i.person_id
		WHERE 1=1
	`
	args := []any{}
	argNum := 1

	if from := c.Query("from_date"); from != "" {
		query += fmt.Sprintf(` AND i.date >= $%d`, argNum)
		args = append(args, from)
		argNum++
	}
	if to := c.Query("to_date"); to != "" {
		query += fmt.Sprintf(` AND i.date <= $%d`, argNum)
		args = append(args, to)
		argNum++
	}
	if shop := strings.TrimSpace(c.Query("shop")); shop != "" {
		query += fmt.Sprintf(` AND s.name = $%d`, argNum)
		args = append(args, shop)
		argNum++
	}
	if personID := strings.TrimSpace(c.Query("person_id")); personID != "" {
		query += fmt.Sprintf(` AND i.person_id = $%d`, argNum)
		args = append(args, personID)
	}

	query += ` ORDER BY i.date DESC, i.id DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list items"})
		return
	}
	defer rows.Close()

	var items []models.Item
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(
			&item.ID, &item.ShopID, &item.Shop,
			&item.ProductID, &item.Product,
			&item.PersonID, &item.Person,
			&item.Amount, &item.Date,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read items"})
			return
		}
		items = append(items, item)
	}

	if items == nil {
		items = []models.Item{}
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) GetItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	item, err := h.fetchItem(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) fetchItem(id int) (models.Item, error) {
	var item models.Item
	err := h.db.QueryRow(`
		SELECT i.id, i.shop_id, s.name,
		       COALESCE(i.product_id, 0), COALESCE(p.name, ''),
		       COALESCE(i.person_id, 0), COALESCE(per.name, ''),
		       i.amount, i.date
		FROM items i
		JOIN shops s ON s.id = i.shop_id
		LEFT JOIN products p ON p.id = i.product_id
		LEFT JOIN persons per ON per.id = i.person_id
		WHERE i.id = $1`, id,
	).Scan(
		&item.ID, &item.ShopID, &item.Shop,
		&item.ProductID, &item.Product,
		&item.PersonID, &item.Person,
		&item.Amount, &item.Date,
	)
	return item, err
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req models.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shopName := strings.TrimSpace(req.Shop)
	if shopName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "shop is required"})
		return
	}

	productName := strings.TrimSpace(req.Product)
	if productName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "product is required"})
		return
	}

	if req.PersonID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id is required"})
		return
	}

	if req.Amount <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount must be greater than zero"})
		return
	}

	gregorianDate, err := jalali.ToGregorian(req.Date)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shopID, err := h.resolveShopID(shopName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve shop"})
		return
	}

	productID, err := h.resolveProductID(productName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve product"})
		return
	}

	var personExists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM persons WHERE id = $1`, req.PersonID).Scan(&personExists); err != nil || personExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid person_id"})
		return
	}

	var itemID int
	err = h.db.QueryRow(
		`INSERT INTO items (shop_id, product_id, person_id, amount, date) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		shopID, productID, req.PersonID, req.Amount, gregorianDate,
	).Scan(&itemID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}

	item, err := h.fetchItem(itemID)
	if err != nil {
		c.JSON(http.StatusCreated, gin.H{"id": itemID})
		return
	}

	c.JSON(http.StatusCreated, item)
}

func (h *Handler) resolveShopID(name string) (int, error) {
	var shopID int
	err := h.db.QueryRow(`SELECT id FROM shops WHERE name = $1`, name).Scan(&shopID)
	if err == nil {
		return shopID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	err = h.db.QueryRow(`INSERT INTO shops (name) VALUES ($1) RETURNING id`, name).Scan(&shopID)
	if err != nil {
		return 0, err
	}

	return shopID, nil
}

func (h *Handler) resolveProductID(name string) (int, error) {
	var productID int
	err := h.db.QueryRow(`SELECT id FROM products WHERE name = $1`, name).Scan(&productID)
	if err == nil {
		return productID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, err
	}

	err = h.db.QueryRow(`INSERT INTO products (name) VALUES ($1) RETURNING id`, name).Scan(&productID)
	if err != nil {
		return 0, err
	}

	return productID, nil
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM items WHERE id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete item"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
