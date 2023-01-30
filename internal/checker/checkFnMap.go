package checker

import "net/http"

type checkFn func(*http.Response, interface{}) bool

var checkFns map[string]checkFn = make(map[string]checkFn, 2)

func init() {
	checkFns["status_code"] = statusCodeCheckFn
	checkFns["text"] = textCheckFn
}
