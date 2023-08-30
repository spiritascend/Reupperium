package filecrypt

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reupperium/utils"
)

type Folder_Container struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Group  int    `json:"group"`
}

type Folder struct {
	State      int                         `json:"state,omitempty"`
	Error      string                      `json:"error,omitempty"`
	Containers map[string]Folder_Container `json:"container,omitempty"`
}

func GetContainers(httpclient *http.Client, config *utils.Config) (Folder, error) {
	GCRet := Folder{}

	request, err := http.NewRequest("POST", fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=myfolder", config.Filecrypttoken), nil)
	if err != nil {
		return GCRet, err
	}

	response, err := httpclient.Do(request)
	if err != nil {
		return GCRet, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&GCRet)

	if err != nil {
		return GCRet, err
	}

	if GCRet.State == 0 {
		return Folder{}, fmt.Errorf("failed to get ddl containers because state is %d", GCRet.State)
	}

	for containername, container := range GCRet.Containers {
		if container.Status == "0" || container.Status == "1" || container.Status == "2" || container.Status == "3" {
			delete(GCRet.Containers, containername)
		}
	}
	return GCRet, nil
}
