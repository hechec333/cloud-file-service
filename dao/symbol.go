package dao

import (
	"errors"
	"orm/dao/cache"
	"orm/dao/db"
	"strconv"
)

type Symbol struct {
	Id  int `gorm:"primary_key"`
	Src int `gorm:"column:SoureceId"`
	Dst int `gorm:"column:DstId"`
}

func (Symbol) TableName() string {
	return "Symbol"
}

func symbolCacheKey(fileId, foid int) string {
	sym := Symbol{}
	return sym.TableName() + ":symbol" + ":" + strconv.Itoa(foid) + ":" + strconv.Itoa(fileId)
}
func CreateSymbol(srcId, DstId int) error {
	sym := Symbol{
		Src: srcId,
		Dst: DstId,
	}
	if err := db.DB.Create(&sym).Error; err != nil {
		return err
	}
	return cache.SetKey(symbolCacheKey(srcId, DstId), 1, 360)
}

func CreateSymbols(srcIds []int, dstId int) error {
	syms := make([]Symbol, len(srcIds))
	for i, v := range srcIds {
		syms[i] = Symbol{
			Src: v,
			Dst: dstId,
		}
	}

	if err := db.DB.Create(&syms).Error; err != nil {
		return err
	}
	keys := []interface{}{}
	for _, v := range syms {
		keys = append(keys, symbolCacheKey(v.Src, v.Dst))
		keys = append(keys, 1)
	}
	return cache.MsetKeyExpire(360, keys...)
}

func RemoveSymbol(srcId, DstId int) error {

	cache.DelKey(symbolCacheKey(srcId, DstId))
	sym := Symbol{
		Src: srcId,
		Dst: DstId,
	}
	if err := db.DB.Delete(&sym).Error; err != nil {
		return err
	}
	return nil
}

func RemoveTargetSymbols(folderId int) error {
	sym := Symbol{}
	keys, err := cache.Keys(sym.TableName() + ":" + strconv.Itoa(folderId))
	if err != nil {
		return err
	}
	cache.DelKeys(keys...)

	if err := db.DB.Where("DstId = ?", folderId).Delete(&sym).Error; err != nil {
		return err
	}
	return nil
}

func QuerySymbol(SrcId, dstId int) (bool, error) {
	_, err := cache.GetKey(symbolCacheKey(SrcId, dstId))
	if err != nil {
		if er := db.DB.Where("SourceId = ? and DstId = ?", SrcId, dstId).First(&Symbol{}).Error; er == nil {
			cache.SetKey(symbolCacheKey(SrcId, dstId), 1, 360)
			return true, er
		} else {
			return false, errors.Join(er, err)
		}
	}
	return true, err
}

func QuerySrcSymbol(srcId int) ([]Symbol, error) {
	syms := []Symbol{}

	if err := db.DB.Where("SourceId = ?", srcId).Find(&syms).Error; err != nil {
		return syms, err
	}

	keys := []interface{}{}
	for _, v := range syms {
		keys = append(keys, symbolCacheKey(v.Src, v.Dst))
		keys = append(keys, 1)
	}
	return syms, cache.MsetKeyExpire(360, keys...)
}

func QueryDstSymbol(dstId int) ([]Symbol, error) {
	syms := []Symbol{}

	if err := db.DB.Where("DstId = ?", dstId).Find(&syms).Error; err != nil {
		return syms, err
	}

	keys := []interface{}{}
	for _, v := range syms {
		keys = append(keys, symbolCacheKey(v.Src, v.Dst))
		keys = append(keys, 1)
	}
	return syms, cache.MsetKeyExpire(360, keys...)
}

func RemoveSrcSymbols(srcId int) ([]Symbol, error) {
	syms := []Symbol{}

	if err := db.DB.Where("SrcId = ?", srcId).Find(&syms).Error; err != nil {
		return syms, err
	}

	keys := []interface{}{}
	for _, v := range syms {
		keys = append(keys, symbolCacheKey(v.Src, v.Dst))
		keys = append(keys, 1)
	}
	return syms, cache.MsetKeyExpire(360, keys...)
}
