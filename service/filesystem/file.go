package service

import (
	"errors"
	"io"
	"orm/dao"
	"orm/reposity"
	"strings"
	"time"

	"gorm.io/gorm"
)

type CreateFileRequest struct {
	FileName   string
	FileSuffix string
	FileType   string
	ParentId   int
	Size       int64
	R          io.Reader
}

type CreateFileResponce struct {
	dao.File
}

// create?parentId=
func CreateFile(req *CreateFileRequest, res *CreateFileResponce) error {

	dir, err := dao.QueryDirById(req.ParentId)
	if err != nil {
		return err
	}
	f := dao.File{
		CreateTime: time.Now(),
		FileName:   req.FileName + req.FileSuffix,
		ParentId:   dir.ID,
		StoreId:    dir.StoreId,
		FileType:   req.FileType,
		FileSize:   req.Size,
	}
	err = dao.CreateFile(&f)
	if err != nil {
		return err
	}
	dao.GenFileHash(&f)
	res.File = f
	return reposity.CreateFile(&f, req.R)
}

type CreateFile2Requeset struct {
	FileName   string
	FileSuffix string
	FileType   string
	StoreId    int
	ParentPath string
	Mode       string
	Size       int64
	R          io.Reader
}

func CreateFile2(req *CreateFile2Requeset, res *CreateFileResponce) {
	dir, err := dao.QueryDirByPath(req.StoreId, req.ParentPath)
	if err != nil {
		if err == gorm.ErrRecordNotFound && req.Mode == "force" {
			dir = dao.MustCreateDir(req.StoreId, req.ParentPath)
		} else {
			panic(err)
		}
	}

	f := dao.File{
		CreateTime: time.Now(),
		FileName:   req.FileName + req.FileSuffix,
		ParentId:   dir.ID,
		StoreId:    dir.StoreId,
		FileType:   req.FileType,
		FileSize:   req.Size,
	}
	err = dao.CreateFile(&f)
	if err != nil {
		panic(err)
	}
	dao.GenFileHash(&f)
	res.File = f

	err = reposity.CreateFile(&f, req.R)
	if err != nil {
		panic(err)
	}
}

type CopyFileRequest struct {
	FileId   int
	ParentId int
}

type CopyFileResponce struct {
	dao.File
}

func CopyFile(req *CopyFileRequest, res *CopyFileResponce) error {

	dir, err := dao.QueryFileById(req.FileId)
	if err != nil {
		return err
	}

	err = dao.CopyFile(req.ParentId, dir)
	if err != nil {
		return err
	}

	res.File, err = dao.QueryFileById(req.FileId)
	res.File.ParentId = req.ParentId
	return err
}

type CopyFile2Request struct {
	FileId  int
	StoreId int
	Path    string
	Mode    string
}

type CopyFile2Responce struct {
	dao.File
}

func CopyFile2(req *CopyFile2Request, res *CopyFile2Responce) {
	dir, err := dao.QueryDirByPath(req.StoreId, req.Path)
	if err != nil {
		if err == gorm.ErrRecordNotFound && req.Mode == "force" {
			dir = dao.MustCreateDir(req.StoreId, req.Path)
		} else {
			panic(err)
		}
	}
	f, err := dao.QueryFileById(req.FileId)
	if err != nil {
		panic(err)
	}
	err = dao.CopyFile(dir.ID, f)
	if err != nil {
		panic(err)
	}
	f.ParentId = dir.ID
	f.StoreId = req.StoreId
	res.File = f
}

type AppendFileRequest struct {
	FileId int
	Size   int64
	r      io.Reader
}

type AppendFileResponce struct {
	dao.File
}

func AppendFile(req *AppendFileRequest, res *AppendFileResponce) error {

	file, err := dao.QueryFileById(req.FileId)
	if err != nil {
		return err
	}
	prev := file.FileSize
	newf, err := dao.AppendFile(&file, req.Size)

	if err != nil {
		return err
	}
	// oldf 不为nil，先拷贝一份
	if newf != nil {
		err = reposity.CopyFile(&file, newf)
		if err != nil {
			return err
		}
		_, err = reposity.AppendFile(newf, prev, req.r)
		return err
	}
	_, err = reposity.AppendFile(&file, prev, req.r)

	return err
}

type RemoveFileRequest struct {
	FileId int
}

func RemoveFile(req *RemoveFileRequest) error {
	file, err := dao.QueryFileById(req.FileId)
	if err != nil {
		return err
	}
	is, err := dao.RemoveFile(file)
	if err != nil {
		return err
	}

	if is {
		err = reposity.DelFile(file)
		if err != nil {
			return err
		}
	}

	return nil
}

type MoveFileRequest struct {
	FileId   int
	ParentId int
}

type MoveFileResponce struct {
	dao.File
}

func MoveFile(req *MoveFileRequest, res *MoveFileResponce) error {
	file, err := dao.QueryFileById(req.FileId)
	if err != nil {
		return err
	}

	f, err := dao.MoveFile(req.ParentId, file)

	res.File = *f
	return err
}

// path : /root/home/var.log
type MoveFile2Request struct {
	FileId  int
	StoreId int
	Path    string
	Mode    string
}

type MoveFile2Responce struct {
	dao.File
}

func MoveFile2(req *MoveFile2Request, res *MoveFile2Responce) error {

	file, err := dao.QueryFileById(req.FileId)
	if err != nil {
		return err
	}
	index := strings.LastIndex(req.Path, "/")
	if index == -1 {
		return errors.New("path not invalid")
	}
	dir, err := dao.QueryDirByPath(req.StoreId, req.Path[:index-1])

	if err == gorm.ErrRecordNotFound && req.Mode == "force" {
		f := dao.MustCreateDir(req.StoreId, req.Path[:index-1])
		_, err = dao.MoveFile(f.ID, file)

	} else if err == nil {
		_, err = dao.MoveFile(dir.ID, file)
	}

	return err
}

type GetFileRequest struct {
	FileId int
}

type GetFileResponce struct {
	dao.File
}

func GetFile(req *GetFileRequest, res *GetFileResponce) error {

	file, err := dao.QueryFileById(req.FileId)

	if err != nil {
		return err
	}

	res.File = file

	return nil
}

type GetFile2Request struct {
	StoreId int
	Path    string
}

func GetFile2(req *GetFile2Request, res *GetFileResponce) error {

	file, err := dao.QueryFileByPath(req.StoreId, req.Path)
	if err != nil {
		return err
	}
	res.File = file
	return nil
}
