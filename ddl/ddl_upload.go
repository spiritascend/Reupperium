package ddl

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reupperium/utils"
	"sync"
)

type DDL_UploadError []struct {
	FileCode   string `json:"file code"`
	FileStatus string `json:"file status"`
}

type GetServer_Resp struct {
	Msg        string `json:"msg"`
	ServerTime string `json:"server_time"`
	Status     int    `json:"status"`
	SessID     string `json:"sess_id"`
	Result     string `json:"result"`
}

type UploadFile_Resp []struct {
	File_size   int    `json:"file_size,omitempty"`
	File_code   string `json:"file_code"`
	File_status string `json:"file_status"`
}

func GetServer(httpclient *http.Client, token string) (string, string, error) {
	GSResp := GetServer_Resp{}

	request, err := http.NewRequest("GET", fmt.Sprintf("https://api-v2.ddownload.com/api/upload/server?key=%s", token), nil)
	if err != nil {
		return "", "", err
	}

	response, err := httpclient.Do(request)
	if err != nil {
		return "", "", err
	}
	defer response.Body.Close()

	if err = json.NewDecoder(response.Body).Decode(&GSResp); err != nil {
		return "", "", err
	}

	if GSResp.Status != 200 {
		return "", "", fmt.Errorf("failed to get ddl upload server: result was %d", GSResp.Status)
	}

	return GSResp.Result, GSResp.SessID, nil
}

func UploadFileSafe(httpclient *http.Client, token string, fp string) (string, error) {
	file, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filestat, err := file.Stat()
	if err != nil {
		return "", err
	}

	serverurl, sid, err := GetServer(httpclient, token)
	if err != nil {
		return "", err
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		defer pw.Close()
		fieldWriter, err := writer.CreateFormField("sess_id")
		if err != nil {
			return
		}
		io.WriteString(fieldWriter, sid)
		partWriter, err := writer.CreateFormFile("file", filepath.Base(UploadFile_SanitizeFileName(fp)))
		if err != nil {
			return
		}
		_, err = io.Copy(partWriter, file)
		if err != nil {
			return
		}

		writer.Close()
	}()

	request, err := http.NewRequest("POST", serverurl, pr)
	if err != nil {
		return "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := httpclient.Do(request)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	UFResp := UploadFile_Resp{}
	err = json.NewDecoder(response.Body).Decode(&UFResp)

	if err != nil {
		return "", err
	}

	if len(UFResp) == 0 {
		return "", errors.New("uploadfailed_ddl")
	}

	if UFResp[0].File_code == "undef" && UFResp[0].File_status == "not enough disk space on your account" {
		return "", errors.New("uploadfailed_ddl_diskspacequotamax")
	}

	fileinfo, err := GetFileInfo(httpclient, token, []string{UFResp[0].File_code})

	if err != nil {
		return "", err
	}

	if len(fileinfo.Result) == 0 {
		return "", errors.New("failed to get file_info_result")
	}

	if fileinfo.Result[0].Size != fmt.Sprint(filestat.Size()) {
		return "", errors.New("sizemismatch_ddl")
	}

	stat, err := file.Stat()

	if err != nil {
		return "", err
	}

	if fileinfo.Result[0].Size != fmt.Sprint(stat.Size()) {
		return "", fmt.Errorf("uploadfailed_ddl_filesizemismatch LocalSize %d, Uploaded Size %s", stat.Size(), fileinfo.Result[0].Size)
	}
	Log("Uploaded File: " + path.Base(fp))

	wg.Wait()

	return fmt.Sprintf("https://ddownload.com/%s?%s", UFResp[0].File_code, filepath.Base(UploadFile_SanitizeFileName(fp))), nil
}

func UploadFile(httpclient *http.Client, fp string) (string, error) {
	config, err := utils.GetConfig()

	if err != nil {
		return "", err
	}

	for tkn := range config.Ddltokens {
		url, err := UploadFileSafe(httpclient, config.Ddltokens[tkn], fp)

		if err != nil {
			if err.Error() == "uploadfailed_ddl_diskspacequotamax" {
				continue
			}
		}
		return url, err
	}

	return "", errors.New("uploadfailed_ddl_diskspacequotamax")
}
