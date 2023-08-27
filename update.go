package main

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reupperium/ddl"
	"reupperium/filecrypt"
	"reupperium/rapidgator"
	"reupperium/utils"
	"strings"
	"sync"
	"time"

	"gopkg.in/resty.v1"
)

type SeriesMirror struct {
	DDLLinks        []string
	RapidGatorLinks []string
}
type Series struct {
	EpisodeMirrors SeriesMirror
}

type SeriesLibrary struct {
	Container map[string]Series
}

func HandleFileUpload(rc *resty.Client, httpclient *http.Client, config *utils.Config, path string) (string, string, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return "", "", err
	}

	timestamp := time.Now().Unix()
	timestampBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(timestampBytes, uint32(timestamp))

	_, err = file.Write(timestampBytes)
	if err != nil {
		return "", "", errors.New("failed to write timestamp")
	}
	file.Close()

	for attempt := 1; attempt <= config.MaxCopyRetries; attempt++ {
		ddllink, err := ddl.UploadFile(rc, httpclient, path)
		if err != nil {
			ddl.Log_Error(err.Error())
			time.Sleep(time.Second * 5)
			continue
		}

		rglink, err := rapidgator.UploadFile(rc, path)
		if err != nil {
			rapidgator.Log_Error(err.Error())
			time.Sleep(time.Second * 5)
			continue
		}

		return ddllink, rglink, nil
	}
	return "", "", fmt.Errorf("upload failed after %d attempts", config.MaxCopyRetries)
}

func processFilesInDirectory(rc *resty.Client, httpclient *http.Client, config *utils.Config, sourceDir string, tempDir string) (SeriesLibrary, error) {
	var lib SeriesLibrary
	lib.Container = make(map[string]Series)

	var rootwg sync.WaitGroup
	currentuploads := 0
	err := filepath.Walk(sourceDir, func(filePath string, fileInfo os.FileInfo, err error) error {
		if currentuploads == config.MaxConcurrentFilesUpload {
			rootwg.Wait()
		}
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".rar") {
			go func() {
				defer rootwg.Done()
				tempFilePath := filepath.Join(tempDir, fileInfo.Name())

				err := utils.CopyFileWithRetryAndVerification(filePath, tempFilePath, config.MaxCopyRetries)
				if err != nil {
					fmt.Println("Error copying file:", err)
				}

				dirName := filepath.Dir(filePath)
				dirBaseName := filepath.Base(dirName)

				ddllink, rglink, err := HandleFileUpload(rc, httpclient, config, tempFilePath)

				if err != nil {
					fmt.Printf("HandleFileUpload error: %s\n", err)
				}

				if obj, ok := lib.Container[dirBaseName]; ok {
					obj.EpisodeMirrors.DDLLinks = append(obj.EpisodeMirrors.DDLLinks, ddllink)
					obj.EpisodeMirrors.RapidGatorLinks = append(obj.EpisodeMirrors.RapidGatorLinks, rglink)
					lib.Container[dirBaseName] = obj
				} else {
					newobj := Series{
						EpisodeMirrors: SeriesMirror{
							DDLLinks:        []string{ddllink},
							RapidGatorLinks: []string{rglink},
						},
					}
					lib.Container[dirBaseName] = newobj
				}

				os.Remove(tempFilePath)
				currentuploads -= 1
			}()
			currentuploads += 1
			rootwg.Add(1)
		}
		return nil
	})
	rootwg.Wait()

	if err != nil {
		return SeriesLibrary{}, err
	}
	return lib, nil
}

func UpdateCheck(rc *resty.Client, httpclient *http.Client) error {
	config, err := utils.GetConfig()
	libraryarr := []SeriesLibrary{}

	if err != nil {
		return err
	}

	filecrypt.Log("Getting Filecrypt IDs")

	ddlids, rgids, err := filecrypt.GetIDS(rc)

	if err != nil {
		return err
	}

	ddldeleted, err := ddl.FilesDeleted(rc, ddlids)

	if err != nil {
		return err
	}
	rgdeleted, err := rapidgator.FilesDeleted(rc, rgids)

	if err != nil {
		return err
	}

	if ddldeleted || rgdeleted {
		fmt.Println("Files Deleted Running Reupload Routine")

		cd, err := os.Getwd()
		if err != nil {
			return err
		}

		tempDir, err := os.MkdirTemp(cd, "temp")
		if err != nil {
			return err
		}

		numdirectoriesuploading := 0
		var duwg sync.WaitGroup

		for directoryidx := range config.MediaPaths {
			dirIdxCopy := directoryidx

			if numdirectoriesuploading == config.MaxConcurrentFoldersUpload {
				duwg.Wait()
			}

			go func() {
				defer duwg.Done()
				library, err := processFilesInDirectory(rc, httpclient, &config, config.MediaPaths[dirIdxCopy], tempDir)
				if err != nil {
					fmt.Printf("Error Processing Files In Directory: %s\n", err)
				}
				libraryarr = append(libraryarr, library)
				numdirectoriesuploading -= 1
			}()
			numdirectoriesuploading += 1
			duwg.Add(1)
		}
		duwg.Wait()

		err = os.Remove(tempDir)
		if err != nil {
			fmt.Println("Error deleting file:", err)
		}

		libjson, _ := json.Marshal(libraryarr)
		fmt.Println(string(libjson))
	}
	return nil
}
