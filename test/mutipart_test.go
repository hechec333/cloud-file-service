package test_test

import (
	"log"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestMutipartForm(t *testing.T) {

	engine := gin.Default()

	engine.POST("/upload", func(ctx *gin.Context) {

		ctx.Request.ParseMultipartForm(128 << 20)
		parts, err := ctx.MultipartForm()
		if err != nil {
			ctx.JSON(400, gin.H{
				"error": err,
			})
		}
		log.Println(parts)

		ctx.JSON(200, gin.H{
			"status": "ok",
		})
	})

	if err := engine.Run(":8080"); err != nil {
		panic(err)
	}
}
