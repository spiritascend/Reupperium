package rapidgator

import (
	"fmt"

	"gopkg.in/resty.v1"
)

func Initialize(rc *resty.Client) (string, error) {

	deleted, err := FilesDeleted(rc, []string{"75fe1a0043dc60d4605dfb5ac10b9c1b"})

	if err != nil {
		fmt.Println(err)
		return "", err
	}

	fmt.Println(deleted)

	return "", nil
}
