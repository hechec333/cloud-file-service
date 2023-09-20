package middleware

import (
	"log"
	"orm/common/jwtx"
	"orm/config"

	"github.com/gin-gonic/gin"
)

func CheckToken(c *gin.Context) {
	token := c.GetHeader("access-token")
	conf := config.GetConfig()
	uid, err := jwtx.ParseToken(token, conf.Secret.AccessSecret)
	if err != nil {
		log.Println(c.ClientIP(), " token:", token, "fail to auth")
		c.HTML(304, "auth-fail.html", gin.H{
			"title": "fobidden",
			"msg":   "token not valid!",
		})
		return
	}

	c.Set("uid", uid)

	c.Next()
}
