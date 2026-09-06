package handlers

import (
	"database/sql"
	"errors"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/armin/expenses/backend/internal/auth"
	"github.com/armin/expenses/backend/internal/database"
	"github.com/armin/expenses/backend/internal/jalali"
	"github.com/armin/expenses/backend/internal/models"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	db *database.DB
}

func New(db *database.DB) *Handler {
	return &Handler{db: db}
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

	persons := make([]models.Person, 0)
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
	q := strings.TrimSpace(c.Query("q"))
	if q != "" && len([]rune(q)) < 3 {
		c.JSON(http.StatusOK, []models.Shop{})
		return
	}

	var (
		rows *sql.Rows
		err  error
	)
	if q != "" {
		rows, err = h.db.Query(
			`SELECT id, name FROM shops WHERE name LIKE ? COLLATE NOCASE ORDER BY name`,
			"%"+q+"%",
		)
	} else {
		rows, err = h.db.Query(`SELECT id, name FROM shops ORDER BY name`)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list shops"})
		return
	}
	defer rows.Close()

	shops := make([]models.Shop, 0)
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

	result, err := h.db.Exec(`INSERT INTO shops (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "shop already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create shop"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, models.Shop{ID: int(id), Name: name})
}

func (h *Handler) UpdateShop(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop id"})
		return
	}

	var req models.UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`UPDATE shops SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "shop already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update shop"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "shop not found"})
		return
	}

	c.JSON(http.StatusOK, models.Shop{ID: id, Name: name})
}

func (h *Handler) DeleteShop(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop id"})
		return
	}

	var inUse int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM expenses WHERE shop_id = ?`, id).Scan(&inUse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check shop usage"})
		return
	}
	if inUse > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "shop is used by expenses"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM shops WHERE id = ?`, id)
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

func (h *Handler) ListItems(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q != "" && len([]rune(q)) < 3 {
		c.JSON(http.StatusOK, []models.Item{})
		return
	}

	var (
		rows *sql.Rows
		err  error
	)
	if q != "" {
		rows, err = h.db.Query(
			`SELECT id, name FROM items WHERE name LIKE ? COLLATE NOCASE ORDER BY name`,
			"%"+q+"%",
		)
	} else {
		rows, err = h.db.Query(`SELECT id, name FROM items ORDER BY name`)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list items"})
		return
	}
	defer rows.Close()

	items := make([]models.Item, 0)
	for rows.Next() {
		var item models.Item
		if err := rows.Scan(&item.ID, &item.Name); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read items"})
			return
		}
		items = append(items, item)
	}

	c.JSON(http.StatusOK, items)
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req models.CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`INSERT INTO items (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "item already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create item"})
		return
	}

	id, _ := result.LastInsertId()
	c.JSON(http.StatusCreated, models.Item{ID: int(id), Name: name})
}

func (h *Handler) UpdateItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var req models.UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	result, err := h.db.Exec(`UPDATE items SET name = ? WHERE id = ?`, name, id)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": "item already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update item"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}

	c.JSON(http.StatusOK, models.Item{ID: id, Name: name})
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item id"})
		return
	}

	var inUse int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM expenses WHERE item_id = ?`, id).Scan(&inUse); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check item usage"})
		return
	}
	if inUse > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "item is used by expenses"})
		return
	}

	result, err := h.db.Exec(`DELETE FROM items WHERE id = ?`, id)
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

