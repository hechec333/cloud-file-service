package dao

import (
	"encoding/json"
	"errors"
	"orm/common/page"
	"orm/dao/cache"
	"orm/dao/db"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
)

// 实现下列写时复制功能的方法还有另一种

// CREATE TABLE `FileFolder`  (

//		`ID` int(11) NOT NULL AUTO_INCREMENT,
//		`StoreId` int(11) NULL,
//		`FileFolderPath` varchar(255) NULL,
//		`FileFolderName` varchar(255) NULL,
//		`ParentFolderId` int(11) NULL,
//		`CreateTime` datetime NULL,
//		`UpdateTime` datetime NULL ON UPDATE CURRENT_TIMESTAMP,
//		PRIMARY KEY (`ID`)
//	  );
type Folder struct {
	ID         int       `gorm:"primary_key" json:"-"`
	StoreId    int       `gorm:"column:StoreId" json:"storeId"` // forigen key -> Store#ID
	Name       string    `gorm:"column:FileFolderName" json:"name"`
	ParentId   int       `gorm:"column:ParentFolderId" json:"parentId"` //不是外键
	CreateTime time.Time `gorm:"column:CreateTime" json:"createTime"`
	UpdateTime time.Time `gorm:"column:UpdateTime" json:"updateTime"`
}

func (Folder) TableName() string {
	return "Folder"
}

func (f Folder) cacheIdKey() string {
	return f.TableName() + ":" + strconv.Itoa(f.ID)
}

func folderCacheKey(id int) string {
	f := Folder{}
	return f.TableName() + ":" + strconv.Itoa(id)
}

const seed int = 0x775889ef
const folderIdBase int = 0x10000000

var Folder_ID = seed

func IsRootDir(f Folder) bool {
	return f.ParentId == 0 && f.Name == ""
}

func getNextID() int {
	ID := Folder_ID
	Folder_ID++
	return ID | folderIdBase
}

func isFolderID(id int) bool {
	return id&0x80000000 == 0x80000000
}

func isFolderNotFoundErr(err error) bool {
	return err.Error() == gorm.ErrRecordNotFound.Error()
}

// 检查文件命名是否冲突
func ValidteDir(f Folder) error {
	// 如果父节点有挂载情况
	// 如果当前父节点有指向其他节点，则需判断新创建的节点和被指向的节点是否命名是否冲突，文件和文件夹不可以重名
	dirs, _ := QueryFolderList(f.ParentId)
	for _, v := range dirs {
		if v.Name == f.Name {
			return errors.New("dir already exitst")
		}
	}
	return nil
}

// 要考虑沿路上所有的挂载点
// 比较耗费时间，比较考虑是否有必要调用
func RelaseMountDir(f Folder) {
	//是否是根节点
	if IsRootDir(f) {
		return
	}
	//所有以本节点为src的挂载点
	syms, _ := QuerySrcSymbol(f.ParentId)
	dir, _ := QueryDirById(f.ParentId)
	// 如果被指向，需要复原所有指向节点的父节点
	if len(syms) != 0 {
		for _, v := range syms {
			src, _ := QueryDirById(v.Dst)
			RelaseMountDir(src) // 处理源路径上的挂载地点
			CloneDir(v.Dst, v.Src)
		}
		RemoveSrcSymbols(dir.ID)
	}
	RelaseMountDir(dir) //继续向上遍历
}

// 不处理挂载地点
func CloneDir(dstId int, srcId int) {
	src, _ := QueryDirById(srcId)
	//	dst, _ := QueryDirById(dstId)
	src.ID = getNextID()
	src.ParentId = dstId
	src.CreateTime = time.Now()
	db.DB.Create(&src)
	jsonc, _ := json.Marshal(src)
	cache.SetKey(src.cacheIdKey(), string(jsonc), 360)
	ff, _ := QueryFolderList(dstId)
	for _, f := range ff {
		CloneDir(src.ID, f.ID)
	}
	fs, _ := QueryFileList(dstId)
	for _, v := range fs {
		//仅调整一个link标志，等到真正对文件进行修改的时候才进行clone操作
		//CopyFile(dstId, v)
		CreateSymbol(v.FileId, dstId)
	}

}

