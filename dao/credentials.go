package dao

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	u "net/url"
	"orm/dao/cache"
	"orm/dao/db"
	"time"
)

type Credentials struct {
	ClienId      string `gorm:"primary_key"`
	ClientSecret string `gorm:"column:ClientSecret"`
	RedirectUrl  string `gorm:"column:RedirectUrl"`
}

func (c *Credentials) CacheKey() string {
	return c.TableName() + ":" + c.ClienId
}

func (Credentials) TableName() string {
	return "Credentials"
}

func CreateCredentialsPair(clientId, redirectUrl string) (*Credentials, error) {

	if _, err := u.Parse(redirectUrl); err != nil {
		return nil, err
	}

	hash := sha256.New()
	//输入数据
	hash.Write([]byte(time.Now().String() + clientId))
	//计算哈希值
	bytes := hash.Sum(nil)
	//将字符串编码为16进制格式,返回字符串
	secret := hex.EncodeToString(bytes)
	//返回哈希值
	cr := &Credentials{
		ClienId:      clientId,
		ClientSecret: secret,
		RedirectUrl:  redirectUrl,
	}
	if err := db.DB.Create(cr).Error; err != nil {
		return nil, err
	}
	jsonc, _ := json.Marshal(cr)

	cache.SetKey(cr.CacheKey(), string(jsonc), 360)
	return cr, nil
}

func QueryCredentialsByClientId(clientId string) (*Credentials, error) {

	cr := &Credentials{
		ClienId: clientId,
	}

	if jsonc, err := cache.GetKey(cr.CacheKey()); err == nil {
		json.Unmarshal([]byte(jsonc), cr)
		return cr, nil
	}

	if err := db.DB.Where(&Credentials{
		ClienId: clientId,
	}).First(cr).Error; err != nil {
		return nil, err
	}

	jsonc, _ := json.Marshal(cr)

	cache.SetKey(cr.CacheKey(), string(jsonc), 360)

	return cr, nil
}
