package rapidgator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reupperium/utils"
)

type LoginResp struct {
	Response struct {
		Token string `json:"token"`
		User  struct {
			Email          string `json:"email"`
			IsPremium      bool   `json:"is_premium"`
			PremiumEndTime any    `json:"premium_end_time"`
			State          int    `json:"state"`
			StateLabel     string `json:"state_label"`
			Traffic        struct {
				Total any `json:"total"`
				Left  any `json:"left"`
			} `json:"traffic"`
			Storage struct {
				Total string `json:"total"`
				Left  int64  `json:"left"`
			} `json:"storage"`
			Upload struct {
				MaxFileSize int64 `json:"max_file_size"`
				NbPipes     int   `json:"nb_pipes"`
			} `json:"upload"`
			RemoteUpload struct {
				MaxNbJobs   int `json:"max_nb_jobs"`
				RefreshTime int `json:"refresh_time"`
			} `json:"remote_upload"`
		} `json:"user"`
	} `json:"response"`
	Status  int `json:"status"`
	Details any `json:"details"`
}

type UserInfo struct {
	Response struct {
		User struct {
			Email          string `json:"email"`
			IsPremium      bool   `json:"is_premium"`
			PremiumEndTime any    `json:"premium_end_time"`
			State          int    `json:"state"`
			StateLabel     string `json:"state_label"`
			Traffic        struct {
				Total any `json:"total"`
				Left  any `json:"left"`
			} `json:"traffic"`
			Storage struct {
				Total string `json:"total"`
				Left  int64  `json:"left"`
			} `json:"storage"`
			Upload struct {
				MaxFileSize int64 `json:"max_file_size"`
				NbPipes     int   `json:"nb_pipes"`
			} `json:"upload"`
			RemoteUpload struct {
				MaxNbJobs   int `json:"max_nb_jobs"`
				RefreshTime int `json:"refresh_time"`
			} `json:"remote_upload"`
		} `json:"user"`
	} `json:"response"`
	Status  int `json:"status"`
	Details any `json:"details"`
}

func IsAuthenticated(httpclient *http.Client, token string) (bool, error) {
	var Resp UserInfo

	url := fmt.Sprintf("https://rapidgator.net/api/v2/user/info?token=%s", token)
	resp, err := httpclient.Get(url)

	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&Resp)

	if err != nil {
		return false, err
	}

	return Resp.Status != 401, nil
}

func RefreshToken(httpclient *http.Client, username string, password string) (string, error) {
	var Resp LoginResp
	url := fmt.Sprintf("https://rapidgator.net/api/v2/user/login?login=%s&password=%s&code=000000", username, password)
	resp, err := httpclient.Get(url)

	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&Resp)

	if err != nil {
		return "", err
	}

	if Resp.Status == 401 {
		return "", errors.New("failed to authenticate rg_refreshtoken return status 401")
	}

	return Resp.Response.Token, nil
}

func GetToken(httpclient *http.Client, config *utils.Config) (string, error) {
	isauthed, err := IsAuthenticated(httpclient, config.RapidGator.Token)

	if err != nil {
		return "", err
	}

	if !isauthed {
		config.RapidGator.Token, err = RefreshToken(httpclient, config.RapidGator.Email, config.RapidGator.Password)

		if err != nil {
			return "", err
		}

		err = utils.OverwriteConfig(*config)

		if err != nil {
			return "", err
		}
	}
	return config.RapidGator.Token, nil
}