func (h *Handler) GetStats(c *gin.Context) {
	fromDate := c.Query("from_date")
	toDate := c.Query("to_date")

	where := []string{"1=1"}
	args := make([]any, 0)
	if fromDate != "" {
		where = append(where, "e.date >= ?")
		args = append(args, fromDate)
	}
	if toDate != "" {
		where = append(where, "e.date <= ?")
		args = append(args, toDate)
	}
	whereSQL := strings.Join(where, " AND ")

	query := `
		SELECT e.date, e.person_id, e.amount,
			COALESCE((SELECT share FROM expense_shares WHERE expense_id = e.id AND person_id = 1), 0),
			COALESCE((SELECT share FROM expense_shares WHERE expense_id = e.id AND person_id = 2), 0)
		FROM expenses e
		WHERE ` + whereSQL

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load expenses for stats"})
		return
	}
	defer rows.Close()

	byMonth := make(map[string]*models.MonthStats)
	for rows.Next() {
		var date string
		var personID int
		var amount, arminShareFrac, raminShareFrac float64
		if err := rows.Scan(&date, &personID, &amount, &arminShareFrac, &raminShareFrac); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read expense stats"})
			return
		}
		month, err := jalali.MonthKeyFromGregorian(date)
		if err != nil {
			continue
		}
		row, ok := byMonth[month]
		if !ok {
			row = &models.MonthStats{Month: month}
			byMonth[month] = row
		}
		row.Total += amount
		if personID == 1 {
			row.Armin += amount
		} else if personID == 2 {
			row.Ramin += amount
		}
		row.ArminShare += amount * arminShareFrac
		row.RaminShare += amount * raminShareFrac
	}

	months := make([]models.MonthStats, 0, len(byMonth))
	for _, row := range byMonth {
		months = append(months, *row)
	}
	sort.Slice(months, func(i, j int) bool {
		return months[i].Month > months[j].Month
	})

	c.JSON(http.StatusOK, models.Stats{ByMonth: months})
}

func validateShares(shares []models.ExpenseShareInput) error {
	seen := make(map[int]bool, len(shares))
	var sum float64
	for _, s := range shares {
		if s.PersonID != 1 && s.PersonID != 2 {
			return errors.New("share person_id must be 1 or 2")
		}
		if s.Share < 0 {
			return errors.New("share must be >= 0")
		}
		if seen[s.PersonID] {
			return errors.New("duplicate person in shares")
		}
		seen[s.PersonID] = true
		sum += s.Share
	}

	for _, id := range []int{1, 2} {
		if !seen[id] {
			return errors.New("shares must include every person")
		}
	}

	if math.Abs(sum-1) > 0.001 {
		return errors.New("shares must sum to 1")
	}
	return nil
}

func validateWholeAmount(amount float64) error {
	if amount < 0 {
		return errors.New("amount must be >= 0")
	}
	if amount != math.Trunc(amount) {
		return errors.New("amount must be a whole number")
	}
	return nil
}

func insertExpenseShares(tx *sql.Tx, expenseID int64, shares []models.ExpenseShareInput) error {
	for _, s := range shares {
		_, err := tx.Exec(
			`INSERT INTO expense_shares (expense_id, person_id, share) VALUES (?, ?, ?)`,
			expenseID, s.PersonID, s.Share,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func upsertItem(tx *sql.Tx, name string) (int64, error) {
	var id int64
	err := tx.QueryRow(`SELECT id FROM items WHERE name = ? COLLATE NOCASE`, name).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return 0, err
	}

	result, err := tx.Exec(`INSERT INTO items (name) VALUES (?)`, name)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE") {
			if err := tx.QueryRow(`SELECT id FROM items WHERE name = ? COLLATE NOCASE`, name).Scan(&id); err != nil {
				return 0, err
			}
			return id, nil
		}
		return 0, err
	}
	return result.LastInsertId()
}

