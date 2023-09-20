package reposity

import (
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"orm/config"
	"orm/dao/cache"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
)

type Oss struct {
}

func (o *Oss) MustGetBucket() *oss.Bucket {
	conf := config.GetConfig()
	// 创建OSSClient实例。
	client, err := oss.New(conf.OssConfig.EndPoint, conf.OssConfig.AccessId, conf.OssConfig.AccessSecret)
	if err != nil {
		fmt.Println("创建实例Error:", err)
		panic(err)
	}

	// 获取存储空间。
	bucket, err := client.Bucket(conf.OssConfig.BucketName)
	if err != nil {
		fmt.Println("获取存储空间Error:", err)
		panic(err)
	}

	return bucket
}

// 上传文件至阿里云
func (o *Oss) UploadObject(filename string, reader io.Reader) error {
	//获取文件后缀
	bucket := o.MustGetBucket()

	defer func() {
		err := recover()
		log.Println(err)
	}()
	// 上传本地文件。

	return bucket.PutObject("files/"+filename, reader, nil)
}

// 从oss下载文件
func (o *Oss) DownloadObject(filename string, writer io.Writer) (int64, error) {
	bucket := o.MustGetBucket()

	defer func() {
		err := recover()
		log.Println(err)
	}()

	// 下载文件到流。
	body, err := bucket.GetObject("files/" + filename)
	if err != nil {
		fmt.Println("Error:", err)
	}
	// 数据读取完成后，获取的流必须关闭，否则会造成连接泄漏，导致请求无连接可用，程序无法正常工作。
	defer body.Close()

	return io.Copy(writer, body)
}

// 从oss删除文件
func (o *Oss) DeleteObject(filename string) error {

	bucket := o.MustGetBucket()

	defer func() {
		err := recover()
		log.Println(err)
	}()
	// 获取存储空间。

	return bucket.DeleteObject("files/" + filename)

}

func (o *Oss) AppendObject(filename string, index int64, r io.Reader) (int64, error) {
	bucket := o.MustGetBucket()

	return bucket.AppendObject("files/"+filename, r, int64(index))

}

func (o *Oss) CopyObject(src string, dst string) error {

	bucket := o.MustGetBucket()

	_, err := bucket.CopyObject("files/"+src, "files/"+dst)

	return err
}

func (o *Oss) PartsUploadObject(objectname string, uploadId string, chunkId int, r io.Reader) (*PartUploadResult, error) {

	bucket := o.MustGetBucket()

	if uploadId == "" {
		resp, err := bucket.InitiateMultipartUpload("files/" + objectname)
		if err != nil {
			return &PartUploadResult{}, err
		}
		x, _ := xml.Marshal(&resp)
		cache.SetKey("repo:oss:"+resp.UploadID, string(x), 0)
		uploadId = resp.UploadID
	}
	mucr := oss.InitiateMultipartUploadResult{}
	x, err := cache.GetKey("repo:oss:" + uploadId)
	if err != nil {
		return &PartUploadResult{}, err
	}
	xml.Unmarshal([]byte(x), &mucr)
	if chunkId == -1 {
		strs, _ := cache.LRange("repo:oss:list:"+uploadId, 0, -1)
		parts := []oss.UploadPart{}

		for _, v := range strs {
			part := oss.UploadPart{}
			xml.Unmarshal([]byte(v), &part)
			parts = append(parts, part)
		}
		_, err := bucket.CompleteMultipartUpload(mucr, parts)
		if err == nil {
			cache.DelKey("repo:oss:list:" + uploadId)
		}
		return &PartUploadResult{}, err
	}

	part, err := bucket.UploadPart(mucr, r, FilePartSlice, chunkId)

	if err != nil {
		return &PartUploadResult{}, nil
	}
	sx, _ := xml.Marshal(&part)
	cache.LPush("repo:oss:list:"+uploadId, string(sx))
	return &PartUploadResult{
		UploadId: uploadId,
		ChunckId: part.PartNumber,
		Md5:      part.ETag,
	}, nil
}
func (o *Oss) ListObjectParts(objectname string, uploadId string) ([]*PartUploadResult, error) {
	result := []*PartUploadResult{}
	bu := o.MustGetBucket()
	mucr := oss.InitiateMultipartUploadResult{}
	x, err := cache.GetKey("repo:oss:" + uploadId)
	if err != nil {
		return result, err
	}
	xml.Unmarshal([]byte(x), &mucr)
	resp, err := bu.ListUploadedParts(mucr)
	for _, v := range resp.UploadedParts {
		result = append(result, &PartUploadResult{
			UploadId:     uploadId,
			ContentLengh: int64(v.Size),
			Md5:          v.ETag,
			ChunckId:     v.PartNumber,
		})
	}
	return result, err
}
func (o *Oss) DownloadParts(objectname string, chunkId int, totalSize int64, w io.Writer) (int64, error) {

	bu := o.MustGetBucket()

	begin := chunkId * FilePartSlice
	ranges := fmt.Sprintf("bytes=%d-", begin)
	if totalSize > (int64(chunkId)+1)*FilePartSlice {
		ranges += fmt.Sprintf("%d", (chunkId+1)*FilePartSlice)
	}

	resp, err := bu.GetObject("files/"+objectname, oss.NormalizedRange(ranges))

	if err != nil {
		return -1, err
	}
	return io.Copy(w, resp)
}
