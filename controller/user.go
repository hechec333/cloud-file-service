package controller

import (
	"fmt"
	"orm/errors"
	service "orm/service/user"

	"github.com/gin-gonic/gin"
)

// post application json
// /register?cid=xx
func UserRegister(ctx *gin.Context) {
	request := service.UserRegisterRequest{}
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println(err)
		ctx.JSON(errors.Error(304, gin.H{}, err.Error()))
	} else {
		request.Cid = ctx.Query("cid")
		res := service.UserRegisterResponce{}
		err := service.UserRegister(&request, &res)
		if err != nil {
			ctx.JSON(errors.NewInvalidArgment(err.Error(), gin.H{}))
		} else {
			ctx.JSON(errors.NewSuccess("", res))
		}

	}

}

// post
func UserLogin(ctx *gin.Context) {
	request := service.UserLoginRequest{}
	if err := ctx.BindJSON(&request); err != nil {
		fmt.Println(err)
		ctx.JSON(errors.Error(304, gin.H{}, err.Error()))
	}

	token := ctx.GetHeader("access-token")
	request.Token = token
	resp := service.UserLoginResponce{}
	err := service.UserLogin(&request, &resp)
	if err != nil {
		ctx.JSON(errors.NewInvalidArgment(err.Error(), gin.H{}))
	} else {
		ctx.Header("access-token", resp.Token)
		ctx.JSON(errors.NewSuccess("", resp))
	}

}

// get/post
func UserInfo(ctx *gin.Context) {
	token := ctx.GetHeader("access-token")
	if token == "" {
		ctx.JSON(errors.Error(304, gin.H{}, "forbidden"))
		return
	}
	if ctx.Request.Method == "GET" {
		request := service.UserInfoRequest{}
		request.Token = token
		resp := service.UserInfoResponce{}
		err := service.UserGetInfo(&request, &resp)
		if err != nil {
			ctx.JSON(errors.NewInvalidArgment(err.Error(), gin.H{}))
		} else {
			ctx.JSON(errors.NewSuccess("", resp))
		}
	} else {
		request := service.UserUpdateInfoRequest{}
		request.Token = token

		if err := ctx.BindJSON(&request); err != nil {
			fmt.Println(err)
			ctx.JSON(errors.Error(304, gin.H{}, err.Error()))
		}
		resp := service.UserUpdateInfoResponce{}
		if err := service.UserUpdateInfo(&request, &resp); err != nil {
			ctx.JSON(errors.NewInvalidArgment(err.Error(), gin.H{}))
		} else {
			ctx.JSON(errors.NewSuccess("", resp))
		}

	}
}
