package ddl

import (
	"fmt"

	"gopkg.in/resty.v1"
)

func Initialize(rc *resty.Client, tkns []string) {
	//Log("Got ddl Token " + tkn)

	isdeleted, err := FilesDeleted(rc, tkns, []string{"07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4", "07gw2rutlhqi", "c54de7itfte2", "J342r089ufd3jn4"})

	if err != nil {
		Log_Error(err.Error())
	}

	fmt.Printf("Need to update: %v\n", isdeleted)

}
