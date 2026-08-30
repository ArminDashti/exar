package database

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type ExpenseFilters struct {
	Owner    string
	FromDate string
	ToDate   string
}

func InsertExpense(form map[string]string) error {
	conn, err := dbConn()
	if err != nil {
		return err
	}

	insertedAt := strings.TrimSpace(form["inserted_at"])
	var insertedAtValue any
	if insertedAt == "" {
		insertedAtValue = nil
	} else {
		insertedAtValue = insertedAt
	}

	_, err = conn.ExecContext(
		context.Background(),
		`INSERT INTO expenses (id, "store-id", purchase, "purchase-time", owner, "inserted-at")
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		strings.TrimSpace(form["id"]),
		strings.TrimSpace(form["store_id"]),
		strings.TrimSpace(form["purchase"]),
		strings.TrimSpace(form["purchase_time"]),
		strings.TrimSpace(form["owner"]),
		insertedAtValue,
	)
	if err != nil {
		return fmt.Errorf("insert expense: %w", err)
	}
	return nil
}

func AddExpense(owner, store, purchase, purchaseDate string) error {
	conn, err := dbConn()
	if err != nil {
		return err
	}

	_, err = conn.ExecContext(
		context.Background(),
		`INSERT INTO expenses (id, "store-id", purchase, "purchase-time", owner)
		 VALUES ($1, $2, $3, $4, $5)`,
		fmt.Sprintf("api-%d", timeNowUnixNano()),
		store,
		purchase,
		purchaseDate,
		owner,
	)
	if err != nil {
		return fmt.Errorf("add expense: %w", err)
	}
	return nil
}

func GetExpenses(filters ExpenseFilters) ([]map[string]string, error) {
	conn, err := dbConn()
	if err != nil {
		return nil, err
	}

	query := `SELECT id, "store-id", purchase::text, "purchase-time"::text, owner::text
	          FROM expenses WHERE 1=1`
	args := make([]any, 0, 3)
	argIndex := 1

	if owner := strings.TrimSpace(filters.Owner); owner != "" {
		query += fmt.Sprintf(" AND owner = $%d", argIndex)
		args = append(args, owner)
		argIndex++
	}
	if fromDate := strings.TrimSpace(filters.FromDate); fromDate != "" {
		query += fmt.Sprintf(" AND DATE(\"purchase-time\") >= $%d::date", argIndex)
		args = append(args, fromDate)
		argIndex++
	}
	if toDate := strings.TrimSpace(filters.ToDate); toDate != "" {
		query += fmt.Sprintf(" AND DATE(\"purchase-time\") <= $%d::date", argIndex)
		args = append(args, toDate)
	}

	rows, err := conn.QueryContext(context.Background(), query, args...)
	if err != nil {
		return nil, fmt.Errorf("list expenses: %w", err)
	}
	defer rows.Close()

	records := make([]map[string]string, 0)
	for rows.Next() {
		var id, storeID, purchase, purchaseTime, ownerValue string
		if scanErr := rows.Scan(&id, &storeID, &purchase, &purchaseTime, &ownerValue); scanErr != nil {
			return nil, fmt.Errorf("scan expense: %w", scanErr)
		}
		purchaseDate := purchaseTime
		if len(purchaseDate) >= 10 {
			purchaseDate = purchaseDate[:10]
		}
		records = append(records, map[string]string{
			"id":             id,
			"store_id":       storeID,
			"store":          storeID,
			"purchase":       purchase,
			"purchase_time":  purchaseTime,
			"purchase-date":  purchaseDate,
			"owner":          ownerValue,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate expenses: %w", err)
	}
	return records, nil
}

func DeleteExpense(purchase, purchaseDate string) (int64, error) {
	conn, err := dbConn()
	if err != nil {
		return 0, err
	}

	result, err := conn.ExecContext(
		context.Background(),
		`DELETE FROM expenses
		 WHERE purchase::text = $1
		   AND DATE("purchase-time") = $2::date`,
		purchase,
		purchaseDate,
	)
	if err != nil {
		return 0, fmt.Errorf("delete expense: %w", err)
	}

	deleted, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete expense rows affected: %w", err)
	}
	return deleted, nil
}

func timeNowUnixNano() int64 {
	return time.Now().UnixNano()
}
