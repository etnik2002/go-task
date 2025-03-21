package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("no env file")
	}
	initDB()
	r := gin.Default()
	setupRoutes(r)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("server running on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func initDB() {
	dbPath := "inventory_management.db"

	var err error
	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("db conn failedf: %v", err)
	}

	err = DB.AutoMigrate(&User{}, &Item{}, &RestockHistory{})
	if err != nil {
		log.Fatalf("db migration failed: %v", err)
	}
}

func setupRoutes(r *gin.Engine) {
	r.POST("/register", Register)
	r.POST("/login", Login)

	authorized := r.Group("/api")
	authorized.Use(JWTAuthMiddleware())
	{
		authorized.GET("/items/:id/restock-history", GetRestockHistory)
		authorized.GET("/items", GetItems)
		authorized.GET("/items/low-stock", GetLowStockItems)
		authorized.GET("/items/:id", GetItem)
		authorized.POST("/items/:id/restock", RestockItem)
		authorized.POST("/items", CreateItem)
		authorized.PUT("/items/:id", UpdateItem)
		authorized.DELETE("/items/:id", DeleteItem)
	}
}