func (h *Handler) CheckDuplicateExpense(c *gin.Context) {
	date := strings.TrimSpace(c.Query("date"))
	if date == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date is required"})
		return
	}

	excludeID := strings.TrimSpace(c.Query("exclude_id"))
	itemIDStr := strings.TrimSpace(c.Query("item_id"))
	name := strings.TrimSpace(c.Query("name"))

	var (
		count int
		err   error
	)

	switch {
	case itemIDStr != "":
		itemID, parseErr := strconv.Atoi(itemIDStr)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid item_id"})
			return
		}
		if excludeID != "" {
			err = h.db.QueryRow(
				`SELECT COUNT(*) FROM expenses WHERE item_id = ? AND date = ? AND id != ?`,
				itemID, date, excludeID,
			).Scan(&count)
		} else {
			err = h.db.QueryRow(
				`SELECT COUNT(*) FROM expenses WHERE item_id = ? AND date = ?`,
				itemID, date,
			).Scan(&count)
		}
	case name != "":
		if excludeID != "" {
			err = h.db.QueryRow(
				`SELECT COUNT(*) FROM expenses e
				 JOIN items i ON i.id = e.item_id
				 WHERE i.name = ? COLLATE NOCASE AND e.date = ? AND e.id != ?`,
				name, date, excludeID,
			).Scan(&count)
		} else {
			err = h.db.QueryRow(
				`SELECT COUNT(*) FROM expenses e
				 JOIN items i ON i.id = e.item_id
				 WHERE i.name = ? COLLATE NOCASE AND e.date = ?`,
				name, date,
			).Scan(&count)
		}
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "item_id or name is required"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check duplicate"})
		return
	}

	c.JSON(http.StatusOK, models.DuplicateCheckResponse{Exists: count > 0, Count: count})
}

func (h *Handler) ListExpenses(c *gin.Context) {
	query := `
		SELECT e.id, e.person_id, p.name, e.shop_id, s.name, e.item_id, i.name, e.date, e.amount
		FROM expenses e
		JOIN persons p ON p.id = e.person_id
		JOIN shops s ON s.id = e.shop_id
		JOIN items i ON i.id = e.item_id
		WHERE 1=1
	`
	args := []any{}

	if personID := c.Query("person_id"); personID != "" {
		query += ` AND e.person_id = ?`
		args = append(args, personID)
	}
	if from := c.Query("from_date"); from != "" {
		query += ` AND e.date >= ?`
		args = append(args, from)
	}
	if to := c.Query("to_date"); to != "" {
		query += ` AND e.date <= ?`
		args = append(args, to)
	}

	query += ` ORDER BY e.date DESC, e.id DESC`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list expenses"})
		return
	}
	defer rows.Close()

	expenses := make([]models.Expense, 0)
	var expenseIDs []int
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(
			&e.ID, &e.PersonID, &e.PersonName, &e.ShopID, &e.ShopName,
			&e.ItemID, &e.Name, &e.Date, &e.Amount,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read expenses"})
			return
		}
		e.Shares = []models.ExpenseShare{}
		expenses = append(expenses, e)
		expenseIDs = append(expenseIDs, e.ID)
	}

	if len(expenseIDs) > 0 {
		sharesByExpense, err := h.loadSharesForExpenses(expenseIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load expense shares"})
			return
		}
		for i := range expenses {
			if shares, ok := sharesByExpense[expenses[i].ID]; ok {
				expenses[i].Shares = shares
			}
		}
	}

	c.JSON(http.StatusOK, expenses)
}

func (h *Handler) loadSharesForExpenses(expenseIDs []int) (map[int][]models.ExpenseShare, error) {
	placeholders := make([]string, len(expenseIDs))
	args := make([]any, len(expenseIDs))
	for i, id := range expenseIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `SELECT sh.expense_id, sh.person_id, p.name, sh.share
		FROM expense_shares sh
		JOIN persons p ON p.id = sh.person_id
		WHERE sh.expense_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY sh.person_id`

	rows, err := h.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int][]models.ExpenseShare)
	for rows.Next() {
		var expenseID int
		var share models.ExpenseShare
		if err := rows.Scan(&expenseID, &share.PersonID, &share.PersonName, &share.Share); err != nil {
			return nil, err
		}
		result[expenseID] = append(result[expenseID], share)
	}

	return result, nil
}

func (h *Handler) GetExpense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	exp, err := h.fetchExpense(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get expense"})
		return
	}

	c.JSON(http.StatusOK, exp)
}

