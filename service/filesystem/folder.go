package service

import (
	"orm/common/page"
	"orm/dao"
	"time"
)

type CreateaFolderRequest struct {
	FolderName string `json:"foldName"`
	ParentId   int    `json:"parentId"`
	StoreId    int    `json:"storeId"`
}
type CreatFolderResponce struct {
	dao.Folder
}

func CreateFolder(req *CreateaFolderRequest, res *CreatFolderResponce) error {
	res.Folder.Name = req.FolderName
	res.Folder.CreateTime = time.Now()
	res.Folder.ParentId = req.ParentId
	res.Folder.StoreId = req.StoreId
	//如果storeid不存在回报外键错误
	id, err := dao.CreateDir(res.Folder)
	if err != nil {
		return err
	}

	res.Folder.ID = id
	return nil
}

type RemoveFolderRequest struct {
	StoreId  int
	FolderId int
}

func RemoveFolder(req *RemoveFolderRequest) error {
	f := dao.Folder{
		ID:      req.FolderId,
		StoreId: req.StoreId,
	}
	return dao.RemoveDir(f)
}

type MoveFolderRequest struct {
	SrcFolderId int
	DstFolderId int
}

func MoveFolder(req *MoveFolderRequest) error {

	dir, err := dao.QueryDirById(req.SrcFolderId)
	if err != nil {
		return err
	}

	return dao.MoveDir(dir, req.DstFolderId)
}

type MustMoveFolderRequest struct {
	SrcFolderId int
	Path        string
}

type MustMoveFolderResponce struct {
	dao.Folder
}

func MustMoveFolder(req *MustMoveFolderRequest, res *MustMoveFolderResponce) {
	dir, err := dao.QueryDirById(req.SrcFolderId)
	if err != nil {
		panic(err)
	}
	res.Folder = dao.MustMoveDir(dir, req.Path)
}

type CopyFolderRequest struct {
	SrcFolderId int
	DstFolderId int
}

func CopyFolder(req *CopyFolderRequest) error {

	dir, err := dao.QueryDirById(req.SrcFolderId)
	if err != nil {
		return err
	}

	return dao.CopyDir(dir, req.DstFolderId)
}

type MustCopyFolderRequest struct {
	SrcFolderId int
	StoreId     int
	Path        string
}

type MustCopyFolderResponce struct {
	dao.Folder
}

func MustCopyFolder(req *MustCopyFolderRequest, res *MustCopyFolderResponce) {

	dir, err := dao.QueryDirById(req.SrcFolderId)
	if err != nil {
		panic(err)
	}

	res.Folder = dao.MustCopyDir(dir, req.StoreId, req.Path)
}

type GetFolderRequest struct {
	FolderId int
}
type GetFolderResponce struct {
	dao.Folder
}

func GetFolder(req *GetFolderRequest, res *GetFolderResponce) (err error) {
	res.Folder, err = dao.QueryDirById(req.FolderId)
	return
}

type GetFolderListRequest struct {
	FolderId int
}
type GetFolderListResponce struct {
	StoreId int
	Data    []interface{}
}

func GetFolderList(req *GetFolderListRequest, res *GetFolderListResponce) error {
	dir, err := dao.QueryDirById(req.FolderId)
	if err != nil {
		return err
	}
	res.StoreId = dir.StoreId
	ff, errx := dao.QueryFolderList(req.FolderId)
	if errx != nil {
		return errx
	}

	for _, v := range ff {
		res.Data = append(res.Data, v)
	}

	fs, errx := dao.QueryFileList(req.FolderId)

	for _, v := range fs {
		res.Data = append(res.Data, v)
	}
	return errx
}

type GetChildPageRequest struct {
	FolderId int
	Page     int //页号
	Size     int //页的容量
	Order    int
}

type GetChildPageResponce struct {
	Size  int
	Total int
	Data  []interface{}
}

func GetChildPage(req *GetChildPageRequest, res *GetChildPageResponce) error {

	pager := page.Pagger{
		PageOrder: page.FolderFirst & page.TimeDesc,
		Page:      req.Page,
		PageSize:  req.Size,
		OffSet:    0,
	}
	folders, files, err := dao.QueryRows(req.FolderId)

	if err != nil {
		return err
	}
	res.Total = int(folders + files)

	if pager.PageOrder&page.FolderFirst == page.FolderFirst {
		if page.GetOffset(&pager) < int(folders) {
			ff, err := dao.QueryFolderPage(req.FolderId, pager)
			if err != nil {
				return err
			}
			for _, v := range ff {
				res.Data = append(res.Data, v)
			}
			if len(ff) < pager.PageSize {
				pager.Page = 1
				pager.PageSize -= len(ff)
			} else {
				return nil
			}
		} else {
			// folders = 13 files = 14 pageSize 5 page 3 => folders 2 files 3
			// pagesize 5 page 4 => offset 2 pagesize 5
			pager.OffSet = pager.PageSize - int(folders)%pager.PageSize
			pager.Page -= int(folders)/pager.PageSize + 1
		}
		fs, err := dao.QueryFilePage(req.FolderId, pager)
		if err != nil {
			return err
		}

		for _, v := range fs {
			res.Data = append(res.Data, v)
		}
	} else {

		if page.GetOffset(&pager) < int(files) {
			fs, err := dao.QueryFilePage(req.FolderId, pager)
			if err != nil {
				return err
			}
			for _, v := range fs {
				res.Data = append(res.Data, v)
			}
			// 不够
			if len(fs) < pager.PageSize {
				pager.Page = 1
				pager.PageSize -= len(fs)
			} else {
				return nil
			}
		} else {
			// folders = 13 files = 14 pageSize 5 page 3 => folders 2 files 3
			// pagesize 5 page 4 => offset 2 pagesize 5
			pager.OffSet = pager.PageSize - int(folders)%pager.PageSize
			pager.Page -= int(folders)/pager.PageSize + 1
		}
		ff, err := dao.QueryFolderPage(req.FolderId, pager)
		if err != nil {
			return err
		}

		for _, v := range ff {
			res.Data = append(res.Data, v)
		}
	}

	return nil
}

type GetChildTreeRequest struct {
	FolderId int
}

type GetChildTreeResponce struct {
	Data []interface{}
}

func GetChildTree(req *GetChildTreeRequest, res *GetChildTreeResponce) error {

	dir, err := dao.QueryDirById(req.FolderId)
	if err != nil {
		return err
	}

	var errx error

	res.Data, errx = dao.DumpTree(dir)

	return errx
}
