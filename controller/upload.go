package controller

import (
	"net/http"
	"orm/errors"
	service "orm/service/filesystem"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
)

// 1.文件大小小于 128MB 不做任何处理
// 2.文件大小如果大于 128MB使用 流式上传 ,如果没有检测到mutipart标志，则拒绝
// 3.文件大小如果大于 1G使用 分块上传 ，必须显式分块成256MB的块来上传，推荐多线程上传
const (
	FILE_SIZE_LIMITS   = 1 << 30
	FILE_MEMORY_LIMITS = 128 << 20
)

// /ul/:fid?
func UploadByPid(ctx *gin.Context) {

	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, FILE_SIZE_LIMITS)
	ctx.Request.ParseMultipartForm(FILE_MEMORY_LIMITS)

	parts, _ := ctx.MultipartForm()
	req := service.CreateFileRequest{}

	v, is := parts.Value["type"]
	if !is {
		ctx.JSON(errors.Error(400, gin.H{}, "field type missing"))
		return
	}

	req.FileType = v[0]
	req.ParentId = ctx.GetInt("fid")
	file := parts.File["file"][0]
	req.Size = file.Size
	req.FileName = path.Base(file.Filename)
	req.FileSuffix = path.Ext(file.Filename)
	f, _ := file.Open()
	defer f.Close()
	req.R = f
	res := service.CreateFileResponce{}
	err := service.CreateFile(&req, &res)

	if err != nil {
		ctx.JSON(errors.NewInvalidArgment(err.Error(), gin.H{}))
	} else {
		ctx.JSON(errors.NewSuccess("", res))
	}
}

// 需要验证 stoerId 有效性的中间件
// /ul/path?mode=
func UploadByPath(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, FILE_SIZE_LIMITS)
	ctx.Request.ParseMultipartForm(FILE_MEMORY_LIMITS)

	parts, err := ctx.MultipartForm()

	if err != nil {
		ctx.JSON(errors.NewInvalidArgment(err.Error(), gin.H{}))
	}
	spath := parts.Value["path"][0]
	if spath[0] != '/' || len(spath) == 0 {
		spath += "/"
	}
	req := service.CreateFile2Requeset{}
	req.ParentPath = spath
	req.Mode = ctx.Query("mode")
	req.StoreId = ctx.GetInt("storeId")
	req.FileType = parts.Value["type"][0]
	file := parts.File["file"][0]

	req.FileName = path.Base(file.Filename)
	req.FileSuffix = path.Ext(file.Filename)
	res := service.CreateFileResponce{}
	service.CreateFile2(&req, &res)

	ctx.JSON(errors.NewSuccess("", res))
}

// @post /ul/chunk?chunkid=xx&uploadToken=xx
func ChunkUpload(ctx *gin.Context) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, 256<<20)
	sreq := service.ShardingUplaodRequest{
		CreateFileRequest: nil,
		UploadToken:       ctx.Query("uploadToken"),
	}

	res := service.ShardingUploadResponce{}
	if ctx.Query("chuckId") == "0" {
		sreq.ChunckId = 0
		ctx.Request.ParseMultipartForm(FILE_MEMORY_LIMITS)
		parts, _ := ctx.MultipartForm()
		req := service.CreateFileRequest{}
		req.FileType = parts.Value["type"][0]
		req.ParentId = ctx.GetInt("fid")
		file := parts.File["file"][0]
		req.Size = file.Size
		req.FileName = path.Base(file.Filename)
		req.FileSuffix = path.Ext(file.Filename)
		f, _ := file.Open()
		defer f.Close()
		sreq.CreateFileRequest = &req
	} else {
		cid, err := strconv.Atoi(ctx.Query("chunkId"))
		if err != nil {
			ctx.JSON(errors.Error(400, gin.H{}, err.Error()))
			return
		}
		sreq.ChunckId = cid
	}

	err := service.ShardingUpload(&sreq, &res)
	if err != nil {
		ctx.JSON(errors.Error(400, gin.H{}, err.Error()))
		return
	} else {
		ctx.JSON(errors.NewSuccess("", res))
	}

}

// @post /ul/chunk/q?uploadToken=xx
func OverChunkUpload(ctx *gin.Context) {

	req := service.FininshShardingUploadRequest{
		UploadToken: ctx.Query("uploadToken"),
	}
	res := service.FinishShardingUploadResponce{}

	err := service.FinishShardingUpload(&req, &res)

	if err != nil {
		ctx.JSON(errors.Error(400, gin.H{}, err.Error()))
	} else {
		ctx.JSON(errors.Sucess(res))
	}
}

// @get /ul/chunk?uploadToken=xx
func GetChunkParts(ctx *gin.Context) {

	res, err := service.ListShardingSlice(ctx.Query("uploadToken"))
	if err != nil {
		ctx.JSON(errors.Error(400, gin.H{}, err.Error()))
	} else {
		ctx.JSON(errors.Sucess(res))
	}
}
