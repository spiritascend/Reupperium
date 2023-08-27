package filecrypt

import (
	"errors"
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
	re := regexp.MustCompile(`file/(.{32})`)
	match := re.FindStringSubmatch(url)
	if len(match) < 2 {
		return "", fmt.Errorf("regex failed to parse rapidgator id")
	}
	return match[1], nil
}

func ExtractDDLID(url string) (string, error) {
	var re *regexp.Regexp
	var suffix string

	switch {
	case strings.Contains(url, "ddownload.com/"):
		re = regexp.MustCompile(`\.com/([a-zA-Z0-9]+)`)
		suffix = "ddownload"
	case strings.Contains(url, "ddl.to/d/"):
		re = regexp.MustCompile(`d/([a-zA-Z0-9]+)`)
		suffix = "ddl"
	default:
		return "", errors.New("failed to parse ddl id due to incorrect URL")
	}

	match := re.FindStringSubmatch(url)
	if len(match) < 2 {
		return "", fmt.Errorf("regex failed to parse %s id", suffix)
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

		if len(cont.Mirrors) > 1 {
			for mirrorname, links := range cont.Mirrors {

				if mirrorname == "mirror_1" {
					for linkidx := range links.Links {
						linkparsedid, err := ExtractDDLID(links.Links[linkidx])
						if err != nil {
							return nil, nil, err
						}
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
		} else {
			if strings.Contains(cont.Mirrors["mirror_1"].Links[0], "ddownload.com") || strings.Contains(cont.Mirrors["mirror_1"].Links[0], "ddl.to") {
				for linkidx := range cont.Mirrors["mirror_1"].Links {
					linkparsedid, err := ExtractDDLID(cont.Mirrors["mirror_1"].Links[linkidx])
					if err != nil {
						return nil, nil, err
					}
					ddllinks = append(ddllinks, linkparsedid)
				}
			} else if strings.Contains(cont.Mirrors["mirror_1"].Links[0], "rg.to") || strings.Contains(cont.Mirrors["mirror_1"].Links[0], "rapidgator.net") {
				for linkidx := range cont.Mirrors["mirror_1"].Links {
					linkparsedid, err := ExtractRGID(cont.Mirrors["mirror_1"].Links[linkidx])
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