// 如果路径不存在,返回error
// storeid name parentid createTime ID #required
func CreateDir(f Folder) (int, error) {
	RelaseMountDir(f)
	f.ID = getNextID()
	if err := ValidteDir(f); err != nil {
		return f.ID, errors.Join(errors.New("Fail to CreateDir"), err)
	}
	if err := db.DB.Create(&f).Error; err != nil {
		return f.ID, err
	}
	jsonc, _ := json.Marshal(f)
	cache.SetKey(f.cacheIdKey(), string(jsonc), 360)
	return f.ID, nil
}

func CreateRootDir(storeId int) (Folder, error) {
	f := Folder{
		ID:         getNextID(),
		Name:       "",
		ParentId:   0,
		StoreId:    storeId,
		CreateTime: time.Now(),
	}

	if err := db.DB.Create(&f).Error; err != nil {
		return f, err
	}

	jsonc, _ := json.Marshal(&f)

	cache.SetKey(folderCacheKey(f.ID), string(jsonc), 360)
	return f, nil
}

func CreateRecursiveDir(ff []Folder) error {
	RelaseMountDir(ff[0])
	if err := ValidteDir(ff[0]); err != nil {
		return err
	}

	if err := db.DB.Create(&ff).Error; err != nil {
		return err
	}
	strs := []string{}
	for _, v := range ff {
		jsonc, _ := json.Marshal(v)
		strs = append(strs, v.cacheIdKey(), string(jsonc))
	}
	return cache.MsetKeyExpire(360, strs)
}

// 在一个目录下创建多个子目录
func CreateDirs(ff []Folder) int {
	RelaseMountDir(ff[0])
	ffs := make([]Folder, len(ff)) //合法的dir个数
	copy(ffs, ff)
	for i, v := range ff {
		if ValidteDir(v) != nil {
			ffs = append(ffs[:i], ffs[i+1:]...)
		}
	}
	raws := db.DB.Create(&ffs).RowsAffected
	strs := []string{}
	for _, v := range ffs {
		jsonc, _ := json.Marshal(v)
		strs = append(strs, v.cacheIdKey(), string(jsonc))
	}

	cache.MsetKeyExpire(360, strs)
	return int(raws)
}

// 如果路径不存在就创建，如果遭遇错误就panic
// 需要传入 path name
func MustCreateDir(storeId int, path string) Folder {
	pathItems := strings.Split(path, "/")
	lastDir := Folder{}
	// 寻找到第一个不存在的目录
	for i := 0; i < len(pathItems); i++ {
		ff, err := QueryDirByPath(storeId, strings.Join(pathItems[i:], "/"))
		if err == nil {
			lastDir = ff
			continue
		} else if isFolderNotFoundErr(err) {
			folders := make([]Folder, len(pathItems)-i)
			for index := 0; index < len(pathItems)-i; index++ {
				folders[index] = Folder{
					StoreId:    storeId,
					ID:         getNextID(),
					ParentId:   lastDir.ID,
					Name:       pathItems[index+i],
					CreateTime: time.Now(),
				}
				lastDir = folders[index]
			}
			err := CreateRecursiveDir(folders)
			if err != nil {
				panic(err)
			}
			return folders[len(folders)-1]
		} else {
			panic(err)
		}
	}
	return lastDir
}
func RemoveDir(f Folder) error {
	// 处理当前目录和父目录以上的挂载情况
	RelaseMountDir(f)
	return removeDir(f)
}

