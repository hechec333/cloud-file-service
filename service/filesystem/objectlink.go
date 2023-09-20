package service

import (
	"io"
	"orm/common/url"
	"orm/config"
	"orm/dao"
	"orm/reposity"
	"time"
)

type GetObjectUrlRequest struct {
	FileId int `form:"fid" json:"fid"`
}
type GetObjectUrlResponce struct {
	Url string `json:"url"`
}

func GetObjectUrl(req *GetObjectUrlRequest, res *GetObjectUrlResponce) error {

	file, err := dao.QueryFileById(req.FileId)
	if err != nil {
		return err
	}
	conf := config.GetConfig()
	res.Url, err = url.GetUrl(conf.Secret.AccessSecret, time.Now().Unix(), conf.Secret.AccessExpire, file.FileId)
	return err
}

type GetObjectUrl2Request struct {
	StoreId int    `form:"storeId" json:"storeId"`
	Path    string `form:"path" json:"path"`
}
type GetObjectUrl2Responce struct {
	Url string `json:"url"`
}

func GetObjectUrl2(req *GetObjectUrl2Request, res *GetObjectUrl2Responce) error {

	file, err := dao.QueryFileByPath(req.StoreId, req.Path)
	if err != nil {
		return err
	}
	conf := config.GetConfig()
	res.Url, err = url.GetUrl(conf.Secret.AccessSecret, time.Now().Unix(), conf.Secret.AccessExpire, file.FileId)
	return err
}

type DownLoadUrlObjectRequest struct {
	Url string
	w   io.Writer
}

func DownLoadUrlObject(req *DownLoadUrlObjectRequest) error {
	conf := config.GetConfig()
	fid, err := url.ParseUrl(req.Url, conf.Secret.AccessSecret)
	if err != nil {
		return err
	}

	file, err := dao.QueryFileById(int(fid))

	if err != nil {
		return err
	}

	return reposity.GetFile(file, req.w)
}
