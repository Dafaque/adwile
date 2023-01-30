package checker

import "net/http"

func statusCodeCheckFn(resp *http.Response, arg interface{}) bool {
	value, ok := arg.(int)
	if !ok {
		value = http.StatusOK
	}
	return resp.StatusCode == value
}
