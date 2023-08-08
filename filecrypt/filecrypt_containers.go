package filecrypt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"gopkg.in/resty.v1"
)

type Container_Root struct {
	Container Container `json:"container"`
}

type Container struct {
	Mirror1 Mirror `json:"mirror_1"`
}

type Mirror struct {
	Links  []string `json:"links"`
	Backup []string `json:"backup"`
}

func GetContainerContents(rc *resty.Client, token string, id string) (*Container_Root, error) {
	var GCC_Ret Container_Root

	resp, err := rc.R().Post(fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=info&container_id=%s", token, id))
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resp.Body(), &GCC_Ret); err != nil {
		return nil, err
	}

	var fc_err filecrypt_error
	if err := json.Unmarshal(resp.Body(), &fc_err); err != nil {
		return nil, err
	}

	if fc_err.State == 0 {
		return nil, errors.New(fc_err.Error)
	}

	return &GCC_Ret, nil
}

func EditContainer(rc *resty.Client, token string, container_id string, links []string) error {

	queryValues := url.Values{}

	for i, mirror := range links {
		queryValues.Add(fmt.Sprintf("mirror_1[0][%d]", i), mirror)
	}

	queryString := queryValues.Encode()

	resp, err := rc.R().Post(fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=editV2&container_id=%s&%s", token, container_id, queryString))
	if err != nil {
		return err
	}

	var fc_err filecrypt_error
	if err := json.Unmarshal(resp.Body(), &fc_err); err != nil {
		return err
	}

	if fc_err.State == 0 {
		return errors.New(fc_err.Error)
	}

	return nil
}
