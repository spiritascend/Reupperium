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

func EditContainer(rc *resty.Client, config *utils.Config, deletedcontainer *DeletedFileStore) error {
	queryValues := url.Values{}

	linksToUpdate := make(map[string][]string)

	container, err := GetContainerContents(rc, deletedcontainer.ParentContainerID)
	if err != nil {
		return err
	}

	if len(container.Mirrors) > 1 {
		if deletedcontainer.DDLDeleted {
			linksToUpdate["mirror_1"] = deletedcontainer.UpdatedDDLLinks
			linksToUpdate["mirror_2"] = container.Mirrors["mirror_2"].Links
		}
		if deletedcontainer.RGDeleted {
			linksToUpdate["mirror_1"] = container.Mirrors["mirror_1"].Links
			linksToUpdate["mirror_2"] = deletedcontainer.UpdatedRGLinks
		}
	} else {
		if deletedcontainer.DDLDeleted {
			linksToUpdate["mirror_1"] = deletedcontainer.UpdatedDDLLinks
		}
		if deletedcontainer.RGDeleted {
			linksToUpdate["mirror_1"] = deletedcontainer.UpdatedRGLinks
		}
	}

	for mirror, links := range linksToUpdate {
		for idx, link := range links {
			queryKey := fmt.Sprintf("%s[0][%d]", mirror, idx)
			queryValues.Add(queryKey, link)
		}
	}

	url := fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=editV2&container_id=%s&%s", config.Filecrypttoken, deletedcontainer.ParentContainerID, queryValues.Encode())
	resp, err := rc.R().Post(url)
	if err != nil {
		return err
	}

	var fcErr filecrypt_error
	if err := json.Unmarshal(resp.Body(), &fcErr); err != nil {
		return err
	}

	if fcErr.State == 0 {
		return fmt.Errorf("filecrypt error: %s", fcErr.Error)
	}

	return nil
}
