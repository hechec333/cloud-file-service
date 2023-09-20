package dao

import (
	"encoding/json"
	"orm/dao/cache"
	"orm/dao/db"
	"strconv"
	"strings"
)

type WhiteList struct {
	ID      int `gorm:"primary_key"`
	GrantId int `gorm:"column:GrantId"`
	GuestId int `gorm:"column:GusetId"`
}

var whiteListSeed = 0x89300290

func getNextWhiteListId() int {
	whiteListSeed++
	return whiteListSeed
}

// 注意这里对缓存键的设置不在以主键设置,考虑到查询多以grantId

func (w *WhiteList) CacheKey() string {

	return w.TableName() + ":" + strconv.Itoa(w.GrantId)
}

func (WhiteList) TableName() string {
	return "WhiteList"
}
func OpenAll(grantId int) {

	w := WhiteList{
		ID:      getNextWhiteListId(),
		GrantId: grantId,
		GuestId: 0,
	}

	db.DB.Create(&w)

	jsonc, _ := json.Marshal(&w)
	cache.SetKey(w.CacheKey(), string(jsonc), 360)
}

func IsOpenToAll(w *WhiteList) bool {
	return w.GuestId == 0
}

func AddGuests(grangId int, userIds []int) error {
	ws := []*WhiteList{}

	for _, v := range userIds {
		ws = append(ws, &WhiteList{
			ID:      getNextWhiteListId(),
			GrantId: grangId,
			GuestId: v,
		})
	}

	if err := db.DB.CreateInBatches(ws, len(ws)).Error; err != nil {
		return err
	}

	cacheReuslt := []string{}

	for _, v := range ws {
		jsonc, _ := json.Marshal(v)
		cacheReuslt = append(cacheReuslt, string(jsonc))
	}
	tmp := WhiteList{
		GrantId: grangId,
	}
	cache.SetKey(tmp.CacheKey(), strings.Join(cacheReuslt, "&"), 360)

	return nil
}

func QueryGrantGuests(grantId int) ([]*WhiteList, error) {

	ws := []*WhiteList{}
	tmp := WhiteList{
		GrantId: grantId,
	}
	if raw, err := cache.GetKey(tmp.CacheKey()); err == nil {
		citems := strings.Split(raw, "&")
		for _, v := range citems {
			json.Unmarshal([]byte(v), &tmp)
			zt := tmp
			ws = append(ws, &zt)
		}
		return ws, nil
	}

	if err := db.DB.Where(&tmp).Find(ws).Error; err != nil {
		return ws, err
	}

	cacheResult := []string{}

	for _, v := range ws {
		jsonc, _ := json.Marshal(&v)
		cacheResult = append(cacheResult, string(jsonc))
	}

	cache.SetKey(tmp.CacheKey(), strings.Join(cacheResult, "&"), 360)
	return ws, nil
}

func RemoveGuest(grantId, userId int) error {
	tmp := WhiteList{
		GrantId: grantId,
	}
	if jsonc, err := cache.GetKey(tmp.CacheKey()); err == nil {
		citems := strings.Split(jsonc, "&")

		for i, v := range citems {
			zt := WhiteList{
				GrantId: grantId,
			}

			json.Unmarshal([]byte(v), &zt)

			if zt.GuestId == userId {
				if len(citems) == 1 {
					cache.DelKey(tmp.CacheKey())
				} else {
					citems = append(citems[:i], citems[i+1:]...)
					cache.SetKey(tmp.CacheKey(), strings.Join(citems, "&"), 360)
				}
				break
			}
		}
	}
	tmp.GuestId = userId
	return db.DB.Delete(&tmp).Error
}
