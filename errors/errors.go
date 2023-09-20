package errors

import "github.com/gin-gonic/gin"

type Responce struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Code    int    `json:"code"`
	Data    any    `json:"data"`
}

func NewSuccess(msg string, data any) (int, Responce) {
	return 200, Responce{
		Success: true,
		Msg:     msg,
		Code:    200,
		Data:    data,
	}
}

func NewInvalidArgment(msg string, data any) (int, Responce) {
	return 400, Responce{
		Success: false,
		Msg:     msg,
		Code:    400,
		Data:    data,
	}
}

func NewError(code int, msg string, data any) (int, Responce) {
	return code, Responce{
		Success: false,
		Msg:     msg,
		Code:    code,
		Data:    data,
	}
}
func Error(code int, data any, msg string) (int, *gin.H) {
	return code, &gin.H{
		"success": false,
		"data":    data,
		"msg":     msg,
		"code":    code,
	}
}

func Sucess(data any) (int, *gin.H) {
	return 200, &gin.H{
		"success": true,
		"data":    data,
		"msg":     "",
		"code":    200,
	}
}
