package saver

import (
	"fmt"
	"log"
	"strings"
)

type StdoutSaver struct{}

func (ss StdoutSaver) Save(url string, passed bool, failed []string, errMsg string) error {
	var result string = url
	if passed {
		result += " ok"
	} else {
		result += " fail"
		result += fmt.Sprintf(" (%s)", strings.Join(failed, ","))
	}

	if len(errMsg) > 0 {
		result += fmt.Sprintf("error: %s", errMsg)
	}

	log.Println(result)

	return nil
}

func (ss StdoutSaver) GetLastStatus(url string) (*bool, error) {
	return nil, nil
}
