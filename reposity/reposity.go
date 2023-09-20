package reposity

import (
	"bytes"
	"errors"
	"io"
	"orm/dao"
	"orm/dao/cache"
	"path"
)

type PartUploadResult struct {
	UploadId     string
	ChunckId     int
	ContentLengh int64
	Md5          string
}

const FilePartSlice = 256 * 1024 * 1024

// 所有仓库层的文件命名都是 idHash+nameHash.{扩展名}
// 不存储目录

type IReposityer interface {
	UploadObject(objectname string, object io.Reader) error
	DownloadObject(objectname string, object io.Writer) (int64, error)
	AppendObject(objectname string, pos int64, object io.Reader) (int64, error)
	DeleteObject(objectname string) error
	CopyObject(string, string) error
	PartsUploadObject(objectname string, uploadId string, chunkId int, r io.Reader) (*PartUploadResult, error)
	ListObjectParts(objectname string, uploadId string) ([]*PartUploadResult, error)
	DownloadParts(objectname string, chunkId int, totalSize int64, w io.Writer) (int64, error)
}

func NewReposityer(types string) IReposityer {

	switch types {
	case "cos":
		return &Cos{}
	case "oss":
		return &Oss{}
	case "hdfs":
		return &Hdfs{}
	default:
		return nil
	}
}

func ValidatePersiter(perist string) bool {
	switch perist {
	case "cos":
		return true
	case "oss":
		return true
	case "hdfs":
		return true
	default:
		return false
	}
}

func getMetaName(f dao.File) string {
	return f.FileHash + path.Ext(f.FileName)
}

// 将一个文件移植到另外一个store中的文件夹下
// 从一个介质转移到另一个介质
func migrateFile(f *dao.File, st1 dao.Store, ff *dao.File, st2 dao.Store) error {
	irepo1 := NewReposityer(st1.Persiter)
	irepo2 := NewReposityer(st2.Persiter)
	buf := bytes.Buffer{}
	_, err := irepo1.DownloadObject(getMetaName(*f), io.ReadWriter(&buf))
	if err != nil {
		return err
	}
	err = irepo2.UploadObject(getMetaName(*f), io.ReadWriter(&buf))
	if err != nil {
		return err
	}

	return nil
}

func GetFile(f dao.File, w io.Writer) error {
	st, err := dao.QueryStoreInfo(f.StoreId)
	if err != nil {
		return err
	}

	irepo := NewReposityer(st.Persiter)
	_, err = irepo.DownloadObject(getMetaName(f), w)
	return err
}

func PutFile(f dao.File, r io.Reader) error {
	st, err := dao.QueryStoreInfo(f.StoreId)
	if err != nil {
		return err
	}
	irepo := NewReposityer(st.Persiter)
	return irepo.UploadObject(getMetaName(f), r)
}

func DelFile(f dao.File) error {
	st, err := dao.QueryStoreInfo(f.StoreId)
	if err != nil {
		return err
	}

	irepo := NewReposityer(st.Persiter)
	return irepo.DeleteObject(f.FileHash + path.Ext(f.FileName))
}

func CreateFile(f *dao.File, r io.Reader) error {

	if f.FileType == "appendable" {
		_, err := AppendFile(f, 0, r)
		return err
	} else {
		return PutFile(*f, r)
	}
}

func AppendFile(f *dao.File, pos int64, r io.Reader) (int64, error) {

	st, err := dao.QueryStoreInfo(f.StoreId)

	if err != nil {
		return -1, err
	}

	irepo := NewReposityer(st.Persiter)

	offset, err := irepo.AppendObject(getMetaName(*f), pos, r)
	//call u
	return offset, err
}

func CopyFile(fSrc *dao.File, fDst *dao.File) error {

	st1, err := dao.QueryStoreInfo(fSrc.StoreId)
	if err != nil {
		return nil
	}
	st2, err := dao.QueryStoreInfo(fDst.StoreId)

	if err != nil {
		return err
	}

	if ValidatePersiter(st1.Persiter) && ValidatePersiter(st2.Persiter) {
		if st1.Persiter == st2.Persiter {
			irepo := NewReposityer(st1.Persiter)
			go irepo.CopyObject(getMetaName(*fSrc), getMetaName(*fDst))
		} else {
			go migrateFile(fSrc, st1, fDst, st2)
			return nil
		}
	} else {
		return errors.New("invalid persiter")
	}

	return nil
}

func CreateShardingFile(f *dao.File, uploadId string, chunkId int, r io.Reader) (*PartUploadResult, error) {

	st, err := dao.QueryStoreInfo(f.StoreId)

	if err != nil {
		return &PartUploadResult{}, err
	}
	irepo := NewReposityer(st.Persiter)

	results, err := irepo.PartsUploadObject(getMetaName(*f), uploadId, chunkId, r)
	if err != nil {
		return results, err
	}
	cache.SetKey("event:upload:"+uploadId, 1, 0)
	return results, nil
}

func FinishShardingFile(f *dao.File, uploadId string) (*PartUploadResult, error) {
	cache.DelKey("event:upload:" + uploadId)
	return CreateShardingFile(f, uploadId, -1, nil)
}

func GetShardingFileBlock(f *dao.File, chunkId int, w io.Writer) (int64, error) {

	st, err := dao.QueryStoreInfo(f.StoreId)

	if err != nil {
		return -1, err
	}
	irepo := NewReposityer(st.Persiter)

	return irepo.DownloadParts(getMetaName(*f), chunkId, f.FileSize, w)
}

func ListShardingBlock(f *dao.File, uploadId string) ([]*PartUploadResult, error) {
	st, err := dao.QueryStoreInfo(f.StoreId)
	results := []*PartUploadResult{}
	if err != nil {
		return results, err
	}
	irepo := NewReposityer(st.Persiter)

	return irepo.ListObjectParts(getMetaName(*f), uploadId)
}
