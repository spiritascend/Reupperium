package filecrypt

import (
	"gopkg.in/resty.v1"
)

type filecrypt_error struct {
	State int    `json:"state"`
	Error string `json:"error"`
}

func Initialize(rc *resty.Client, tkn string) {
	Log("Got Filecrypt Token " + tkn)

	err := EditContainer(rc, tkn, "677260D89C", []string{"http://ddl.to/d/4QVRY", "http://ddl.to/d/4QVRT"})

	if err != nil {
		Log_Error(err.Error())
		return
	}

}
