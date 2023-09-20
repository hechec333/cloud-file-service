package router

import (
	"orm/controller"
	"orm/middleware"

	"github.com/gin-gonic/gin"
)

func ImportRoutes() *gin.Engine {

	gin := gin.Default()

	gin.Use(middleware.HttpRecover)

	gin.GET("/captcha/vertify/:cid", controller.VertifyCaptcha)
	capt := gin.Group("/captcha")
	middleware.GenRaterLimiter("captcha", 100, 200)
	capt.Use(middleware.CaptchaRateLimte)
	{
		capt.GET("/get", controller.GetCaptcha)
		capt.GET("/refresh/:cid", controller.RefreshCaptcha)
	}

	group := gin.Group("/u")
	{
		group.POST("/register", controller.UserRegister)
		group.POST("/login", controller.UserLogin)
		group.GET("/info", controller.UserInfo)
		group.POST("/info", controller.UserInfo)
	}

	fgroup := gin.Group("/fs")

	fgroup.Use(middleware.CheckToken)
	{

	}

	middleware.Init()
	return gin
}
