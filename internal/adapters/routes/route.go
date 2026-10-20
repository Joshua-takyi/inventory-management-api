package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/joshua-takyi/stockflow-api/container"
)

func Router(ctn *container.Container) *gin.Engine {
	r := gin.Default()
	r.GET("/", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "welcome",
		})
	})
	v1 := r.Group("/api/v1")
	v1.GET("/healthz", ctn.HealthHandler.Health)
	v1.POST("/products", ctn.ProductHandler.Create())
	v1.GET("/products", ctn.ProductHandler.List())
	v1.GET("/products/:id", ctn.ProductHandler.GetByID())
	v1.PATCH("/products/:id", ctn.ProductHandler.UpdateProduct())
	v1.PATCH("/products/remove/:id", ctn.ProductHandler.Delete())
	return r
}
