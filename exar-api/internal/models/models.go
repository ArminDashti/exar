package models

type Person struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Shop struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Item struct {
	ID        int     `json:"id,omitempty"`
	ShopID    int     `json:"shop_id,omitempty"`
	Shop      string  `json:"shop,omitempty"`
	ProductID int     `json:"product_id,omitempty"`
	Product   string  `json:"product,omitempty"`
	PersonID  int     `json:"person_id,omitempty"`
	Person    string  `json:"person,omitempty"`
	Amount    float64 `json:"amount"`
	Date      string  `json:"date"`
}

type CreateItemRequest struct {
	Shop     string  `json:"shop" binding:"required"`
	Product  string  `json:"product" binding:"required"`
	PersonID int     `json:"person_id" binding:"required"`
	Amount   float64 `json:"amount" binding:"required"`
	Date     string  `json:"date" binding:"required"`
}

type CreateShopRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateProductRequest struct {
	Name string `json:"name" binding:"required"`
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
