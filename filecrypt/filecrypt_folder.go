package filecrypt

import (
	"encoding/json"
	"errors"
	"fmt"

	"gopkg.in/resty.v1"
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

func GetContainers(rc *resty.Client, token string) (*Folder, error) {
	var GCRet Folder
	resp, err := rc.R().Post(fmt.Sprintf("http://filecrypt.cc/api.php?api_key=%s&fn=containerV2&sub=myfolder", token))

	if err != nil {
		Log_Error(err.Error())
		return nil, err
	}

	if err := json.Unmarshal(resp.Body(), &GCRet); err != nil {
		Log_Error(err.Error())
		return nil, err
	}

	var fc_err filecrypt_error
	if err := json.Unmarshal(resp.Body(), &fc_err); err != nil {
		return nil, err
	}

	if fc_err.State == 0 {
		return nil, errors.New(fc_err.Error)
	}

	return &GCRet, nil
}
