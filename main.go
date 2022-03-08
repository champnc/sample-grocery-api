package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
	model "github.com/champnc/sample-grocery-api/model"
)

func main() {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Product{})

	r := gin.Default()
	handler := newHandler(db)

	r.GET("/grocery/:id", handler.getProductHandler)
	r.GET("/grocery", handler.getProductListHandler)
	r.POST("/grocery", handler.createProductHandler)
	r.DELETE("/grocery/:id", handler.deleteProductHandler)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type Handler struct {
	db *gorm.DB
}

func newHandler(db *gorm.DB) *Handler {
	return &Handler{db}
}

func (h *Handler) getProductHandler(c *gin.Context) {
	var product model.Product

	if err := h.db.Where("id = ?", c.Param("id")).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	  }

	c.JSON(http.StatusOK, &product)
}

func (h *Handler) getProductListHandler(c *gin.Context) {
	var product []model.Product

	if result := h.db.Find(&product); result.Error != nil {
		return
	}

	c.JSON(http.StatusOK, &product)
}

func (h *Handler) deleteProductHandler(c *gin.Context) {
	var product model.Product

	if result := h.db.Delete(&product, c.Params); result.Error != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) createProductHandler(c *gin.Context) {
	var product model.Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result := h.db.Create(&product); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &product)
}
