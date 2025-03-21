package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func CreateItem(c *gin.Context) {
	var item Item
	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
		return
	}
	now := time.Now()
	item.LastRestock = now

	if result := DB.Create(&item); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to create item"})
		return
	}
	c.JSON(http.StatusCreated, item)
}

func GetItems(c *gin.Context) {
	var items []Item
	if result := DB.Find(&items); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetLowStockItems(c *gin.Context) {
	var items []Item
	if result := DB.Where("quantity <= ?", 20).Find(&items); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to retrieve low stock items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

func GetItem(c *gin.Context) {
	id := c.Param("id")
	var item Item
	if result := DB.First(&item, id); result.Error != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "item not found"})
		return
	}

	c.JSON(http.StatusOK, item)
}

func UpdateItem(c *gin.Context) {
	id := c.Param("id")
	var existingItem Item
	if result := DB.First(&existingItem, id); result.Error != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "item not found"})
		return
	}

	var updateData Item
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid request format"})
		return
	}

	updates := map[string]interface{}{
		"name":        updateData.Name,
		"description": updateData.Description,
		"quantity":    updateData.Quantity,
	}

	if result := DB.Model(&existingItem).Updates(updates); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update item"})
		return
	}

	DB.First(&existingItem, id)
	c.JSON(http.StatusOK, existingItem)
}

func DeleteItem(c *gin.Context) {
	id := c.Param("id")
	var item Item
	if result := DB.First(&item, id); result.Error != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "item not found"})
		return
	}

	if result := DB.Delete(&item); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to delete item"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "item deleted successfully"})
}

func RestockItem(c *gin.Context) {
	id := c.Param("id")
	itemID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid item ID"})
		return
	}

	var req RestockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "out of range (10-1000)"})
		return
	}

	var item Item
	if result := DB.First(&item, id); result.Error != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "item not found"})
		return
	}

	var restockCount int64
	timeCutoff := time.Now().Add(-24 * time.Hour)
	if result := DB.Model(&RestockHistory{}).Where("item_id = ? AND created_at > ?", itemID, timeCutoff).Count(&restockCount); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to check restock history"})
		return
	}

	if restockCount >= 3 {
		c.JSON(http.StatusTooManyRequests, ErrorResponse{Error: "rate limit exceeded: maximum 3 restocks per item in 24 hours"})
		return
	}

	restockHistory := RestockHistory{
		ItemID:    uint(itemID),
		Amount:    req.Amount,
		Timestamp: time.Now(),
	}

	if result := DB.Create(&restockHistory); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "restock history failed"})
		return
	}

	item.Quantity += req.Amount
	item.LastRestock = time.Now()
	if result := DB.Save(&item); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "failed to update item quantity"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "item restocked successfully",
		"item":     item,
		"restocks": restockCount + 1,
	})
}

func GetRestockHistory(c *gin.Context) {
	id := c.Param("id")
	itemID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid ID"})
		return
	}

	var item Item
	if result := DB.First(&item, id); result.Error != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "item not found"})
		return
	}

	var history []RestockHistory
	if result := DB.Where("item_id = ?", itemID).Order("created_at desc").Find(&history); result.Error != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "couldnt retrieve restock history"})
		return
	}

	c.JSON(http.StatusOK, history)
}
