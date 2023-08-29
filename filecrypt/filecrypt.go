package filecrypt

import (
	"errors"
	"fmt"
	"regexp"
	"reupperium/ddl"
	"reupperium/rapidgator"
	"reupperium/utils"
	"strings"
	"sync"

	"gopkg.in/resty.v1"
)

type filecrypt_error struct {
	State int    `json:"state"`
	Error string `json:"error"`
}

type DeletedFileStore struct {
	ParentContainerID   string
	ParentContainerName string
	DDLDeleted          bool
	RGDeleted           bool
	UpdatedDDLLinks     []string
	UpdatedRGLinks      []string
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

func GetDeletedContainers(rc *resty.Client, config *utils.Config) ([]DeletedFileStore, error) {
	Containers, err := GetContainers(rc, config)
	if err != nil {
		return nil, err
	}

	var ContainerRet []DeletedFileStore
	var wg sync.WaitGroup
	errCh := make(chan error, len(Containers.Containers))

	for _, FolderContainer := range Containers.Containers {
		container := FolderContainer
		wg.Add(1)
		go func() {
			defer wg.Done()
			dfstemp := DeletedFileStore{}

			dfstemp.ParentContainerName = container.Name
			dfstemp.ParentContainerID = container.ID

			MirrorContainer, err := GetContainerContents(rc, container.ID)
			if err != nil {
				errCh <- err
				return
			}

			if len(MirrorContainer.Mirrors) > 1 {
				for mirrorname, mirrorcontents := range MirrorContainer.Mirrors {
					switch mirrorname {
					case "mirror_1":
						ddlids := make([]string, 0, len(mirrorcontents.Links))
						for _, link := range mirrorcontents.Links {
							ddlextractedid, err := ExtractDDLID(link)
							if err != nil {
								errCh <- err
								return
							}
							ddlids = append(ddlids, ddlextractedid)
						}
						dfstemp.DDLDeleted, err = ddl.FilesDeleted(rc, config, ddlids)
						if err != nil {
							errCh <- err
							return
						}

					case "mirror_2":
						rgids := make([]string, 0, len(mirrorcontents.Links))
						for _, link := range mirrorcontents.Links {
							rgextractedid, err := ExtractRGID(link)
							if err != nil {
								errCh <- err
								return
							}
							rgids = append(rgids, rgextractedid)
						}
						dfstemp.RGDeleted, err = rapidgator.FilesDeleted(rc, config, rgids)
						if err != nil {
							errCh <- err
							return
						}
					}
					if dfstemp.DDLDeleted || dfstemp.RGDeleted {
						ContainerRet = append(ContainerRet, dfstemp)
						return
					}

				}
			} else {
				if strings.Contains(MirrorContainer.Mirrors["mirror_1"].Links[0], "ddownload.com") || strings.Contains(MirrorContainer.Mirrors["mirror_1"].Links[0], "ddl.to") {
					ddlids := make([]string, 0, len(MirrorContainer.Mirrors["mirror_1"].Links))
					for ddllinkidx := range MirrorContainer.Mirrors["mirror_1"].Links {
						ddlparsedid, err := ExtractDDLID(MirrorContainer.Mirrors["mirror_1"].Links[ddllinkidx])
						if err != nil {
							errCh <- err
							return
						}
						ddlids = append(ddlids, ddlparsedid)
					}
					dfstemp.DDLDeleted, err = ddl.FilesDeleted(rc, config, ddlids)
					if err != nil {
						errCh <- err
					}

				} else if strings.Contains(MirrorContainer.Mirrors["mirror_1"].Links[0], "rg.to") || strings.Contains(MirrorContainer.Mirrors["mirror_1"].Links[0], "rapidgator.net") {
					rgids := make([]string, 0, len(MirrorContainer.Mirrors["mirror_1"].Links))
					for linkidx := range MirrorContainer.Mirrors["mirror_1"].Links {
						rgparsedid, err := ExtractRGID(MirrorContainer.Mirrors["mirror_1"].Links[linkidx])
						if err != nil {
							errCh <- err
							return
						}
						rgids = append(rgids, rgparsedid)
					}
					dfstemp.RGDeleted, err = rapidgator.FilesDeleted(rc, config, rgids)
					if err != nil {
						errCh <- err
						return
					}
				}

				if dfstemp.DDLDeleted || dfstemp.RGDeleted {
					ContainerRet = append(ContainerRet, dfstemp)
				}
				return
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return nil, err
		}
	}

	return ContainerRet, nil
}
