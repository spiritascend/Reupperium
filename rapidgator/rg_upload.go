package rapidgator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"reupperium/utils"
	"time"

	"github.com/google/uuid"
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

func GetEndpoint(httpclient *http.Client, config *utils.Config, hash string, size string, name string) (GetEndpointResp, error) {
	GEResp := GetEndpointResp{}

	formData := map[string]string{
		"hash":      hash,
		"size":      size,
		"name":      name,
		"folder_id": "0",
		"id":        "0",
		"uuid":      uuid.New().String(),
		"multipart": "false",
		"__token":   config.RapidGator.Cookie.Token,
	}

	req, err := http.NewRequest("POST", "https://rapidgator.net/storage/GetEndpoint2", nil)
	if err != nil {
		return GEResp, err
	}

	q := req.URL.Query()
	for key, value := range formData {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	req.Header.Set("Cookie", fmt.Sprintf("lang=%s;user__=%s;PHPSESSID=%s;__token=%s", config.RapidGator.Cookie.Lang, config.RapidGator.Cookie.UserInfo, config.RapidGator.Cookie.Session, config.RapidGator.Cookie.Token))
	req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"116\", \"Not)A;Brand\";v=\"24\", \"Brave\";v=\"116\"")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")

	resp, err := httpclient.Do(req)
	if err != nil {
		return GEResp, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&GEResp)
	if err != nil {
		return GEResp, err
	}

	if resp.StatusCode != http.StatusOK {
		return GEResp, errors.New("rapidgator_upload_getendpoint error: request didn't respond with status code 200")
	}

	if len(GEResp.Error) > 0 {
		return GEResp, fmt.Errorf("rapidgator_upload_getendpoint error: %s", GEResp.Error)
	}

	return GEResp, nil
}

func GetFileUploadInfo(httpclient *http.Client, config *utils.Config, uuid, filename string) (string, string, error) {
	var usresp map[string]GetFileInfoData

	formData := url.Values{}
	formData.Add("uuid[0][uuid]", uuid)
	formData.Add("__token", config.RapidGator.Cookie.Token)

	reqBody := []byte(formData.Encode())

	req, err := http.NewRequest("POST", "https://rapidgator.net/storage/UploadState", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Cookie", fmt.Sprintf("lang=%s;user__=%s;PHPSESSID=%s;__token=%s", config.RapidGator.Cookie.Lang, config.RapidGator.Cookie.UserInfo, config.RapidGator.Cookie.Session, config.RapidGator.Cookie.Token))
	req.Header.Set("sec-ch-ua", "\"Chromium\";v=\"116\", \"Not)A;Brand\";v=\"24\", \"Brave\";v=\"116\"")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/116.0.0.0 Safari/537.36")

	resp, err := httpclient.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&usresp)
	if err != nil {
		return "", "", err
	}

	return fmt.Sprintf("https://rapidgator.net/file/%s/%s.html", usresp[uuid].Id32, filename), usresp[uuid].Id32, nil
}

func UploadFile(httpclient *http.Client, config *utils.Config, filepath string) (string, error) {
	filename, filehash, filesize, err := utils.GetFileInfo(filepath)

	if err != nil {
		return "", err
	}

	urlinfo, err := GetEndpoint(httpclient, config, filehash, fmt.Sprint(filesize), filename)

	if err != nil {
		return "", err
	}

	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	req, err := http.NewRequest("POST", fmt.Sprintf("%s?ajax=true&qquuid=%s&qqfilename=%s&file=%s", urlinfo.Endpoint, urlinfo.UUID, filename, filename), file)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/octet-stream")

	resp, err := httpclient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var respbody map[string]bool
	err = json.NewDecoder(resp.Body).Decode(&respbody)
	if err != nil {
		return "", err
	}

	if !respbody["success"] {
		return "", errors.New("rapidgator_upload error: upload unsuccessful")
	}

	for {
		url, id, err := GetFileUploadInfo(httpclient, config, urlinfo.UUID, filename)
		if err != nil {
			return "", err
		}

		if len(id) > 0 {
			fileinfo, err := GetFileInfo(httpclient, config, id)
			if err != nil {
				return "", err
			}
			if fileinfo.Response.File.Size != int(filesize) {
				return "", fmt.Errorf("rapidgator_upload_error LocalSize %d, Uploaded Size %d", int(filesize), fileinfo.Response.File.Size)
			}
			Log("Uploaded File: " + path.Base(filepath))
			return url, nil
		}
		time.Sleep(1000)
		continue
	}
}
