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

	"gopkg.in/resty.v1"
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

type Auth_error struct {
	Msg        string `json:"msg"`
	ServerTime string `json:"server_time"`
	Status     int    `json:"status"`
}

type UploadFile_Resp []struct {
	File_size   int    `json:"file_size,omitempty"`
	File_code   string `json:"file_code"`
	File_status string `json:"file_status"`
}

func GetServer(rc *resty.Client, token string) (string, string, error) {
	var GSResp GetServer_Resp
	resp, err := rc.R().Get(fmt.Sprintf("https://api-v2.ddownload.com/api/upload/server?key=%s", token))

	if err != nil {
		return "", "", err
	}

	if err = json.Unmarshal(resp.Body(), &GSResp); err != nil {
		return "", "", err
	}

	var GSAE Auth_error

	if err = json.Unmarshal(resp.Body(), &GSAE); err != nil {
		return "", "", err
	}

	if GSAE.Status == 400 {
		return "", "", errors.New(GSAE.Msg)
	}

	return GSResp.Result, GSResp.SessID, nil
}

func UploadFileSafe(rc *resty.Client, httpclient *http.Client, token string, fp string) (string, error) {
	file, err := os.Open(fp)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filestat, _ := file.Stat()

	serverurl, sid, err := GetServer(rc, token)

	if err != nil {
		return "", err
	}

	pr, pw := io.Pipe()
	writer := multipart.NewWriter(pw)

	go func() {
		defer pw.Close()
		fieldWriter, err := writer.CreateFormField("sess_id")
		if err != nil {
			fmt.Println("Error creating sess_id field:", err)
			return
		}
		io.WriteString(fieldWriter, sid)
		partWriter, err := writer.CreateFormFile("file", fp)
		if err != nil {
			fmt.Println("Error creating form file:", err)
			return
		}
		_, err = io.Copy(partWriter, file)
		if err != nil {
			fmt.Println("Error copying file data:", err)
			return
		}

		writer.Close()
	}()

	request, err := http.NewRequest("POST", serverurl, pr)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return "", err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	// Using the http library because restyv2 doesn't support this type of request with a file streaming

	response, err := httpclient.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return "", err
	}
	defer response.Body.Close()

	respbody, err := io.ReadAll(response.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return "", err
	}

	var UFResp UploadFile_Resp
	if err = json.Unmarshal(respbody, &UFResp); err != nil {
		return "", err
	}

	if len(UFResp) == 0 {
		return "", errors.New("uploadfailed_ddl")
	}

	if UFResp[0].File_code == "undef" && UFResp[0].File_status == "not enough disk space on your account" {
		return "", errors.New("uploadfailed_ddl_diskspacequotamax")
	}

	fileinfo, err := GetFileInfo(rc, token, []string{UFResp[0].File_code})

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
	return fmt.Sprintf("https://ddownload.com/%s?%s", UFResp[0].File_code, filepath.Base(fp)), nil
}

func UploadFile(rc *resty.Client, httpclient *http.Client, fp string) (string, error) {
	config, err := utils.GetConfig()

	if err != nil {
		return "", err
	}

	for tkn := range config.Ddltokens {
		url, err := UploadFileSafe(rc, httpclient, config.Ddltokens[tkn], fp)

		if err != nil {
			if err.Error() == "uploadfailed_ddl_diskspacequotamax" {
				continue
			}
		}
		return url, err
	}

	return "", errors.New("uploadfailed_ddl_diskspacequotamax")
}
