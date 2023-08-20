package filecrypt

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reupperium/utils"

	"gopkg.in/resty.v1"
)

type Mirror struct {
	Links  []string `json:"links"`
	Backup []string `json:"backup"`
}

type MirrorContainer struct {
	Mirrors map[string]Mirror `json:"container"`
}

func GetContainerContents(rc *resty.Client, id string) (MirrorContainer, error) {
	config, err := utils.GetConfig()

	if err != nil {
		return MirrorContainer{}, err
	}

	var GCC_Ret MirrorContainer

	resp, err := rc.R().Post(fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=info&container_id=%s", config.Filecrypttoken, id))
	if err != nil {
		return MirrorContainer{}, err
	}

	if err := json.Unmarshal(resp.Body(), &GCC_Ret); err != nil {
		return MirrorContainer{}, err
	}

	var fc_err filecrypt_error
	if err := json.Unmarshal(resp.Body(), &fc_err); err != nil {
		return MirrorContainer{}, err
	}

	if fc_err.State == 0 && len(fc_err.Error) > 0 {
		return MirrorContainer{}, errors.New(fc_err.Error)
	}

	return GCC_Ret, nil
}

func EditContainer(rc *resty.Client, container_id string, mirror_type string, links []string) error {
	config, err := utils.GetConfig()

	if err != nil {
		return err
	}
	queryValues := url.Values{}

	for i, mirror := range links {

		if mirror_type == "mirror_1" { // ddl
			queryValues.Add(fmt.Sprintf("mirror_1[0][%d]", i), mirror)
			continue
		}
		if mirror_type == "mirror_2" { // rapidgator
			queryValues.Add(fmt.Sprintf("mirror_2[0][%d]", i), mirror)
			continue
		}

		if i < len(links)-1 {
			return errors.New("invalid_mirror_type_edit")
		}
	}

	resp, err := rc.R().Post(fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=editV2&container_id=%s&%s", config.Filecrypttoken, container_id, queryValues.Encode()))
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
