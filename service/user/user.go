package service

import (
	"errors"
	"fmt"
	"orm/common/jwtx"
	"orm/config"
	"orm/dao"
	"orm/dao/cache"
	_ "orm/errors"
	"time"

	"github.com/jinzhu/gorm"
)

//	type User struct {
//		Id         int64
//		UserName   string `json:"userName"`
//		UserAvator string `json:"userAvator"`
//		UserAuth   string `json:"userAuth"`
//	}
type UserLoginRequest struct {
	UserName string `json:"userName"`
	UserAuth string `json:"userAuth"`
	Token    string
}
type UserLoginResponce struct {
	Token      string `json:"-"`
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserAvator string `json:"userAvator"`
}

func UserLogin(req *UserLoginRequest, res *UserLoginResponce) error {
	u, err := dao.QueryUserByName(req.UserName)
	if err != nil {
		return err
	}
	if u.UserAuth != req.UserAuth {
		return errors.New("userAuth not equal")
	}
	conf := config.GetConfig()
	token, err := jwtx.GetToken(conf.Secret.AccessSecret, time.Now().Unix(), conf.Secret.AccessExpire, u.ID)

	res.Token = token
	res.UserAuth = u.UserAuth
	res.UserAvator = u.UserAvator
	res.UserName = u.UserName

	return err
}

type UserInfoRequest struct {
	Token string
}
type UserInfoResponce struct {
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserAvator string `json:"userAvator"`
}

func UserGetInfo(req *UserInfoRequest, res *UserInfoResponce) error {
	conf := config.GetConfig()
	var err error
	var uid int64
	if uid, err = jwtx.ParseToken(req.Token, conf.Secret.AccessSecret); err != nil {
		return err
	}

	u, err := dao.QueryUserById(uid)
	if err != nil {
		return err
	}

	res.UserName = u.UserName
	res.UserAvator = u.UserAvator
	res.UserAuth = u.UserAuth
	return nil
}

type UserUpdateInfoRequest struct {
	Token      string
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserAvator string `json:"userAvator"`
}
type UserUpdateInfoResponce struct {
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserAvator string `json:"userAvator"`
}

func UserUpdateInfo(req *UserUpdateInfoRequest, res *UserUpdateInfoResponce) error {
	conf := config.GetConfig()
	var err error
	var uid int64
	if uid, err = jwtx.ParseToken(req.Token, conf.Secret.AccessSecret); err != nil {
		return err
	}
	u := dao.User{
		ID:         uid,
		UserName:   req.UserName,
		UserAuth:   req.UserAuth,
		UserAvator: req.UserAvator,
	}
	return dao.UpdateUser(u)
}

type UserRegisterRequest struct {
	Cid        string
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserAvator string `json:"userAvator"`
}
type UserRegisterResponce struct {
	UserName   string `json:"userName"`
	UserAuth   string `json:"userAuth"`
	UserAvator string `json:"userAvator"`
}

func UserRegister(req *UserRegisterRequest, res *UserRegisterResponce) error {

	// 验证验证码是否通过
	if _, err := cache.GetKey("uuid:" + req.Cid + ":y"); err != nil {
		return errors.New("cid not valid")
	}

	cache.DelKey("uuid:" + req.Cid + ":y")

	_, err := dao.QueryUserByName(req.UserName)

	fmt.Println(gorm.IsRecordNotFoundError(err))
	if err != nil && err.Error() != gorm.ErrRecordNotFound.Error() {
		return errors.Join(errors.New("UserRegister"), err)
	}
	u := dao.User{
		UserName:   req.UserName,
		UserAuth:   req.UserAuth,
		UserAvator: req.UserAvator,
	}
	res.UserAuth = req.UserAuth
	res.UserName = req.UserName
	res.UserAvator = req.UserAvator
	return dao.AddUser(u)
}

type ValidateUserInfoRequest struct {
	UserName string
	UserAuth string
}

type ValidteUserInfoResponce struct {
	Success bool
}

func ValidateUserInfo(req *ValidateUserInfoRequest, res *ValidteUserInfoResponce) error {

	u, err := dao.QueryUserByName(req.UserName)
	if err != nil {
		return err
	}
	if u.UserAuth != req.UserAuth {
		res.Success = false
	}
	res.Success = true
}