// 所有文件计数减1 ，如果文件计数为0 则删除
func removeDir(f Folder) error {
	//查看当前父节点是否被挂载
	//RelaseMountDir(f)
	// 当前目录可能有链接
	// 当前目录可能被链接
	// 当前目录的子目录有可能挂载
	folders, _ := QueryFolderList(f.ID)
	for _, v := range folders {
		if is, _ := QuerySymbol(v.ID, f.ID); is {
			// 如果当前目录是挂载的子目录，移除这份挂载关系
			RemoveSymbol(v.ID, f.ID)
			continue
		}
		syms, _ := QuerySrcSymbol(v.ID)
		if len(syms) == 1 {
			cache.DelKey(f.TableName() + ":" + strconv.Itoa(v.ID))
			err := db.DB.Where("ID = ?", v.ID).Update("ParentFolderId", syms[0].Dst).Error
			if err != nil {
				return err
			}
			continue
		} else if len(syms) >= 1 {
			for _, v := range syms {
				CloneDir(v.Dst, v.Src)
			}
		}
		err := removeDir(v)
		if err != nil {
			return err
		}
		db.DB.Delete(&v)
	}

	// 处理文件删除

	files, _ := QueryFileList(f.ID)
	for _, v := range files {
		RemoveFile(v)
	}
	return nil
}

// 拷贝到目标路径
// 源目录必须存在
// path是目标要复制的目录
// f下的所有文件计数增1，
func CopyDir(f Folder, id int) error {
	dir, err := QueryDirById(id)
	RelaseMountDir(dir)
	if err != nil {
		return err
	}
	return CreateSymbol(f.ID, dir.ID)
}

// 如果目前不存在就创建到目标路径
// 源目录必须存在
// path仅需给到目标父目录
// if error -> panic(error)
// f.path => /etc/hosts path=> /root <==> create /root/hosts
func MustCopyDir(f Folder, storeId int, path string) Folder {
	dir, err := QueryDirByPath(storeId, path)
	if err != nil && !isFolderNotFoundErr(err) {
		panic(err)
	}
	ff := Folder{}
	if isFolderNotFoundErr(err) {
		ff = MustCreateDir(storeId, path+"/"+f.Name)
	} else {
		ff = dir
	}

	err = CreateSymbol(f.ID, ff.ID)
	if err != nil {
		panic(err)
	}
	// copy all files

	fs := f
	fs.ParentId = dir.ID
	return fs
}

func RenameFolder(f Folder, newName string) error {
	RelaseMountDir(f)
	cache.DelKey(f.cacheIdKey())
	f.Name = newName
	if err := db.DB.Where("ID = ?", f.ID).UpdateColumns(&f).Error; err != nil {
		return err
	}
	jsonc, _ := json.Marshal(f)
	return cache.SetKey(f.cacheIdKey(), string(jsonc), 360)
}

// 移动路径
func MoveDir(oldf Folder, id int) error {
	dir, err := QueryDirById(id)
	if err != nil {
		return err
	}
	//处理 oldf以上的挂载点，不处理oldf往下的挂载点
	RelaseMountDir(oldf)
	cache.DelKey(oldf.TableName() + ":" + strconv.Itoa(oldf.ID))
	if err := db.DB.Where("ID = ?", oldf.ID).Update("ParentFolderId", dir.ID).Error; err != nil {
		return err
	}
	return nil
}

// 如果新路径不存在则创建,如果遇到其他错误就panic
func MustMoveDir(oldf Folder, path string) Folder {

	dir, err := QueryDirByPath(oldf.StoreId, path)
	if err != nil && !isFolderNotFoundErr(err) {
		panic(err)
	} else if err == nil {
		err := MoveDir(oldf, dir.ID)
		if err != nil {
			panic(err)
		}
	} else {
		MustCreateDir(oldf.StoreId, path)
		err := MoveDir(oldf, dir.ID)
		if err != nil {
			panic(err)
		}
	}

	f := oldf
	f.ParentId = dir.ID
	return f
}

// 只会缓存一级
func QueryDirByPath(storeId int, path string) (Folder, error) {
	pathItems := strings.Split(path, "/")
	f := Folder{}
	db.DB.Where("ParentFolderId = ? and StoreId = ?", 0, storeId).Find(&f)
	i := 0
	for {
		ff := []Folder{}
		if err := db.DB.Where("ParentFoldeId = ? and StoreId = ?", f.ID, storeId).Find(&ff).Error; err != nil {
			return Folder{}, err
		}
		symbols, _ := QuerySrcSymbol(f.ID)
		for _, v := range symbols {
			fd, _ := QueryDirById(v.Src)
			ff = append(ff, fd)
		}
		i++
		//路径1,本子树上
		for _, v := range ff {
			if v.Name == pathItems[i] {
				jsonc, _ := json.Marshal(v)
				cache.SetKey(v.cacheIdKey(), string(jsonc), 360)
				if i == len(pathItems)-1 {
					return v, nil
				}
				f = v
				continue
			}
		}

		return Folder{}, gorm.ErrRecordNotFound
	}

}