func (h *Handler) fetchExpense(id int) (models.Expense, error) {
	var e models.Expense
	err := h.db.QueryRow(`
		SELECT e.id, e.person_id, p.name, e.shop_id, s.name, e.item_id, i.name, e.date, e.amount
		FROM expenses e
		JOIN persons p ON p.id = e.person_id
		JOIN shops s ON s.id = e.shop_id
		JOIN items i ON i.id = e.item_id
		WHERE e.id = ?`, id,
	).Scan(
		&e.ID, &e.PersonID, &e.PersonName, &e.ShopID, &e.ShopName,
		&e.ItemID, &e.Name, &e.Date, &e.Amount,
	)
	if err != nil {
		return e, err
	}

	sharesByExpense, err := h.loadSharesForExpenses([]int{id})
	if err != nil {
		return e, err
	}
	e.Shares = sharesByExpense[id]
	if e.Shares == nil {
		e.Shares = []models.ExpenseShare{}
	}

	return e, nil
}

func (h *Handler) CreateExpenses(c *gin.Context) {
	var req models.CreateExpensesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PersonID != 1 && req.PersonID != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id must be 1 or 2"})
		return
	}

	for _, item := range req.Items {
		if err := validateShares(item.Shares); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if strings.TrimSpace(item.Name) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "item name is required"})
			return
		}
		if err := validateWholeAmount(item.Amount); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	var shopExists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM shops WHERE id = ?`, req.ShopID).Scan(&shopExists); err != nil || shopExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop_id"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	createdIDs := make([]int, 0, len(req.Items))
	for _, item := range req.Items {
		name := strings.TrimSpace(item.Name)
		itemID, err := upsertItem(tx, name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve item"})
			return
		}
		result, err := tx.Exec(
			`INSERT INTO expenses (person_id, shop_id, item_id, date, amount) VALUES (?, ?, ?, ?, ?)`,
			req.PersonID, req.ShopID, itemID, req.Date, item.Amount,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create expense"})
			return
		}
		expenseID, _ := result.LastInsertId()
		if err := insertExpenseShares(tx, expenseID, item.Shares); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create expense shares"})
			return
		}
		createdIDs = append(createdIDs, int(expenseID))
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save expenses"})
		return
	}

	created := make([]models.Expense, 0, len(createdIDs))
	for _, id := range createdIDs {
		exp, err := h.fetchExpense(id)
		if err != nil {
			continue
		}
		created = append(created, exp)
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) UpdateExpense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	var req models.UpdateExpenseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PersonID != 1 && req.PersonID != 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "person_id must be 1 or 2"})
		return
	}

	if err := validateShares(req.Shares); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := validateWholeAmount(req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	name := strings.TrimSpace(req.Name)
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}

	var shopExists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM shops WHERE id = ?`, req.ShopID).Scan(&shopExists); err != nil || shopExists == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid shop_id"})
		return
	}

	var exists int
	if err := h.db.QueryRow(`SELECT COUNT(*) FROM expenses WHERE id = ?`, id).Scan(&exists); err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	itemID, err := upsertItem(tx, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to resolve item"})
		return
	}

	if _, err := tx.Exec(
		`UPDATE expenses SET person_id = ?, shop_id = ?, item_id = ?, date = ?, amount = ? WHERE id = ?`,
		req.PersonID, req.ShopID, itemID, req.Date, req.Amount, id,
	); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update expense"})
		return
	}

	if _, err := tx.Exec(`DELETE FROM expense_shares WHERE expense_id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to replace expense shares"})
		return
	}

	if err := insertExpenseShares(tx, int64(id), req.Shares); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create expense shares"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save expense"})
		return
	}

	exp, err := h.fetchExpense(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"id": id})
		return
	}

	c.JSON(http.StatusOK, exp)
}

func (h *Handler) DeleteExpense(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid expense id"})
		return
	}

	tx, err := h.db.Begin()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start transaction"})
		return
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM expense_shares WHERE expense_id = ?`, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete expense shares"})
		return
	}

	result, err := tx.Exec(`DELETE FROM expenses WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete expense"})
		return
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "expense not found"})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to commit delete"})
		return
	}

	c.Status(http.StatusNoContent)
}
