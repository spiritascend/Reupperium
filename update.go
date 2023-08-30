package main

import (
	"encoding/binary"
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

func HandleFileUpload(rc *resty.Client, httpclient *http.Client, config *utils.Config, path string, deletedddl bool, deletedrg bool) (string, string, error) {
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
		var ddllink, rglink string

		if deletedddl {
			ddllink, err = ddl.UploadFile(rc, httpclient, path)
			if err != nil {
				ddl.Log_Error(err.Error())
				time.Sleep(time.Second * 5)
				continue
			}
		}

		if deletedrg {
			rglink, err = rapidgator.UploadFile(rc, config, path)
			if err != nil {
				rapidgator.Log_Error(err.Error())
				time.Sleep(time.Second * 5)
				continue
			}
		}

		return ddllink, rglink, nil
	}
	return "", "", fmt.Errorf("upload failed after %d attempts", config.MaxCopyRetries)
}

func ProcessDirectory(rc *resty.Client, httpclient *http.Client, DeletedContainer *filecrypt.DeletedFileStore, config *utils.Config, directorypath string, temppath string) error {
	dir, err := os.Open(directorypath)
	if err != nil {
		return err
	}
	defer dir.Close()

	fileInfos, err := dir.Readdir(-1)
	if err != nil {
		return err
	}

	var fcwg sync.WaitGroup
	var fumtx sync.Mutex
	semaphore := make(chan struct{}, config.MaxConcurrentFilesUpload)

	errCh := make(chan error, len(fileInfos))

	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".rar") {
			filename := fileInfo.Name()

			semaphore <- struct{}{}

			fcwg.Add(1)
			go func() {
				defer fcwg.Done()
				defer func() { <-semaphore }()

				tempfilepath := filepath.Join(temppath, filename)
				srcfilepath := filepath.Join(directorypath, filename)

				if err := utils.CopyFileWithRetryAndVerification(srcfilepath, tempfilepath, config.MaxCopyRetries); err != nil {
					errCh <- err
					return
				}

				ddllink, rglink, err := HandleFileUpload(rc, httpclient, config, tempfilepath, DeletedContainer.DDLDeleted, DeletedContainer.RGDeleted)
				if err != nil {
					errCh <- fmt.Errorf("handlefileupload error: %s", err)
				} else {
					fumtx.Lock()
					DeletedContainer.UpdatedDDLLinks = append(DeletedContainer.UpdatedDDLLinks, ddllink)
					DeletedContainer.UpdatedRGLinks = append(DeletedContainer.UpdatedRGLinks, rglink)
					fumtx.Unlock()
				}

				if err := os.Remove(tempfilepath); err != nil {
					fmt.Println(err)
					errCh <- err
				}
			}()
		}
	}

	fcwg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

func UpdateCheckV2(rc *resty.Client, httpclient *http.Client) error {
	config, err := utils.GetConfig()

	if err != nil {
		fmt.Println(err)
	}

	deletedcontainers, err := filecrypt.GetDeletedContainers(rc, &config)

	if err != nil {
		fmt.Println(err)
	}

	if len(deletedcontainers) > 0 {
		fmt.Println("Deleted Files Detected")

		cd, err := os.Getwd()
		if err != nil {
			fmt.Println("Failed To Get Current Directory")
		}

		tempDir, err := os.MkdirTemp(cd, "temp")
		if err != nil {
			fmt.Printf("Failed to Create Temp Directory %s", err)
		}
		defer os.RemoveAll(tempDir)

		for deletedcontaineridx := range deletedcontainers {
			folderpath, bffsuccessful := utils.SearchFolderV2(config.MediaPaths, deletedcontainers[deletedcontaineridx].ParentContainerName)

			if !bffsuccessful {
				fmt.Printf("Failed To Find Folder: %s\n", deletedcontainers[deletedcontaineridx].ParentContainerName)
				continue
			}

			err = ProcessDirectory(rc, httpclient, &deletedcontainers[deletedcontaineridx], &config, folderpath, tempDir)
			if err != nil {
				return err
			}

			err = filecrypt.EditContainer(rc, &config, &deletedcontainers[deletedcontaineridx])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
