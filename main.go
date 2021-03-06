package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	"golang.org/x/time/rate"

	model "github.com/champnc/sample-grocery-api/model"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"


	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/gin-swagger/swaggerFiles"
	_ "github.com/champnc/sample-grocery-api/docs"
)

var (
	r = rate.Every(2 * time.Second)
	lim = rate.NewLimiter(r ,1)
)

func main() {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db.AutoMigrate(&model.Product{})

	r := gin.Default()
	handler := newHandler(db,lim)

	r.POST("/login", loginHandler)

	protected := r.Group("/", authorizationMiddleware)

	protected.GET("/grocery/:id", handler.getProductHandler)
	protected.GET("/grocery", handler.getProductListHandler)
	protected.POST("/grocery", handler.createProductHandler)
	protected.DELETE("/grocery/:id", handler.deleteProductHandler)

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

type Handler struct {
	db *gorm.DB
	lim *rate.Limiter
}

func newHandler(db *gorm.DB, lim *rate.Limiter) *Handler {
	return &Handler{db,lim}
}

func loginHandler(c *gin.Context) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.StandardClaims{
		ExpiresAt: time.Now().Add(5 * time.Minute).Unix(),
	})

	ss, err := token.SignedString([]byte("MySignature"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"token": ss,
	})
}

// @Summary      Show a product
// @Description  get product by ID
// @Tags         product
// @Produce      json
// @Param        id   path      int  true  "product ID"
// @Success      200  {object}  model.Product
// @Failure      400  {string}  string ""
// @Router       /grocery/{id} [get]
func (h *Handler) getProductHandler(c *gin.Context) {
	s := c.Request.Header.Get("Authorization")

	token := strings.TrimPrefix(s, "Bearer ")

	if !h.lim.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many request",
		})
		return
	}

	if err := validateToken(token); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	var product model.Product

	if err := h.db.Where("id = ?", c.Param("id")).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Record not found!"})
		return
	  }

	c.JSON(http.StatusOK, &product)
}

func (h *Handler) getProductListHandler(c *gin.Context) {
	var product []model.Product

	if !h.lim.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many request",
		})
		return
	}

	if result := h.db.Find(&product); result.Error != nil {
		return
	}

	c.JSON(http.StatusOK, &product)
}

func (h *Handler) deleteProductHandler(c *gin.Context) {
	var product model.Product

	if !h.lim.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many request",
		})
		return
	}

	if result := h.db.Delete(&product, c.Params); result.Error != nil {
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *Handler) createProductHandler(c *gin.Context) {
	var product model.Product

	if !h.lim.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": "too many request",
		})
		return
	}

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

func validateToken(token string) error {
	_, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}

		return []byte("MySignature"), nil
	})

	return err
}

func authorizationMiddleware(c *gin.Context) {
	s := c.Request.Header.Get("Authorization")

	token := strings.TrimPrefix(s, "Bearer ")

	if err := validateToken(token); err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
}
