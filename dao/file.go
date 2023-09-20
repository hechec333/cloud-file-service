package dao

import (
	"encoding/json"
	"errors"
	"orm/common/hash"
	"orm/dao/cache"
	"orm/dao/db"
	"path"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

// CREATE TABLE `File`  (
// 	`FileId` bigint(11) NOT NULL AUTO_INCREMENT,
// 	`StoreId` int(11) NULL,
// 	`ParentFolderId` int(11) NULL,
// 	`FileName` varchar(255) NULL,
// 	`FilePath` varchar(255) NULL,
// 	`FileHash` varchar(255) NULL,
// 	`CreateTime` datetime NULL,
// 	`UpdateTime` datetime NULL ON UPDATE CURRENT_TIMESTAMP,
// 	PRIMARY KEY (`FileId`)
//   );

const file_seed = 0x887899ef
const fileIdBase = 0x7fffffff

var file_seed_var int = file_seed

func nextFileID() int {
	id := file_seed_var
	file_seed_var++
	return id & fileIdBase
}

func isFileID(id int) bool {
	return id&0x80000000 != 0x80000000
}

func FileCacheKey(id int) string {
	f := File{}
	return f.TableName() + ":" + strconv.Itoa(id)
}

type File struct {
	FileId     int       `gorm:"primary_key" json:"fileId"`
	StoreId    int       `gorm:"column:StoreId" json:"storeId"`
	ParentId   int       `gorm:"column:ParentFolderId" json:"parentId"`
	FileName   string    `gorm:"column:FileName" json:"fileName"`
	FileHash   string    `gorm:"column:FileHash" json:"-"` //存储层定位符，不需要暴露给前端
	FileType   string    `gorm:"column:FileType" json:"-"` //存储类型 fixedable appendable
	FileSize   int64     `gorm:"column:FileSize" json:"fileSize"`
	CreateTime time.Time `gorm:"column:CreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:UpadetTime" json:"updateTime"`
}

type ReplaceFunc = func(*File) error

func (File) TableName() string {
	return "File"
}

func GenFileHash(f *File) {

	f.FileHash = hash.FileHash(f.StoreId, f.FileId, path.Base(f.FileName))
}

// 检查父目录是否存在
// 检查是否有重名文件
func ValidateFile(f File) (bool, error) {
	dir, err := QueryDirById(f.ParentId)
	if err != nil {
		return false, err
	}
	files, _ := QueryFileList(dir.ID)
	for _, v := range files {
		if v.FileName == f.FileName {
			return false, errors.New(v.FileName + " already exitst!")
		}
	}

	return true, nil
}

func CreateFile(f *File) error {
	f.FileId = nextFileID()
	GenFileHash(f)
	is, msg := ValidateFile(*f)
	if !is {
		return msg
	}
	dir, _ := QueryDirById(f.ParentId)

	RelaseMountDir(dir)

	if err := db.DB.Create(f).Error; err != nil {
		return err
	}
	jsonc, _ := json.Marshal(f)
	return cache.SetKey(FileCacheKey(f.FileId), string(jsonc), 360)
}

func BindUpdateFileSymbol(oldf *File, newf *File) error {
	syms, _ := QuerySrcSymbol(oldf.FileId)
	if len(syms) >= 1 {
		// 1.生成另一个实例代替本地实例
		ff := *newf
		ff.FileId = getNextID()
		//GenFileHash(&ff) 不能顺便改filehash，任何查询不得跳过后端直接查询oss，可能oss还没有新建的这个实例，因此如果不修改内容
		db.DB.Create(&ff)
		jsonc, _ := json.Marshal(ff)

		cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)
		// 2.修改本地实例为挂载文件
		oldf.ParentId = syms[0].Dst //将这个挂载文件的父级改成第二个目录
		db.DB.Where(&File{
			FileId: oldf.FileId,
		}).UpdateColumns(&oldf)
		jsonx, _ := json.Marshal(oldf)
		cache.SetKey(FileCacheKey(oldf.FileId), string(jsonx), 360)
	} else {
		db.DB.Where(&File{
			FileId: oldf.FileId,
		}).UpdateColumns(&newf)
		jsonx, _ := json.Marshal(newf)
		cache.SetKey(FileCacheKey(newf.FileId), string(jsonx), 360)
	}
	return nil
}

// return true to remove file in repo
func RemoveFile(f File) (bool, error) {
	dir, _ := QueryDirById(f.ParentId)
	RelaseMountDir(dir)

	syms, _ := QuerySrcSymbol(f.FileId)
	if len(syms) == 0 {
		cache.DelKey(FileCacheKey(f.FileId))
		db.DB.Where(&File{
			FileId: f.FileId,
		}).Delete(&File{})

		return true, nil
	} else {
		ff := f
		ff.ParentId = syms[0].Dst
		cache.DelKey(FileCacheKey(f.FileId))
		if err := db.DB.Where("ID = ?", ff.FileId).UpdateColumns(&ff).Error; err != nil {
			return false, err
		}
		RemoveSymbol(f.FileId, syms[0].Dst)

		return false, nil
	}
}

