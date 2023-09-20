package middleware

import (
	"log"
	"orm/errors"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func HttpRecover(ctx *gin.Context) {

	defer func() {
		if r := recover(); r != nil {
			log.Println(r)
			debug.PrintStack()
			errors.Error(500, gin.H{}, "internal server error")
		}
	}()

	ctx.Next()
}
