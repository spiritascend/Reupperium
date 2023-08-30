package ddl

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"reupperium/utils"
	"strings"
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

func GetFileInfo(httpclient *http.Client, token string, filecode []string) (FileInfo_Resp, error) {
	Ret := FileInfo_Resp{}

	request, err := http.NewRequest("GET", fmt.Sprintf("https://api-v2.ddownload.com/api/file/info?key=%s&file_code=%s", token, strings.Join(filecode, ",")), nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return Ret, err
	}

	response, err := httpclient.Do(request)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return Ret, err
	}
	defer response.Body.Close()

	if json.NewDecoder(response.Body).Decode(&Ret) != nil {
		return Ret, nil
	}

	if Ret.Status != 200 {
		return Ret, fmt.Errorf("failed to get ddl file info because response status was %d", Ret.Status)
	}

	return Ret, nil
}

func FilesDeleted_Safe(httpclient *http.Client, token string, fileids []string) (bool, error) {
	if len(fileids) == 0 {
		return false, nil
	}

	if len(fileids) >= 50 {
		finfo, err := GetFileInfo(httpclient, token, fileids[:50])

		if err != nil {
			return false, err
		}

		for infoidx := range finfo.Result {
			if finfo.Result[infoidx].Status != 200 || len(finfo.Result) != len(fileids) {
				return true, nil
			}
		}
		return FilesDeleted_Safe(httpclient, token, fileids[50:])

	} else {
		finfo, err := GetFileInfo(httpclient, token, fileids)

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

func FilesDeleted(httpclient *http.Client, config *utils.Config, fileids []string) (bool, error) {
	for tkn := range config.Ddltokens {
		isdeleted, err := FilesDeleted_Safe(httpclient, config.Ddltokens[tkn], fileids)

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
