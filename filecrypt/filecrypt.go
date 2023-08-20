package filecrypt

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/resty.v1"
)

type filecrypt_error struct {
	State int    `json:"state"`
	Error string `json:"error"`
}

func ExtractRGID(url string) (string, error) {
	re := regexp.MustCompile(`/file/(\d+)/`)
	match := re.FindStringSubmatch(url)
	if len(match) < 2 {
		return "", fmt.Errorf("regex failed to parse rapidgator id")
	}
	return match[1], nil
}

func GetIDS(rc *resty.Client) ([]string, []string, error) { // ddl,rapidgator
	ddllinks := []string{}
	rapidgatorlinks := []string{}
	Containers, _ := GetContainers(rc)

	for _, containerdetails := range Containers.Containers {
		cont, err := GetContainerContents(rc, containerdetails.ID)

		if err != nil {
			Log_Error(err.Error())
			return nil, nil, err
		}

		for mirrorname, links := range cont.Mirrors {

			if mirrorname == "mirror_1" {
				for linkidx := range links.Links {
					linkparsedid := strings.TrimPrefix(links.Links[linkidx], "http://ddl.to/d/")
					ddllinks = append(ddllinks, linkparsedid)
				}
			}
			if mirrorname == "mirror_2" {
				for linkidx := range links.Links {
					linkparsedid, err := ExtractRGID(links.Links[linkidx])

					if err != nil {
						return nil, nil, err
					}
					rapidgatorlinks = append(rapidgatorlinks, linkparsedid)
				}
			}
		}
	}
	return ddllinks, rapidgatorlinks, nil
}

func Initialize(rc *resty.Client) {

	ddllinks, _, err := GetIDS(rc)

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(ddllinks)

}
