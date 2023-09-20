package reposity

import (
	"io"
	"io/fs"
	"log"
	"orm/common/util"
	"orm/config"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/colinmarc/hdfs"
)

type Hdfs struct {
}

func MustNewClient() *hdfs.Client {
	conf := config.GetConfig()

	client, err := hdfs.New(conf.HdfsMaster)
	if err != nil {
		panic(err)
	}
	return client
}
func (h *Hdfs) UploadObject(objectname string, r io.Reader) error {

	client := MustNewClient()

	defer func() {
		err := recover()
		log.Println(err)
	}()

	w, err := client.CreateFile("files/"+objectname, 2, 1024*1024*64, 0777)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, r)

	return err
}
func (h *Hdfs) DownloadObject(objectname string, w io.Writer) (int64, error) {

	client := MustNewClient()

	defer func() {
		err := recover()
		log.Println(err)
	}()

	r, err := client.Open("files/" + objectname)
	if err != nil {
		return -1, err
	}

	return io.Copy(w, r)
}
func (h *Hdfs) DeleteObject(objectname string) error {

	client := MustNewClient()

	defer func() {
		err := recover()
		log.Println(err)
	}()

	return client.Remove("files/" + objectname)
}

func (h *Hdfs) AppendObject(objectname string, index int64, r io.Reader) (int64, error) {

	client := MustNewClient()

	w, err := client.Append("files/" + objectname)
	if err != nil {
		return -1, err
	}
	defer w.Close()
	return io.Copy(w, r)
}

func (h *Hdfs) CopyObject(src string, dst string) error {

	client := MustNewClient()

	r, err := client.Open("files/" + src)

	if err != nil {
		return err
	}

	defer r.Close()
	w, err := client.Create("files/" + dst)

	if err != nil {
		return err
	}

	defer w.Close()
	_, err = io.Copy(w, r)

	return err
}

func (h *Hdfs) PartsUploadObject(objectname string, uploadId string, chunkId int, r io.Reader) (*PartUploadResult, error) {
	if uploadId == "" {
		uploadId = util.Uuid()
	}

	hh := MustNewClient()

	if _, err := hh.Stat("files/part"); err != nil {
		os.Mkdir("files/part", 0777)
	}

	filename := "files/part/" + path.Base(objectname) + "-" + strconv.Itoa(chunkId)

	if chunkId == -1 {
		h.CopyObject("files/part/"+path.Base(objectname)+"-0", "files/"+objectname)
		i := 1
		for {
			if _, err := hh.Stat("files/part" + path.Base(objectname) + strconv.Itoa(i)); err != nil {
				rs, _ := hh.Open("files/part" + path.Base(objectname) + strconv.Itoa(i-1))
				m5, _ := rs.Checksum()
				defer rs.Close()
				return &PartUploadResult{
					UploadId: uploadId,
					ChunckId: i,
					Md5:      string(m5),
				}, nil
			}
			w, _ := hh.Append("files/" + objectname)
			r, _ := hh.Open("files/part" + path.Base(objectname) + strconv.Itoa(i))
			defer w.Close()
			defer r.Close()
			io.Copy(w, r)
		}
	}

	//md5 := md5.New()

	wr, _ := hh.CreateFile(filename, 3, 64*1024*1024, 0777)
	defer wr.Close()

	lens, err := io.Copy(wr, r)
	if err != nil {
		return &PartUploadResult{}, err
	}

	rs, _ := hh.Open(filename)
	defer rs.Close()
	md5, _ := rs.Checksum()
	return &PartUploadResult{
		UploadId:     uploadId,
		ChunckId:     chunkId,
		ContentLengh: lens,
		Md5:          string(md5),
	}, nil
}
func (h *Hdfs) ListObjectParts(objectname string, uploadId string) ([]*PartUploadResult, error) {

	hdfs := MustNewClient()
	results := []*PartUploadResult{}
	hdfs.Walk("files/part", func(path string, info fs.FileInfo, err error) error {
		if err == nil {
			index := strings.Index(info.Name(), "-") + 1
			id, _ := strconv.Atoi(info.Name()[:index+1])
			results = append(results, &PartUploadResult{
				UploadId:     uploadId,
				ContentLengh: info.Size(),
				Md5:          "",
				ChunckId:     id,
			})
		}
		return nil
	})

	return results, nil
}
func (h *Hdfs) DownloadParts(objectname string, chunkId int, totalSize int64, w io.Writer) (int64, error) {

	hh := MustNewClient()

	file, err := hh.Open("files/" + objectname)
	if err != nil {
		return -1, err
	}

	if int64((chunkId+1)*FilePartSlice) > totalSize {
		file.Seek(int64(chunkId)*FilePartSlice-totalSize, 2)
		return io.Copy(w, file)
	}
	file.Seek(int64(chunkId*FilePartSlice), 0)
	return io.Copy(w, file)
}
