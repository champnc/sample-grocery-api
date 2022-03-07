package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"net/http"
)

type Product struct {
	gorm.Model
	Name  string
	Code  string
	Price uint
}

var db *gorm.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&Product{})

	r := gin.Default()

	r.GET("/grocery/:id", getProductHandler)
	r.GET("/grocery", getProductListHandler)
	r.POST("/grocery", createProductHandler)
	r.DELETE("/grocery/:id", deleteProductHandler)

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func getProductHandler(c *gin.Context) {
	var product Product

	if err := db.Where("id = ?", c.Param("id")).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	  }

	c.JSON(http.StatusOK, &product)
}

func getProductListHandler(c *gin.Context) {
	var product []Product

	if result := db.Find(&product); result.Error != nil {
		return
	}

	c.JSON(http.StatusOK, &product)
}

func deleteProductHandler(c *gin.Context) {
	var product Product

	if result := db.Delete(&product, c.Params); result.Error != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func createProductHandler(c *gin.Context) {
	var product Product

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	if result := db.Create(&product); result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, &product)
}
