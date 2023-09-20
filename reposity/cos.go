package reposity

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"orm/config"
	"path"

	"github.com/tencentyun/cos-go-sdk-v5"
)

type Cos struct {
	bassUrl string
}

func (c *Cos) MustNewClient() *cos.ObjectService {
	conf := config.GetConfig()
	u, _ := url.Parse(fmt.Sprintf("https://%v.cos.%v.myqcloud.com", conf.CosConfig.BucketName, conf.CosConfig.BucketRegiom))
	c.bassUrl = u.String()
	baseUrl := &cos.BaseURL{
		BucketURL: u,
	}
	cc := cos.NewClient(baseUrl, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  conf.CosConfig.SecretId,
			SecretKey: conf.CosConfig.SecretKey,
		},
	})
	return cc.Object
}

func (c *Cos) UploadObject(filename string, reader io.Reader) error {
	fileSuffix := path.Ext(filename)
	obj := c.MustNewClient()

	_, err := obj.Put(context.TODO(), "files/"+filename+fileSuffix, reader, nil)
	if err != nil {
		return err
	}

	return nil
}
func (c *Cos) DownloadObject(filename string, writer io.Writer) (int64, error) {
	cc := c.MustNewClient()

	resp, err := cc.Get(context.Background(), "files/"+filename, nil)
	if err != nil {
		return -1, nil
	}
	defer resp.Body.Close()

	return io.Copy(writer, resp.Body)
}

func (c *Cos) DeleteObject(filename string) error {

	cc := c.MustNewClient()

	_, err := cc.Delete(context.Background(), "files/"+filename, nil)

	return err
}

func (c *Cos) AppendObject(objectname string, index int64, object io.Reader) (int64, error) {
	cc := c.MustNewClient()

	pos, _, err := cc.Append(context.Background(), "files/"+objectname, int(index), object, nil)
	return int64(pos), err
}

func (c *Cos) CopyObject(src string, dst string) error {
	cc := c.MustNewClient()

	_, _, err := cc.Copy(context.Background(), "file/"+dst, c.bassUrl+"/files/"+src, nil)
	return err
}

func (c *Cos) PartsUploadObject(objectname string, uploadId string, chunkId int, r io.Reader) (*PartUploadResult, error) {

	cc := c.MustNewClient()
	// 第一次调用，为objectname对象初始化分片操作
	if uploadId == "" {
		v, _, err := cc.InitiateMultipartUpload(context.Background(), "files/"+objectname, nil)
		if err != nil {
			return &PartUploadResult{}, err
		}
		uploadId = v.UploadID
	}
	// 结束事件
	if chunkId == -1 {
		_, _, err := cc.CompleteMultipartUpload(context.Background(), "files/"+objectname, uploadId, nil)
		return &PartUploadResult{}, err
	}

	resp, err := cc.UploadPart(context.Background(), "files/"+objectname, uploadId, chunkId, r, nil)

	if err != nil {
		return &PartUploadResult{}, err
	}

	return &PartUploadResult{
		UploadId:     uploadId,
		ChunckId:     chunkId,
		ContentLengh: resp.ContentLength,
		Md5:          resp.Header.Get("ETag"),
	}, nil
}
func (c *Cos) ListObjectParts(objectname string, uploadId string) ([]*PartUploadResult, error) {
	cc := c.MustNewClient()

	resp, _, err := cc.ListParts(context.Background(), "files/"+objectname, uploadId, nil)

	if err != nil {
		return []*PartUploadResult{}, err
	}

	results := []*PartUploadResult{}

	for _, v := range resp.Parts {
		results = append(results, &PartUploadResult{
			UploadId:     uploadId,
			ChunckId:     v.PartNumber,
			ContentLengh: v.Size,
			Md5:          v.ETag,
		})
	}

	return results, nil
}
func (c *Cos) DownloadParts(objectname string, chunkId int, totalSize int64, w io.Writer) (int64, error) {

	begin := chunkId * FilePartSlice
	ranges := fmt.Sprintf("bytes=%d-", begin)
	if totalSize > (int64(chunkId)+1)*FilePartSlice {
		ranges += fmt.Sprintf("%d", (chunkId+1)*FilePartSlice)
	}
	cc := c.MustNewClient()
	options := &cos.ObjectGetOptions{
		Range: ranges,
	}
	v, err := cc.Get(context.Background(), "files/"+objectname, options)
	if err != nil {
		return -1, err
	}
	defer v.Body.Close()
	return io.Copy(w, v.Body)
}
