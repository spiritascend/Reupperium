package ddl

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

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

func UploadFileSafe(rc *resty.Client, token string, fp string) (string, error) {

	file, err := os.OpenFile(fp, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)

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

	file, err = os.Open(fp)
	if err != nil {
		return "", err
	}
	defer file.Close()

	filestat, _ := file.Stat()

	serverurl, sid, err := GetServer(rc, token)

	if err != nil {
		return "", err
	}

	var UFResp UploadFile_Resp
	resp, err := rc.R().
		SetFileReader("file", UploadFile_SanitizeFileName(filepath.Base(fp)), file).
		SetFormData(map[string]string{"sess_id": sid}).
		Post(serverurl)

	if err != nil {
		return "", err
	}

	if err = json.Unmarshal(resp.Body(), &UFResp); err != nil {
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

	if fileinfo.Result[0].Size != fmt.Sprint(filestat.Size()) {
		return "", errors.New("sizemismatch_ddl")
	}

	return fmt.Sprintf("https://ddownload.com/%s?%s", UFResp[0].File_code, filepath.Base(fp)), nil
}

func UploadFile(rc *resty.Client, tokens []string, fp string) (string, error) {
	for tkn := range tokens {
		url, err := UploadFileSafe(rc, tokens[tkn], fp)

		if err != nil {
			if err.Error() == "uploadfailed_ddl_diskspacequotamax" {
				continue
			}
		}
		return url, err
	}

	return "", errors.New("uploadfailed_ddl_diskspacequotamax")
}
