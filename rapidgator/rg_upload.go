package rapidgator

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reupperium/utils"
	"time"

	"github.com/google/uuid"
	"gopkg.in/resty.v1"
)

type GetEndpointResp struct {
	Endpoint string `json:"endpoint"`
	Type     int    `json:"type"`
	UUID     string `json:"uuid"`
	Error    string `json:"error,omitempty"`
}
type GetFileInfoData struct {
	Success bool   `json:"success"`
	Id32    string `json:"id32"`
}

func GetEndpoint(rc *resty.Client, hash string, size string, name string) (GetEndpointResp, error) {
	var Ret GetEndpointResp

	config, err := utils.GetConfig()

	if err != nil {
		return GetEndpointResp{}, err
	}

	resp, err := rc.R().
		SetFormData(map[string]string{
			"hash":      hash,
			"size":      size,
			"name":      name,
			"folder_id": "0",
			"id":        "0",
			"uuid":      uuid.New().String(),
			"multipart": "false",
			"__token":   config.RapidGator.Cookie.Token,
		}).
		SetHeaders(map[string]string{
			"Cookie":           fmt.Sprintf("lang=%s;user__=%s;PHPSESSID=%s;__token=%s", config.RapidGator.Cookie.Lang, config.RapidGator.Cookie.UserInfo, config.RapidGator.Cookie.Session, config.RapidGator.Cookie.Token),
			"sec-ch-ua":        "\"Chromium\";v=\"116\", \"Not)A;Brand\";v=\"24\", \"Brave\";v=\"116\"",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		}).
		Post("https://rapidgator.net/storage/GetEndpoint2")

	if err != nil {
		return GetEndpointResp{}, err
	}

	err = json.Unmarshal(resp.Body(), &Ret)

	if err != nil {
		return GetEndpointResp{}, err
	}

	if resp.StatusCode() != 200 {
		return GetEndpointResp{}, errors.New("rapidgator_upload_getendpoint error: request didn't respond with status code 200")
	}

	if len(Ret.Error) > 0 {
		return GetEndpointResp{}, fmt.Errorf("rapidgator_upload_getendpoint error: %s", Ret.Error)
	}

	return Ret, nil
}

func GetFileUploadInfo(rc *resty.Client, uuid string, filename string) (string, string, error) {
	config, err := utils.GetConfig()

	if err != nil {
		return "", "", err
	}

	resp, err := rc.R().
		SetFormData(map[string]string{
			"uuid[0][uuid]": uuid,
			"__token":       config.RapidGator.Cookie.Token,
		}).
		SetHeaders(map[string]string{
			"Cookie":           fmt.Sprintf("lang=%s;user__=%s;PHPSESSID=%s;__token=%s", config.RapidGator.Cookie.Lang, config.RapidGator.Cookie.UserInfo, config.RapidGator.Cookie.Session, config.RapidGator.Cookie.Token),
			"sec-ch-ua":        "\"Chromium\";v=\"116\", \"Not)A;Brand\";v=\"24\", \"Brave\";v=\"116\"",
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36",
		}).
		Post("https://rapidgator.net/storage/UploadState")

	if err != nil {
		return "", "", err
	}

	var usresp map[string]GetFileInfoData

	err = json.Unmarshal(resp.Body(), &usresp)

	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("https://rapidgator.net/file/%s/%s.html", usresp[uuid].Id32, filename), usresp[uuid].Id32, nil
}

func UploadFile(rc *resty.Client, filepath string) (string, error) {

	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

	if err != nil {
		return "", err
	}

	// Change Hash
	timestamp := time.Now().Unix()
	timestampBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timestampBytes, uint32(timestamp))

	_, err = file.Write(timestampBytes)
	if err != nil {
		return "", err
	}
	file.Close()

	filename, filehash, filesize, err := utils.GetFileInfo(filepath)

	if err != nil {
		return "", err
	}

	urlinfo, err := GetEndpoint(rc, filehash, fmt.Sprint(filesize), filename)

	if err != nil {
		return "", err
	}

	file, err = os.Open(filepath)
	if err != nil {
		return "", err
	}

	resp, err := rc.R().
		SetHeaders(map[string]string{"Content-Type": "application/octet-stream"}).
		SetBody(file).
		Post(fmt.Sprintf("%s?ajax=true&qquuid=%s&qqfilename=%s&file=%s", urlinfo.Endpoint, urlinfo.UUID, filename, filename))

	if err != nil {
		return "", err
	}

	var respbody map[string]bool

	err = json.Unmarshal(resp.Body(), &respbody)
	if err != nil {
		return "", err
	}

	if !respbody["success"] {
		return "", errors.New("rapidgator_upload error: upload unsuccessful")
	}

	for {
		url, id, err := GetFileUploadInfo(rc, urlinfo.UUID, filename)
		if err != nil {
			return "", err
		}

		if len(id) > 0 {
			return url, nil
		}
		time.Sleep(1000)
		continue
	}
}
