package controller

import (
	"orm/errors"
	"orm/service/auth"

	"github.com/gin-gonic/gin"
)

// @/get
func GetCaptcha(ctx *gin.Context) {
	req := &auth.GetCaptchaRequest{
		W: ctx.Writer,
	}
	ctx.Header("Content-Type", "application/octet-stream")
	res := &auth.GetCaptchaResponce{}
	auth.GetCaptcha(req, res)
}

// @/refresh/:cid
func RefreshCaptcha(ctx *gin.Context) {
	cid := ctx.Param("cid")

	req := auth.RefreshCaptchaRequest{
		Cid: cid,
		W:   ctx.Writer,
	}
	ctx.Header("Content-Type", "application/octet-stream")
	err := auth.RefreshCaptcha(&req)

	if err != nil {
		ctx.JSON(errors.Error(400, gin.H{}, err.Error()))
		ctx.Abort()
		return
	}

}

// @/vertify/:cid?code=xxxx
func VertifyCaptcha(ctx *gin.Context) {

	req := auth.VertifyCaptchaRequest{
		Cid:  ctx.Param("cid"),
		Code: ctx.Query("code"),
	}
	res := auth.VertifyCaptchaResponce{}

	err := auth.VertifyCaptcha(&req, &res)

	if err != nil {
		ctx.JSON(errors.Error(400, gin.H{}, err.Error()))
		ctx.Abort()
		return
	} else {
		ctx.JSON(errors.NewSuccess("", gin.H{
			"success": res.Success,
		}))
	}
}
