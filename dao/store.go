package dao

import (
	"encoding/json"
	"orm/common/util"
	"orm/dao/cache"
	"orm/dao/db"
	"strconv"
	"time"
)

type Store struct {
	ID         int       `gorm:"primary_key" json:"-"`
	CurrentUse int       `gorm:"column:CurrentUse" json:"currentUse"`
	Name       string    `gorm:"column:Name" json:"name"`
	Limits     int       `gorm:"column:Limits" json:"limtis"`
	UserId     int       `gorm:"column:UserId" json:"userId"`
	Persiter   string    `gorm:"column:Persiter" json:"persiter"`
	CreateTime time.Time `gorm:"column:CreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:UpdateTime" json:"updateTime"`
}

func (Store) TableName() string {
	return "Store"
}
func storeCacheKey(id int) string {
	s := Store{}
	return s.TableName() + ":" + strconv.Itoa(id)
}

func CreateStore(userid int, name string, limits int64, persiter string) (Store, error) {
	st := Store{
		ID:         int(util.Uid()),
		CurrentUse: 0,
		Name:       name,
		Limits:     int(limits),
		CreateTime: time.Now(),
		Persiter:   persiter,
		UpdateTime: time.Now(),
	}
	if err := db.DB.Create(&st).Error; err != nil {
		return st, err
	}
	jsonc, _ := json.Marshal(st)
	return st, cache.SetKey(storeCacheKey(st.ID), string(jsonc), 360)
}

func SetStore(s Store) error {
	cache.DelKey(storeCacheKey(s.ID))

	if err := db.DB.Select("CurrentUse", "Limits").
		Where("ID = ?", s.ID).
		Updates(&s).Error; err != nil {
		return err
	}
	json, _ := json.Marshal(s)
	return cache.SetKey(storeCacheKey(s.ID), string(json), 360)
}

func IncrStoreUsage(st Store, incr int) error {
	cache.DelKey(storeCacheKey(st.ID))
	st.CurrentUse += incr
	if err := db.DB.Select("CurrentUse").
		Where("ID = ?", st.ID).
		Updates(&st).Error; err != nil {
		return err
	}
	json, _ := json.Marshal(st)
	return cache.SetKey(storeCacheKey(st.ID), string(json), 360)
}

func QueryStoreInfo(storeId int) (Store, error) {
	st := Store{}
	jsonc, err := cache.GetKey(storeCacheKey(storeId))
	if err == nil {
		json.Unmarshal([]byte(jsonc), &st)
		return st, nil
	} else {
		if err := db.DB.Where(&Store{
			ID: storeId,
		}).First(&st).Error; err != nil {
			return st, err
		}
		jsonx, _ := json.Marshal(&st)
		return st, cache.SetKey(storeCacheKey(st.ID), string(jsonx), 360)
	}
}

func QueryUserStoreInfo(userid int) ([]Store, error) {

	key := Store{}.TableName() + ":UserId::" + strconv.Itoa(userid)
	users := []Store{}
	if jsonc, err := cache.GetKey(key); err == nil {
		ids := []int{}
		json.Unmarshal([]byte(jsonc), &ids)
		u := make([]Store, len(ids))
		for i, id := range ids {
			result, _ := cache.GetKey(cacheKey(int64(id)))
			json.Unmarshal([]byte(result), &u[i])
		}
		return u, nil
	}

	if err := db.DB.Where("UserId = ?", userid).Find(&users).Error; err != nil {
		return users, err
	}
	ids := []int{}
	for _, v := range users {
		ids = append(ids, v.ID)
		jsonc, _ := json.Marshal(v)
		cache.SetKey(cacheKey(int64(v.ID)), string(jsonc), 360)
	}
	jsonc, _ := json.Marshal(ids)
	cache.SetKey(Store{}.TableName()+":UserId::"+strconv.Itoa(userid), string(jsonc), 360)

	return users, nil
}
