package ddl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"reupperium/utils"
	"strings"

	"gopkg.in/resty.v1"
)

type FileInfo_Resp struct {
	Msg        string `json:"msg"`
	ServerTime string `json:"server_time"`
	Status     int    `json:"status"`
	Result     []struct {
		Status   int    `json:"status"`
		Filecode string `json:"filecode"`
		Name     string `json:"name,omitempty"`
		Download string `json:"download,omitempty"`
		Size     string `json:"size,omitempty"`
		Uploaded string `json:"uploaded,omitempty"`
	} `json:"result"`
}

type FilesDeleted_Resp struct {
	Msg        string `json:"msg"`
	ServerTime string `json:"server_time"`
	Status     int    `json:"status"`
	Result     []struct {
		FileCode      string `json:"file_code"`
		Name          string `json:"name"`
		Deleted       string `json:"deleted"`
		DeletedAgoSec string `json:"deleted_ago_sec"`
	} `json:"result"`
}

func UploadFile_SanitizeFileName(filename string) string {
	regex, err := regexp.Compile(`[^\x20-\x7E]|[\!$%&]`)
	if err != nil {
		return filename
	}
	return regex.ReplaceAllString(filename, "x")
}

func GetFileInfo(rc *resty.Client, token string, filecode []string) (FileInfo_Resp, error) {
	var Ret FileInfo_Resp

	url := fmt.Sprintf("https://api-v2.ddownload.com/api/file/info?key=%s&file_code=%s", token, strings.Join(filecode, ","))
	resp, err := rc.R().Get(url)
	if err != nil {
		return Ret, fmt.Errorf("error making GET request: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return Ret, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	if err := json.Unmarshal(resp.Body(), &Ret); err != nil {
		return Ret, fmt.Errorf("error unmarshaling JSON response: %w", err)
	}

	return Ret, nil
}

func FilesDeleted_Safe(rc *resty.Client, token string, fileids []string) (bool, error) {
	var Newfileids []string

	if len(fileids) == 0 {
		return false, nil
	}

	if len(fileids) >= 50 {
		finfo, err := GetFileInfo(rc, token, fileids[:50])
		Newfileids = fileids[50:]

		if err != nil {
			return false, err
		}

		for infoidx := range finfo.Result {
			if finfo.Result[infoidx].Status != 200 || len(finfo.Result) != len(fileids) {
				return true, nil
			}
		}
		return FilesDeleted_Safe(rc, token, Newfileids)

	} else {
		finfo, err := GetFileInfo(rc, token, fileids)

		if err != nil {
			return false, err
		}

		if finfo.Status != 200 {
			return true, nil
		}

		for infoidx := range finfo.Result {
			if finfo.Result[infoidx].Status != 200 {
				return true, nil
			}
		}
		return false, nil
	}
}

func FilesDeleted(rc *resty.Client, config *utils.Config, fileids []string) (bool, error) {
	for tkn := range config.Ddltokens {
		isdeleted, err := FilesDeleted_Safe(rc, config.Ddltokens[tkn], fileids)

		if err != nil {
			return false, err
		}

		if !isdeleted {
			continue
		} else {
			return true, nil
		}
	}

	return false, nil
}
