package checker

import (
	"io"
	"net/http"
)

func textCheckFn(resp *http.Response, arg interface{}) bool {
	defer resp.Body.Close()
	body, errReadBody := io.ReadAll(resp.Body)
	if errReadBody != nil {
		return false
	}

	value, ok := arg.(string)
	if !ok {
		value = "ok"
	}

	return string(body) == value
}
