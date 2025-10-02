package data

import (
	"fmt"
	"strconv"
)

type Runtime int32

func (r Runtime) MarshalJSON() ([]byte, error) {

	js := fmt.Sprintf(`%d mins`, r)
	return []byte(strconv.Quote(js)), nil

}
