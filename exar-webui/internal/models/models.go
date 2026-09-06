package models

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Shop struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ExpenseShare struct {
	PersonID   int     `json:"person_id"`
	PersonName string  `json:"person_name,omitempty"`
	Share      float64 `json:"share"`
}

type Expense struct {
	ID         int            `json:"id,omitempty"`
	PersonID   int            `json:"person_id"`
	ShopID     int            `json:"shop_id"`
	ItemID     int            `json:"item_id"`
	Date       string         `json:"date"`
	Name       string         `json:"name"`
	Amount     float64        `json:"amount"`
	PersonName string         `json:"person_name,omitempty"`
	ShopName   string         `json:"shop_name,omitempty"`
	Shares     []ExpenseShare `json:"shares"`
}

type ExpenseShareInput struct {
	PersonID int     `json:"person_id" binding:"required"`
	Share    float64 `json:"share"`
}

type CreateExpenseLine struct {
	Name    string              `json:"name" binding:"required"`
	Amount  float64             `json:"amount" binding:"required"`
	Shares  []ExpenseShareInput `json:"shares" binding:"required,min=1,dive"`
}

type CreateExpensesRequest struct {
	PersonID int                 `json:"person_id" binding:"required"`
	ShopID   int                 `json:"shop_id" binding:"required"`
	Date     string              `json:"date" binding:"required"`
	Items    []CreateExpenseLine `json:"items" binding:"required,min=1,dive"`
}

type UpdateExpenseRequest struct {
	PersonID int                 `json:"person_id" binding:"required"`
	ShopID   int                 `json:"shop_id" binding:"required"`
	Date     string              `json:"date" binding:"required"`
	Name     string              `json:"name" binding:"required"`
	Amount   float64             `json:"amount" binding:"required"`
	Shares   []ExpenseShareInput `json:"shares" binding:"required,min=1,dive"`
}

type CreateShopRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateShopRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateItemRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateItemRequest struct {
	Name string `json:"name" binding:"required"`
}

type DuplicateCheckResponse struct {
	Exists bool `json:"exists"`
	Count  int  `json:"count"`
}

type MonthStats struct {
	Month      string  `json:"month"`
	Armin      float64 `json:"armin"`
	Ramin      float64 `json:"ramin"`
	Total      float64 `json:"total"`
	ArminShare float64 `json:"armin_share"`
	RaminShare float64 `json:"ramin_share"`
}

type Stats struct {
	ByMonth []MonthStats `json:"by_month"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Username string `json:"username"`
	PersonID int    `json:"person_id"`
}
