package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/armin/expenses/backend/internal/database"
	"github.com/armin/expenses/backend/internal/handlers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	dbPath := envOr("DATABASE_PATH", "./data/expenses.db")
	staticDir := envOr("STATIC_DIR", "./static")
	addr := envOr("ADDR", ":8080")

	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(cors.Default())

	h := handlers.New(db)

	api := r.Group("/api")
	{
		api.GET("/persons", h.ListPersons)
		api.GET("/shops", h.ListShops)
		api.POST("/shops", h.CreateShop)
		api.PUT("/shops/:id", h.UpdateShop)
		api.DELETE("/shops/:id", h.DeleteShop)
		api.GET("/items", h.ListItems)
		api.POST("/items", h.CreateItem)
		api.PUT("/items/:id", h.UpdateItem)
		api.DELETE("/items/:id", h.DeleteItem)
		api.GET("/stats", h.GetStats)
		api.GET("/expenses/check-duplicate", h.CheckDuplicateExpense)
		api.GET("/expenses", h.ListExpenses)
		api.GET("/expenses/:id", h.GetExpense)
		api.POST("/expenses", h.CreateExpenses)
		api.PUT("/expenses/:id", h.UpdateExpense)
		api.DELETE("/expenses/:id", h.DeleteExpense)
	}

	if info, err := os.Stat(staticDir); err == nil && info.IsDir() {
		r.Static("/assets", filepath.Join(staticDir, "assets"))
		r.NoRoute(func(c *gin.Context) {
			if strings.HasPrefix(c.Request.URL.Path, "/api") {
				c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
				return
			}

			rel := strings.TrimPrefix(c.Request.URL.Path, "/")
			if rel != "" && !strings.Contains(rel, "..") {
				candidate := filepath.Join(staticDir, filepath.FromSlash(rel))
				if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
					c.File(candidate)
					return
				}
			}

			c.File(filepath.Join(staticDir, "index.html"))
		})
	}

	log.Printf("listening on %s (db: %s)", addr, dbPath)
	if err := r.Run(addr); err != nil {
		log.Fatal(err)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
