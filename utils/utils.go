package utils

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
)

type Config struct {
	Filecrypttoken string   `json:"filecrypttoken"`
	Ddltokens      []string `json:"ddltokens"`
	RapidGator     struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		Token    string `json:"token"`
	} `json:"RapidGator"`
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

func calculateMD5(filePath string, bufferSize int) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()

	buffer := make([]byte, bufferSize)
	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return "", err
		}
		if n == 0 {
			break
		}
		hash.Write(buffer[:n])
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func GetFileInfo(filepath string) (string, int64, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return "", 0, err
	}

	filesize := fileInfo.Size()
	hash, err := calculateMD5(filepath, int(filesize/10))

	if err != nil {
		return "", 0, err
	}

	return hash, filesize, nil
}