// 返回的文件对象如果不为空，则在这个文件对象持有之前的拷贝
func AppendFile(f *File, size int64) (*File, error) {
	cache.DelKey(FileCacheKey(f.FileId))
	syms, _ := QuerySrcSymbol(f.FileId)
	if len(syms) >= 1 {
		ff := *f
		ff.FileId = getNextID()
		ff.FileSize += size
		//ff.ParentId = syms[0].Dst
		GenFileHash(&ff)
		db.DB.Create(&ff)
		jsonc, _ := json.Marshal(ff)
		cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)

		f.ParentId = syms[0].Dst
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)

		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)
		return &ff, nil
	} else {
		f.FileSize += size
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)
		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)
	}
	return nil, nil
}

// 如果返回 false,说明本节点已经被挂载，新写入的内容应该新开slot写入
func OverWriteFile(f *File) (bool, error) {
	cache.DelKey(FileCacheKey(f.FileId))
	syms, _ := QuerySrcSymbol(f.FileId)

	if len(syms) >= 1 {
		ff := *f
		ff.FileId = getNextID()
		db.DB.Create(&ff)
		jsonc, _ := json.Marshal(ff)
		cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)

		f.ParentId = syms[0].Dst //将这个挂载文件的父级改成第二个目录
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)
		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)

		*f = ff //将本届点的改动返回
		return false, nil
	}
	return true, nil
}

func IncrFileSize(f *File, size int64) error {
	cache.DelKey(FileCacheKey(f.FileId))
	syms, _ := QuerySrcSymbol(f.FileId)
	if len(syms) >= 1 {
		// 1.生成另一个实例代替本地实例
		ff := f
		ff.FileId = getNextID()
		ff.FileSize = size
		//GenFileHash(&ff) 不能顺便改filehash，任何查询不得跳过后端直接查询oss，
		db.DB.Create(&ff)
		jsonc, _ := json.Marshal(ff)

		cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)
		// 2.修改本地实例为挂载文件
		f.ParentId = syms[0].Dst //将这个挂载文件的父级改成第二个目录
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)
		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)
	} else {
		f.FileSize = size
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)
		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)
	}
	return nil
}

// 修改操作
func RenameFile(f File, name string) error {

	cache.DelKey(FileCacheKey(f.FileId))
	syms, _ := QuerySrcSymbol(f.FileId)
	if len(syms) >= 1 {
		// 1.生成另一个实例代替本地实例
		ff := f
		ff.FileId = getNextID()
		ff.FileName = name
		//GenFileHash(&ff) 不能顺便改filehash，任何查询不得跳过后端直接查询oss，
		db.DB.Create(&ff)
		jsonc, _ := json.Marshal(ff)

		cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)
		// 2.修改本地实例为挂载文件
		f.ParentId = syms[0].Dst //将这个挂载文件的父级改成第二个目录
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)
		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)
	} else {
		f.FileName = name
		db.DB.Where(&File{
			FileId: f.FileId,
		}).UpdateColumns(&f)
		jsonx, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(f.FileId), string(jsonx), 360)
	}
	return nil
}

func CopyFile(parentId int, f File) error {
	dir, err := QueryDirById(parentId)
	if err != nil {
		return err
	}
	RelaseMountDir(dir)

	return CreateSymbol(f.FileId, dir.ID)
}

func CloneFile(parentId int, f File) error {
	dir, err := QueryDirById(parentId)
	if err != nil {
		return err
	}
	RelaseMountDir(dir)

	ff := f
	ff.FileId = getNextID()
	nameNoSuffix := path.Ext(ff.FileName)
	ff.ParentId = parentId
	ff.FileHash = hash.FileHash(ff.StoreId, ff.FileId, nameNoSuffix)
	if err := db.DB.Create(&ff).Error; err != nil {
		return err
	}

	jsonc, _ := json.Marshal(ff)
	cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)
	return nil
}

// 并不会改变repo中的位置
func MoveFile(parentId int, f File) (*File, error) {
	dir, err := QueryDirById(parentId)
	if err != nil {
		return nil, err
	}
	RelaseMountDir(dir)
	ff := f
	ff.ParentId = parentId
	cache.DelKey(FileCacheKey(f.FileId))
	if err := db.DB.Where(&File{
		FileId: f.FileId,
	}).UpdateColumns(&ff).Error; err != nil {
		return nil, err
	}

	jsonc, _ := json.Marshal(f)

	return &ff, cache.SetKey(FileCacheKey(ff.FileId), string(jsonc), 360)
}

func QueryFileById(id int) (File, error) {
	result, err := cache.GetKey(FileCacheKey(id))
	f := File{}
	if err != nil {
		if err = db.DB.Where(&File{
			FileId: id,
		}).Find(&f).Error; err != nil {
			return f, err
		}

		jsonc, _ := json.Marshal(f)
		cache.SetKey(FileCacheKey(id), string(jsonc), 360)
		return f, nil
	}

	json.Unmarshal([]byte(result), &f)
	return f, nil
}

func QueryFileByPath(storeId int, path string) (File, error) {
	paths := strings.Split(path, "/")
	f := File{}
	dir, err := QueryDirByPath(storeId, strings.Join(paths[:len(paths)-1], "/"))
	if err != nil {
		return f, err
	}

	childs, _ := QueryFileList(dir.ID)
	for _, v := range childs {
		if v.FileName == paths[len(paths)-1] {
			return v, nil
		}
	}

	return f, gorm.ErrRecordNotFound
}
