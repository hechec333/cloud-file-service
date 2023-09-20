package service

import (
	"orm/common/upload"
	"orm/config"
	"orm/dao"
	"orm/dao/cache"
	"orm/reposity"
	"time"
)

type ShardingUplaodRequest struct {
	*CreateFileRequest
	ChunckId    int
	UploadToken string
}

type ShardingUploadResponce struct {
	CreateFileResponce
	UploadToken   string `json:"uploadToken"`
	ChunckId      int    `json:"chunckId"`
	ContentLength int64  `json:"contentLength"`
	Md5           string `json:"md5"`
}

func ShardingUpload(req *ShardingUplaodRequest, res *ShardingUploadResponce) error {
	var file dao.File
	var uploadId string
	conf := config.GetConfig()
	if req.UploadToken == "" {
		folder, err := dao.QueryDirById(req.ParentId)
		if err != nil {
			return err
		}
		f := dao.File{
			StoreId:    folder.StoreId,
			ParentId:   folder.ID,
			CreateTime: time.Now(),
			FileType:   req.FileType,
			FileSize:   req.Size,
			FileName:   req.FileName + req.FileSuffix,
		}

		err = dao.CreateFile(&f)
		if err != nil {
			return err
		}
		file = f
	} else {

		ulId, fileId, err := upload.ParseUploadToken(req.UploadToken, conf.Secret.AccessSecret)
		if err != nil {
			return err
		}

		file, err = dao.QueryFileById(fileId)
		if err != nil {
			return err
		}

		uploadId = ulId
	}
	result, err := reposity.CreateShardingFile(&file, uploadId, req.ChunckId, req.R)
	if err != nil {
		return err
	}
	res.ChunckId = result.ChunckId
	res.UploadToken, _ = upload.GetUploadToken(conf.Secret.AccessSecret, time.Now().Unix(), 3600, result.UploadId, file.FileId)
	return nil
}

type FininshShardingUploadRequest struct {
	UploadToken string
}
type ShardingResult struct {
	ContentLength int64  `json:"contentLength"`
	ChunckId      int    `json:"chunkId"`
	MD5           string `json:"md5"`
}

type FinishShardingUploadResponce struct {
	f     dao.File
	Last  *ShardingResult
	Total []*ShardingResult
}

func FinishShardingUpload(req *FininshShardingUploadRequest, res *FinishShardingUploadResponce) error {
	conf := config.GetConfig()
	ulId, fileId, err := upload.ParseUploadToken(req.UploadToken, conf.Secret.AccessSecret)
	if err != nil {
		return err
	}

	file, err := dao.QueryFileById(fileId)
	if err != nil {
		return err
	}

	ud, err := reposity.FinishShardingFile(&file, ulId)
	if err != nil {
		return err
	}

	res.f = file
	res.Last = &ShardingResult{
		ChunckId:      ud.ChunckId,
		ContentLength: ud.ContentLengh,
		MD5:           ud.Md5,
	}

	uds, err := reposity.ListShardingBlock(&file, ulId)

	if err != nil {
		return err
	}
	var totalSize int64 = 0
	for _, ud := range uds {
		res.Total = append(res.Total, &ShardingResult{
			ChunckId:      ud.ChunckId,
			ContentLength: ud.ContentLengh,
			MD5:           ud.Md5,
		})
		totalSize += ud.ContentLengh
	}

	dao.IncrFileSize(&file, totalSize)

	return nil
}

func ListShardingSlice(uploadToken string) ([]*ShardingResult, error) {
	conf := config.GetConfig()
	results := []*ShardingResult{}
	ulId, fileId, err := upload.ParseUploadToken(uploadToken, conf.Secret.AccessSecret)
	if err != nil {
		return results, err
	}
	file, err := dao.QueryFileById(fileId)
	if err != nil {
		return results, err
	}
	_, err = cache.GetKey("event:upload:" + ulId)
	if err == nil {
		uds, err := reposity.ListShardingBlock(&file, ulId)
		if err != nil {
			return results, err
		}
		for _, v := range uds {
			results = append(results, &ShardingResult{
				ChunckId:      v.ChunckId,
				ContentLength: v.ContentLengh,
				MD5:           v.Md5,
			})
		}
	}

	return results, err
}
