package middleware

import (
	"orm/dao"
	"strconv"

	"github.com/gin-gonic/gin"
)

// http://{endpoints}/{router}?fid=xxxx
// 判断fid是否存在且有效，并且检查fid和userId的对应情况,需要uid，如果解析成功，提供fid对应storeId
// 一定要放在checktoken中间件后面
func ValidateFid(ctx *gin.Context) {

	fids := ctx.Query("fid")

	fid, err := strconv.Atoi(fids)
	if err != nil {
		ctx.Abort()
		ctx.HTML(304, "fid-fail.html", gin.H{
			"title": "fobidden",
			"msg":   err.Error(),
		})
		return
	}
	uid, _ := ctx.Get("uid")

	file, err := dao.QueryFileById(fid)
	if err != nil {
		ctx.Abort()
		ctx.HTML(404, "404.html", gin.H{
			"title": "not-found",
			"msg":   err.Error(),
		})
		return
	}
	st, err := dao.QueryStoreInfo(file.StoreId)
	if err != nil {
		ctx.Abort()
		ctx.HTML(404, "404.html", gin.H{
			"title": "not-found",
			"msg":   err.Error(),
		})
		return
	}
	if st.UserId == uid.(int) {
		ctx.Set("storeId", st.ID)
		ctx.Next()
	} else {
		ctx.Abort()
	}
}
