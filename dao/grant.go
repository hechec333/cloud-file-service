package dao

import (
	"encoding/json"
	"errors"
	"orm/dao/cache"
	"orm/dao/db"
	"strconv"
	"strings"
)

type Grant struct {
	GrantId    int    `gorm:"primary_key"`
	ObjectId   int    `gorm:"column:ObjectId"`
	ObjectType string `gorm:"column:ObjectType"`
	GrantType  string `gorm:"column: GrantType"`
	OwnerId    int    `gorm:"column:OwnerId"`
}

var fileMaxGrants = []string{"r", "w", "d"}
var folderMaxGrants = []string{"r", "w", "c", "d"}
var grantSeed = 0x08829943

func getNextGrantId() int {
	grantSeed++
	return grantSeed
}

func (Grant) TableName() string {
	return "Grant"
}

func (g *Grant) CacheKey() string {
	return g.TableName() + ":" + strconv.Itoa(g.GrantId)
}

func isSubset(a, b []string) bool {
	set := make(map[string]bool)
	for _, v := range b {
		set[v] = true
	}
	for _, v := range a {
		if !set[v] {
			return false
		}
	}
	return true
}

// file: r+w+d folder: r+w+c+d store: r+w+c+d
func ValidateGrantType(grant string, objectType string) bool {
	grants := strings.Split(grant, "+")
	if objectType == "folder" || objectType == "store" {

		return isSubset(grants, folderMaxGrants)
	}

	return isSubset(grants, fileMaxGrants)
}

func CreateFileGrantProps(fileId int, userId int, types string) error {

	if !ValidateGrantType(types, "file") {
		return errors.New("invalid grant props")
	}
	g := &Grant{
		GrantId:    getNextGrantId(),
		ObjectId:   fileId,
		ObjectType: "file",
		GrantType:  types,
		OwnerId:    userId,
	}
	if err := db.DB.Create(g).Error; err != nil {
		return err
	}

	jsonc, _ := json.Marshal(g)

	cache.SetKey(g.CacheKey(), string(jsonc), 360)
	return nil
}

func CreateFolderGrantProps(folderId, userId int, types string) error {
	if !ValidateGrantType(types, "folder") {
		return errors.New("invalid grant props")
	}
	g := &Grant{
		GrantId:    getNextGrantId(),
		ObjectId:   folderId,
		ObjectType: "folder",
		GrantType:  types,
		OwnerId:    userId,
	}
	if err := db.DB.Create(g).Error; err != nil {
		return err
	}

	jsonc, _ := json.Marshal(g)

	cache.SetKey(g.CacheKey(), string(jsonc), 360)
	return nil
}

func CreateStoreGrantProps(folderId, userId int, types string) error {
	if !ValidateGrantType(types, "folder") {
		return errors.New("invalid grant props")
	}
	g := &Grant{
		GrantId:    getNextGrantId(),
		ObjectId:   folderId,
		ObjectType: "folder",
		GrantType:  types,
		OwnerId:    userId,
	}
	if err := db.DB.Create(g).Error; err != nil {
		return err
	}

	jsonc, _ := json.Marshal(g)

	cache.SetKey(g.CacheKey(), string(jsonc), 360)
	return nil
}

func QueryGrantByOwnerId(ownerId int) ([]*Grant, error) {
	g := &Grant{
		OwnerId: ownerId,
	}
	result := []*Grant{}
	if err := db.DB.Where(g).Find(&result).Error; err != nil {
		return result, err
	}

	cacheReuslt := []any{}

	for _, v := range result {
		jsonc, _ := json.Marshal(v)
		cacheReuslt = append(cacheReuslt, v.CacheKey(), string(jsonc))
	}

	cache.MsetKeyExpire(360, cacheReuslt...)
	return result, nil
}

func QueryGrantByObjectId(objId int) (*Grant, error) {

	g := &Grant{
		ObjectId: objId,
	}
	result := &Grant{}
	if err := db.DB.Where(g).First(&result).Error; err != nil {
		return result, err
	}
	jsonc, _ := json.Marshal(result)
	cache.SetKey(result.CacheKey(), string(jsonc), 360)
	return result, nil
}

func QueryGrantById(id int) (*Grant, error) {
	g := &Grant{
		GrantId: id,
	}
	if jsonc, err := cache.GetKey(g.CacheKey()); err == nil {
		json.Unmarshal([]byte(jsonc), g)
		return g, nil
	}

	if err := db.DB.Where(&Grant{
		GrantId: id,
	}).First(&g).Error; err != nil {
		return g, err
	}

	jsonc, _ := json.Marshal(g)
	cache.SetKey(g.CacheKey(), string(jsonc), 360)
	return g, nil
}
