package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
	"reupperium/filecrypt"
	"sync"

	"gopkg.in/resty.v1"
)

type Config struct {
	Filecrypttoken string `json:"filecrypttoken"`
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

func main() {
	client := resty.New()
	configfile, err := GetConfig()

	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		filecrypt.Initialize(client, configfile.Filecrypttoken)
	}()

	wg.Wait()
}
