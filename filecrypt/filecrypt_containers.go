package filecrypt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reupperium/utils"
)

type Mirror struct {
	Links  []string `json:"links"`
	Backup []string `json:"backup"`
}

type MirrorContainer struct {
	Mirrors map[string]Mirror `json:"container"`
	State   int               `json:"state,omitempty"`
	Error   string            `json:"error,omitempty"`
}

type EditContainerResp struct {
	Container struct {
		Link        string   `json:"link"`
		Name        string   `json:"name"`
		Tags        []string `json:"tags"`
		Hoster      []string `json:"hoster"`
		Size        int      `json:"size"`
		SizeHuman   string   `json:"size_human"`
		Links       int      `json:"links"`
		Status      int      `json:"status"`
		StatusimgID string   `json:"statusimg_id"`
		Created     int      `json:"created"`
		Edited      int      `json:"edited"`
		Views       struct {
			Today int `json:"today"`
			Week  int `json:"week"`
			All   int `json:"all"`
		} `json:"views"`
	} `json:"container"`
	State int `json:"state"`
}

func GetContainerContents(httpclient *http.Client, config *utils.Config, id string) (MirrorContainer, error) {
	GCC_Ret := MirrorContainer{}

	request, err := http.NewRequest("POST", fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=info&container_id=%s", config.Filecrypttoken, id), nil)
	if err != nil {
		return GCC_Ret, err
	}

	response, err := httpclient.Do(request)
	if err != nil {
		return GCC_Ret, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&GCC_Ret)

	if err != nil {
		return GCC_Ret, err
	}

	if len(GCC_Ret.Mirrors) == 0 {
		return GCC_Ret, fmt.Errorf("failed to get container contents: %s", GCC_Ret.Error)
	}

	return GCC_Ret, nil
}

func EditContainer(httpclient *http.Client, config *utils.Config, deletedcontainer *DeletedFileStore) error {
	queryValues := url.Values{}

	linksToUpdate := make(map[string][]string)

	container, err := GetContainerContents(httpclient, config, deletedcontainer.ParentContainerID)
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

	EditContainerRet := EditContainerResp{}

	request, err := http.NewRequest("POST", fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=editV2&container_id=%s&%s", config.Filecrypttoken, deletedcontainer.ParentContainerID, queryValues.Encode()), nil)
	if err != nil {
		return err
	}

	response, err := httpclient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&EditContainerRet)

	if err != nil {
		return err
	}

	if EditContainerRet.State == 0 {
		return fmt.Errorf("failed to edit filecrypt container state: %d", EditContainerRet.State)
	}

	return nil
}