func QueryDirById(id int) (Folder, error) {
	f := Folder{}
	jsonc, err := cache.GetKey(folderCacheKey(id))
	if err != nil {

		err = db.DB.Where("ID = ?", id).First(&f).Error
		if err != nil {
			return f, err
		}
	}

	json.Unmarshal([]byte(jsonc), &f)
	return f, nil
}

func QueryFileList(id int) ([]File, error) {
	files := []File{}
	if err := db.DB.Where(&File{
		ParentId: id,
	}).Find(&files).Error; err != nil {
		return files, err
	}

	syms, _ := QueryDstSymbol(id)
	for _, v := range syms {
		if isFileID(v.Src) {
			f, _ := QueryFileById(v.Src)
			f.ParentId = v.Dst
			files = append(files, f)
		}
	}
	strs := []interface{}{}
	for _, v := range files {
		jsonc, _ := json.Marshal(v)
		strs = append(strs, FileCacheKey(v.FileId), string(jsonc))
	}

	cache.MsetKeyExpire(360, strs...)
	return files, nil
}

// 返回的结果包括所有挂载点的目录
func QueryFolderList(parentid int) ([]Folder, error) {

	folders := []Folder{}
	if err := db.DB.Where("ParentFolderId = ?", parentid).Find(&folders).Error; err != nil {
		return folders, err
	}
	symbolDst, _ := QueryDstSymbol(parentid)
	for _, v := range symbolDst {
		if isFolderID(v.Src) {
			ff, _ := QueryDirById(v.Src)
			ff.ParentId = v.Dst
			folders = append(folders, ff)
		}
	}

	// 不用考虑此节点以上节点有挂载情况

	strs := []interface{}{}
	for _, v := range folders {
		jsonc, _ := json.Marshal(v)
		strs = append(strs, v.cacheIdKey(), string(jsonc))
	}

	cache.MsetKeyExpire(360, strs...)

	return folders, nil
}

// 包含文件夹
func QueryFolderPage(parentId int, p page.Pagger) ([]Folder, error) {
	ff := []Folder{}
	if p.PageOrder == page.DefaultOrderOption {
		if err := db.DB.Scopes(page.DefaultOrderPageWarrper(p)).
			Where(&Folder{ParentId: parentId}).Find(&ff).Error; err != nil {
			return ff, err
		}
	}
	return ff, nil
}

func QueryFilePage(parentId int, p page.Pagger) ([]File, error) {
	fs := []File{}
	if p.PageOrder == page.DefaultOrderOption {
		if err := db.DB.Scopes(page.DefaultOrderPageWarrper(p)).
			Where(&File{ParentId: parentId}).Find(&fs).Error; err != nil {
			return fs, err
		}
	}
	return fs, nil
}

// return folderNum fileNum  error
func QueryRows(parentId int) (int64, int64, error) {
	folders := int64(0)
	err := db.DB.Where(&Folder{
		ParentId: parentId,
	}).Count(&folders).Error
	if err != nil {
		return folders, 0, err
	}

	files := int64(0)

	err = db.DB.Where(&File{
		ParentId: parentId,
	}).Count(&files).Error
	return folders, files, err
}

func DumpTree(f Folder) ([]interface{}, error) {
	tree := []interface{}{}

	folders, err := QueryFolderList(f.ID)
	if err != nil {
		return tree, err
	}
	files, err := QueryFileList(f.ID)

	if err != nil {
		return tree, err
	}
	for _, v := range files {
		tree = append(tree, v)
	}

	for _, v := range folders {
		dir, err := DumpTree(v)
		if err != nil {
			return tree, err
		}
		tree = append(tree, dir)
	}
	return tree, nil
}
