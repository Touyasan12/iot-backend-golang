package handlers

import (
	"net/http"

	"iot-backend-cursor/database"
	"iot-backend-cursor/models"

	"github.com/gin-gonic/gin"
)

// GetStock returns current food stock
func GetStock(c *gin.Context) {
	var stock models.Stock
	if err := database.DB.First(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stock not found"})
		return
	}

	c.JSON(http.StatusOK, stock)
}

// UpdateStock updates the food stock
func UpdateStock(c *gin.Context) {
	var stock models.Stock
	if err := database.DB.First(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Stock not found"})
		return
	}

	var updateReq struct {
		AmountGram int `json:"amount_gram" binding:"required"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updateReq.AmountGram < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Amount cannot be negative"})
		return
	}

	stock.AmountGram = updateReq.AmountGram
	if err := database.DB.Save(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stock)
}

