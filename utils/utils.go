package utils

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"golang.org/x/sys/windows/registry"
)

func GetWindowsProxy() (string, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Microsoft\Windows\CurrentVersion\Internet Settings`, registry.READ)
	if err != nil {
		return "", err
	}
	defer k.Close()

	proxyServer, _, err := k.GetStringValue("ProxyServer")
	if err != nil {
		return "", err
	}

	regexPattern := `[a-z]+=([0-9.:]+)`

	re := regexp.MustCompile(regexPattern)
	matches := re.FindAllStringSubmatch(proxyServer, -1)

	if len(matches) > 0 && len(matches[0]) > 1 {
		return matches[0][1], nil
	}

	return "", nil
}

type Config struct {
	Filecrypttoken string   `json:"filecrypttoken"`
	Ddltokens      []string `json:"ddltokens"`
	RapidGator     struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Token    string `json:"token"`
		Cookie   struct {
			Lang     string `json:"Lang"`
			UserInfo string `json:"UserInfo"`
			Session  string `json:"Session"`
			Token    string `json:"Token"`
		} `json:"cookie"`
	} `json:"RapidGator"`
	MediaPaths                    []string `json:"MediaPaths"`
	MaxCopyRetries                int      `json:"MaxCopyRetries"`
	MaxConcurrentFilesUpload      int      `json:"MaxConcurrentFilesUpload"`
	MaxConcurrentFoldersUpload    int      `json:"MaxConcurrentFoldersUpload"`
	TimeBeforeNextDeletedCheckMs  int      `json:"TimeBeforeNextDeletedCheckMs"`
	CPUProfilePort                string   `json:"CPUProfilePort"`
	MaxTimeoutUploadBeforeRetryMs int      `json:"MaxTimeoutUploadBeforeRetryMs"`
}

func GetConfig() (Config, error) {
	var Ret Config

	rawconfig, err := os.Open("config.json")
	if err != nil {
		log.Fatal(err)
		return Config{}, errors.New("failed to open config file")
	}
	defer rawconfig.Close()

	decoder := json.NewDecoder(rawconfig)
	err = decoder.Decode(&Ret)

	if err != nil {
		return Config{}, errors.New("failed to decode config file")
	}

	return Ret, nil
}

func OverwriteConfig(config Config) error {
	file, err := os.Create("config.json")
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "    ")

	if err := encoder.Encode(config); err != nil {
		return err
	}
	return nil
}

func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GetFileInfo(filepath string) (string, string, int64, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return "", "", 0, err
	}

	filesize := fileInfo.Size()

	hash, err := calculateMD5(filepath)

	if err != nil {
		return "", "", 0, err
	}

	return fileInfo.Name(), hash, filesize, nil
}

func ExtractDirectoryName(fullDirPath string) (string, error) {
	regex := regexp.MustCompile(`[^\\]+$`)

	match := regex.FindString(fullDirPath)
	if match == "" {
		return "", errors.New("directory name not found")
	}
	return match, nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func CopyFileWithRetryAndVerification(src, dst string, maxRetries int) error {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		err := copyFile(src, dst)
		if err == nil {
			if err := compareFileContents(src, dst); err == nil {
				return nil
			} else {
				fmt.Printf("Attempt %d: Copied file contents do not match\n", attempt)
			}
		} else {
			fmt.Printf("Attempt %d failed: %s\n", attempt, err)
		}
	}

	return fmt.Errorf("exceeded maximum retries")
}

func compareFileContents(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Open(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	sourceBuffer := make([]byte, 5)
	destBuffer := make([]byte, 5)

	_, err = sourceFile.Read(sourceBuffer)
	if err != nil && err != io.EOF {
		return err
	}

	_, err = destFile.Read(destBuffer)
	if err != nil && err != io.EOF {
		return err
	}

	_, err = sourceFile.Seek(-5, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = destFile.Seek(-5, io.SeekEnd)
	if err != nil {
		return err
	}

	_, err = sourceFile.Read(sourceBuffer)
	if err != nil && err != io.EOF {
		return err
	}

	_, err = destFile.Read(destBuffer)
	if err != nil && err != io.EOF {
		return err
	}

	if !bytes.Equal(sourceBuffer, destBuffer) {
		return fmt.Errorf("first and last 5 bytes of file contents do not match")
	}

	return nil
}

func SearchFolder(rootPaths []string, targetFolder string) (string, bool) {
	var wg sync.WaitGroup
	resultChan := make(chan string, len(rootPaths))

	for _, root := range rootPaths {
		wg.Add(1)
		go func(rootPath string) {
			defer wg.Done()
			err := filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if info.IsDir() && strings.EqualFold(info.Name(), targetFolder) {
					resultChan <- path
				}
				return nil
			})

			if err != nil {
				fmt.Println("Error:", err)
			}
		}(root)
	}

	wg.Wait()
	close(resultChan)

	select {
	case result := <-resultChan:
		return result, true
	default:
		return "", false
	}
}

func SearchFolderV2(rootPaths []string, targetFolder string) (string, bool) {
	for rootpathsidx := range rootPaths {
		_, err := os.Stat(filepath.Join(rootPaths[rootpathsidx], targetFolder))
		if err != nil && os.IsNotExist(err) {
			continue
		} else {
			return filepath.Join(rootPaths[rootpathsidx], targetFolder), true
		}
	}
	return "", false
}
