package controller

import (
	"net/http"
	"orm/errors"
	service "orm/service/filesystem"

	"github.com/gin-gonic/gin"
)

// /url?fid=xxxx
// @guard by checkToken.go -> validatefid.go
func GetObjectUrl(ctx *gin.Context) {
	//_, _ := ctx.Get("uid")
	// check fid below the uid
	req := service.GetObjectUrlRequest{}
	res := service.GetObjectUrlResponce{}

	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err := service.GetObjectUrl(&req, &res)

	if err != nil {
		ctx.JSON(errors.Error(http.StatusBadRequest, gin.H{}, err.Error()))
	} else {
		ctx.JSON(errors.Sucess(&res))
	}
}

// /url?storeId=xx&Path=xxx
func GetObjectByPath(ctx *gin.Context) {
	req := service.GetObjectUrl2Request{}
	res := service.GetObjectUrl2Responce{}
	if err := ctx.ShouldBindQuery(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{})
		return
	}

	err := service.GetObjectUrl2(&req, &res)
	if err != nil {
		ctx.JSON(errors.Error(http.StatusBadRequest, gin.H{}, err.Error()))
	} else {
		ctx.JSON(errors.Sucess(&res))
	}
}

// /dl?url=xxx
func DownLoadUrl(ctx *gin.Context) {

}
