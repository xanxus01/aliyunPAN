package aliyunDriver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"
)

type FileItem struct {
	CreatedAt time.Time `json:"created_at"`
	DriveId string `json:"drive_id"`
	FileId string	`json:"file_id"`
	ParentFileId string `json:"parent_file_id"`
	Name string `json:"name"`
	Type string `json:"type"`
	Starred bool `json:"starred"`
	UpdatedAt time.Time `json:"updated_at"`
	UserMeta string `json:"user_meta"`
}

type FileDownload struct {
	DomainId          string `json:"domain_id"`
	DriveId         string `json:"drive_id"`
	FileId          string `json:"file_id"`
	RevisionId      string `json:"revision_id"`
	Method          string `json:"method"`
	Url             string `json:"url"`
	InternalUrl     string `json:"internal_url"`
	Expiration      string `json:"expiration"`
	Size            int `json:"size"`
	Crc64Hash       string `json:"crc_64_hash"`
	ContentHash     string `json:"content_hash"`
	ContentHashName string `json:"content_hash_name"`
}

func (d Driver)List(parentFileId ...string) ([]FileItem, error) {
	//param := `{"drive_id":"53986141","parent_file_id":"root","limit":20,"all":false,"url_expire_sec":14400,"image_thumbnail_process":"image/resize,w_256/format,jpeg","image_url_process":"image/resize,w_1920/format,jpeg/interlace,1","video_thumbnail_process":"video/snapshot,t_1000,f_jpg,ar_auto,w_256","fields":"*","order_by":"updated_at","order_direction":"DESC"}`
	if len(parentFileId) == 0 {
		parentFileId = append(parentFileId, "root")
	}
	param := map[string]interface{}{
		"drive_id":                "53986141",
		"parent_file_id":          parentFileId[0],
		"limit":                   20,
		"all":                     false,
		"url_expire_sec":          14400,
		"image_thumbnail_process": "image/resize,w_256/format,jpeg",
		"image_url_process":       "image/resize,w_1920/format,jpeg/interlace,1",
		"video_thumbnail_process": "video/snapshot,t_1000,f_jpg,ar_auto,w_256",
		"fields":                  "*",
		"order_by":                "updated_at",
		"order_direction":         "DESC",
	}
	list := struct {
		Items []FileItem
	}{}
	paramData, err := json.Marshal(param)
	if err != nil {
		return nil,err
	}
	r, err := d.post(makeUrl("/adrive/v3/file/list"), bytes.NewReader(paramData), &list)
	log(string(r.Data))
	if err != nil {
		return nil,err
	}
	return list.Items,err
}

func (d Driver)GetDownloadUrl(item FileItem) (string, error) {
	if item.Type != "file" {
		return "",newError(item.Name + " type is not file")
	}
	m := map[string]string{
		"drive_id":item.DriveId,
		"file_id":item.FileId,
	}
	downRes := FileDownload{}
	r,err := d.post(makeUrl("/v2/file/get_download_url"), paramBody(m), &downRes)
	fmt.Printf("%+v\n",item)
	fmt.Printf("%+v\n",downRes)
	fmt.Println(string(r.Data))
	if err != nil {
		return "", err
	}

	return downRes.Url, nil
}

func Download(url string, file string) (bool,error) {
	f,err := os.OpenFile(file, os.O_CREATE, 0666)
	if err != nil {
		return false, err
	}
	req,err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false,err
	}
	req.Header.Set("Referer","https://www.aliyundrive.com/")
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return false,err
	}
	data,_ := ioutil.ReadAll(response.Body)
	_, err = f.Write(data)
	if err != nil {
		return false,err
	}
	return true,nil
}