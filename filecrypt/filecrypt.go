package filecrypt

import (
	"fmt"

	"gopkg.in/resty.v1"
)

type filecrypt_error struct {
	State int    `json:"state"`
	Error string `json:"error"`
}

func Initialize(rc *resty.Client, tkn string) {
	Log("Got Filecrypt Token " + tkn)

	cont, err := GetContainerContents(rc, tkn, "677260D89C")

	if err != nil {
		Log_Error(err.Error())
		return
	}

	fmt.Println(len(cont.Mirrors))

}
