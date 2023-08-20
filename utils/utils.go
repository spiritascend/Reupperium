package utils

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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
	RootPath string `json:"RootPath"`
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

func copyFile(src, dst string, wg *sync.WaitGroup, errChan chan<- error) {
	defer wg.Done()

	sourceFile, err := os.Open(src)
	if err != nil {
		errChan <- err
		return
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		errChan <- err
		return
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		errChan <- err
		return
	}

	oldhash, err := calculateMD5(src)
	if err != nil {
		errChan <- err
		return
	}

	newhash, err := calculateMD5(dst)
	if err != nil {
		errChan <- err
		return
	}

	if oldhash != newhash {
		errChan <- fmt.Errorf("file_copy_hash_mismatch oldhash: %s | newhash: %s", oldhash, newhash)
	}
}

func CopyAll() (string, error) {

	config, err := GetConfig()

	if err != nil {
		return "", err
	}

	cd, err := os.Getwd()

	if err != nil {
		return "", err
	}

	tempDir, err := os.MkdirTemp(cd, "temp")
	if err != nil {
		return "", err
	}

	var wg sync.WaitGroup
	errChan := make(chan error, 1)

	err = filepath.Walk(config.RootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(info.Name(), ".rar") {
			wg.Add(1)
			go copyFile(path, filepath.Join(tempDir, info.Name()), &wg, errChan)
		}
		return nil
	})

	if err != nil {
		return "", err
	} else {
		wg.Wait()
		close(errChan)
		for err := range errChan {
			if err != nil {
				return "", err
			}
			fmt.Println(err)
		}
		return tempDir, nil
	}
}
