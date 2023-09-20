package service

import (
	"errors"
	"orm/common/jwtx"
	"orm/config"
	"orm/dao"
	"orm/reposity"
)

type CreateStoreRequest struct {
	Token    string `json:"-"`
	Limits   int    `json:"limits"`
	Name     string `json:"name"`
	Persiter string `json:"persiter"`
}

type CreateStoreResponce struct {
	StoreId int
}

func CreateStore(req *CreateStoreRequest, res *CreateStoreResponce) error {
	conf := config.GetConfig()
	var err error
	var uid int64
	if uid, err = jwtx.ParseToken(req.Token, conf.Secret.AccessSecret); err != nil {
		return err
	}

	if !reposity.ValidatePersiter(req.Persiter) {
		return errors.New("invalid repositer type")
	}

	store, err := dao.CreateStore(int(uid), req.Name, int64(req.Limits), req.Persiter)

	res.StoreId = store.ID
	return err
}

type GetStoreInfoRequest struct {
	StoreId int `json:"storeId"`
}

type GetStoreInfoResponce struct {
	dao.Store
}

func GetStoreInfo(req *GetStoreInfoRequest, res *GetStoreInfoResponce) (err error) {
	res.Store, err = dao.QueryStoreInfo(req.StoreId)
	return
}

type SetStoreInfoRequest struct {
	Token   string `json:"-"`
	StoreId int    `json:"storeId"`
	Limits  int    `json:"limits"`
	Name    string `json:"name"`
}

func SetStoreInfo(req *SetStoreInfoRequest) error {

	return dao.SetStore(dao.Store{
		Limits: req.Limits,
		Name:   req.Name,
	})
}

type DelStoreInfoRequest struct {
	Token   string
	StoreId int
}

func DelStore(req *DelStoreInfoRequest) error {
	return nil
}

type GetUserStoresRequest struct {
	UserId int
}
type GetUserStoresResponce struct {
	Data []dao.Store
}

func GetUserStores(req *GetUserStoresRequest, res *GetUserStoresResponce) error {
	var err error
	res.Data, err = dao.QueryUserStoreInfo(req.UserId)

	return err
}
