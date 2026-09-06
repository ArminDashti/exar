package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/armin/expenses/backend/internal/auth"
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
		api.POST("/auth/login", h.Login)

		secured := api.Group("")
		secured.Use(auth.Middleware())
		{
			secured.GET("/persons", h.ListPersons)
			secured.GET("/shops", h.ListShops)
			secured.POST("/shops", h.CreateShop)
			secured.PUT("/shops/:id", h.UpdateShop)
			secured.DELETE("/shops/:id", h.DeleteShop)
			secured.GET("/items", h.ListItems)
			secured.POST("/items", h.CreateItem)
			secured.PUT("/items/:id", h.UpdateItem)
			secured.DELETE("/items/:id", h.DeleteItem)
			secured.GET("/stats", h.GetStats)
			secured.GET("/expenses/check-duplicate", h.CheckDuplicateExpense)
			secured.GET("/expenses", h.ListExpenses)
			secured.GET("/expenses/:id", h.GetExpense)
			secured.POST("/expenses", h.CreateExpenses)
			secured.PUT("/expenses/:id", h.UpdateExpense)
			secured.DELETE("/expenses/:id", h.DeleteExpense)
		}
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
