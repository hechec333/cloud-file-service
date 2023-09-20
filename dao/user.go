package dao

import (
	"encoding/json"
	"fmt"
	"orm/common/util"
	"orm/dao/cache"
	"orm/dao/db"
	"strconv"
)

type User struct {
	ID         int64  `gorm:"primary_key"`
	UserName   string `gorm:"column:UserName" json:"userName"`
	UserAvator string `gorm:"column:UserAvator" json:"userAvator"`
	UserAuth   string `gorm:"column:UserPassword" json:"userAuth"`
}

func (User) TableName() string {
	return "User"
}

func cacheKey(id int64) string {
	u := User{}
	return u.TableName() + ":" + strconv.FormatInt(id, 10)
}

func QueryUserById(id int64) (*User, error) {
	user := User{}
	var err error
	jsonc, err := cache.GetKey(user.TableName() + ":" + strconv.FormatInt(id, 10))
	if err != nil {
		fmt.Println(err)
		err = db.DB.Find(&user, id).Error
		if err != nil {
			return &user, err
		}
		bys, _ := json.Marshal(&user)
		cache.SetKey(user.TableName()+":"+strconv.FormatInt(id, 10), string(bys), 360)
		return &user, nil
	} else {

		json.Unmarshal([]byte(jsonc), &user)
		return &user, nil
	}

}

func QueryUserByName(name string) (*User, error) {
	user := User{}
	var err error
	// "UserName like ?" -> "user_name like "
	err = db.DB.Where("UserName = ?", name).First(&user).Error

	if err != nil {
		return &user, err
	}
	bys, _ := json.Marshal(&user)
	err = cache.SetKey(user.TableName()+":"+strconv.FormatInt(user.ID, 10), string(bys), 360)
	return &user, err
}

func AddUser(u User) error {
	u.ID = int64(util.Uid())
	if err := db.DB.Create(u).Error; err != nil {
		return err
	}
	jsonc, _ := json.Marshal(&u)
	err := cache.SetKey(cacheKey(u.ID), string(jsonc), 360)

	return err
}

func DelUser(id int64) error {

	cache.DelKey(cacheKey((id)))
	u := User{
		ID: id,
	}
	if err := db.DB.Where("Id = ?", id).Delete(&u).Error; err != nil {
		return err
	}
	return nil
}

func UpdateUser(u User) error {

	cache.DelKey(cacheKey(u.ID))
	if err := db.DB.Updates(&u).Error; err != nil {
		return err
	}
	jsonc, _ := json.Marshal(&u)
	cache.SetKey(cacheKey(u.ID), string(jsonc), 360)
	return nil
}
