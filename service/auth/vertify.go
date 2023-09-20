package auth

import (
	"io"
	"orm/common/code"
	"orm/common/util"
	"orm/dao/cache"
)

type GetCaptchaRequest struct {
	W io.Writer
}
type GetCaptchaResponce struct {
	Cid string
}

// 这个接口需要限流
func GetCaptcha(req *GetCaptchaRequest, res *GetCaptchaResponce) error {

	key := code.GenCaptcha(req.W)

	cid := util.Uuid()

	res.Cid = cid
	return cache.SetKey("uuid:"+cid, key, 0)
}

type RefreshCaptchaRequest struct {
	Cid string
	W   io.Writer
}

func RefreshCaptcha(req *RefreshCaptchaRequest) error {

	if _, err := cache.GetKey("uuid:" + req.Cid); err != nil {
		return err
	}

	key := code.GenCaptcha(req.W)

	return cache.SetKey("uuid:"+req.Cid, key, 0)
}

type VertifyCaptchaRequest struct {
	Cid  string
	Code string
}

type VertifyCaptchaResponce struct {
	Success bool
}

func VertifyCaptcha(req *VertifyCaptchaRequest, res *VertifyCaptchaResponce) error {
	var key string
	var err error
	if key, err = cache.GetKey("uuid:" + req.Cid); err != nil {
		return err
	}

	if key != req.Code {
		res.Success = false
	} else {
		res.Success = true
		cache.DelKey("uuid:" + req.Cid)
		cache.SetKey("uuid:"+req.Cid+":y", 1, 0) //
	}

	return nil
}
