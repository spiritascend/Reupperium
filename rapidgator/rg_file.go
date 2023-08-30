package rapidgator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reupperium/utils"
)

type FolderFile struct {
	FileID    string `json:"file_id"`
	Mode      int    `json:"mode"`
	ModeLabel string `json:"mode_label"`
	FolderID  string `json:"folder_id"`
	Name      string `json:"name"`
	Hash      string `json:"hash"`
	Size      int    `json:"size"`
	Created   int    `json:"created"`
	URL       string `json:"url"`
}

type FolderResp struct {
	Response struct {
		Folder struct {
			FolderID       string       `json:"folder_id"`
			Mode           int          `json:"mode"`
			ModeLabel      string       `json:"mode_label"`
			ParentFolderID any          `json:"parent_folder_id"`
			Name           string       `json:"name"`
			URL            string       `json:"url"`
			NbFolders      int          `json:"nb_folders"`
			NbFiles        int          `json:"nb_files"`
			SizeFiles      int64        `json:"size_files"`
			Created        int          `json:"created"`
			Folders        []any        `json:"folders"`
			Files          []FolderFile `json:"files"`
		} `json:"folder"`
		Pager struct {
			Current int `json:"current"`
			Total   int `json:"total"`
		} `json:"pager"`
	} `json:"response"`
	Status  int `json:"status"`
	Details any `json:"details"`
}

type FileInfoResp struct {
	Response struct {
		File struct {
			FileID      string `json:"file_id"`
			Mode        int    `json:"mode"`
			ModeLabel   string `json:"mode_label"`
			FolderID    string `json:"folder_id"`
			Name        string `json:"name"`
			Hash        string `json:"hash"`
			Size        int    `json:"size"`
			Created     int    `json:"created"`
			URL         string `json:"url"`
			NbDownloads int    `json:"nb_downloads"`
		} `json:"file"`
	} `json:"response"`
	Status  int `json:"status"`
	Details any `json:"details"`
}

func GetFileInfo(httpclient *http.Client, config *utils.Config, id string) (FileInfoResp, error) {
	fileInfo := FileInfoResp{}

	resp, err := httpclient.Get(fmt.Sprintf("https://rapidgator.net/api/v2/file/info/?file_id=%s&token=%s", id, config.RapidGator.Token))
	if err != nil {
		return fileInfo, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&fileInfo)
	if err != nil {
		return fileInfo, err
	}

	return fileInfo, nil
}

func FolderFilesExist(array []FolderFile, target string) bool {

	for i := 0; i < len(array); i++ {
		if array[i].FileID == target {
			return true
		}
	}
	return false
}

func GetFilesFromPageIndex(httpclient *http.Client, config *utils.Config, pageidx int) ([]FolderFile, int, error) {
	FR := FolderResp{}
	tkn, err := GetToken(httpclient, config)

	if err != nil {
		return nil, 0, err
	}

	url := fmt.Sprintf("https://rapidgator.net/api/v2/folder/content?page=%d&token=%s", pageidx, tkn)
	resp, err := httpclient.Get(url)

	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&FR)

	if err != nil {
		return nil, 0, err
	}

	if FR.Status != 200 {
		return nil, 0, errors.New("rapidgator_getfilesfrompageindex error: request didn't respond with status code 200")
	}

	return FR.Response.Folder.Files, FR.Response.Pager.Total, nil
}

func FilesDeleted(httpclient *http.Client, config *utils.Config, fileids []string) (bool, error) {
	_, numofpages, err := GetFilesFromPageIndex(httpclient, config, 1)

	if err != nil {
		return false, err
	}

	for cpidx := 1; cpidx <= numofpages; cpidx++ { //cpidx == current page index
		currentpagefiles, _, err := GetFilesFromPageIndex(httpclient, config, cpidx)

		if err != nil {
			return false, err
		}
		for fididx := 0; fididx < len(fileids); fididx++ { //fididx == file ids index
			ispresent := FolderFilesExist(currentpagefiles, fileids[fididx])

			if ispresent {
				return false, nil
			}
		}
	}
	return true, nil
}
