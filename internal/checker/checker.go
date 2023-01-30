package checker

import (
	"context"
	"healthcheck/internal/config"
	"healthcheck/internal/saver"
	"log"
	"net/http"
	"reflect"
	"time"
)

type checker struct {
	cl    http.Client
	saver saver.Saver
}

func (c *checker) Check(ctx context.Context, cfg config.ConfigUrl) {
	iCheckResult := newCheckResult(cfg)
	defer c.save(iCheckResult)
	// Mark: -Do request
	req, errMakeReq := http.NewRequestWithContext(ctx, http.MethodGet, cfg.Url, nil)
	if errMakeReq != nil {
		log.Printf("errMakeReq %s: %s", cfg.Url, errMakeReq)
		iCheckResult.Fatal(errMakeReq.Error())
		return
	}

	resp, errDoReq := c.cl.Do(req)

	if errDoReq != nil {
		log.Printf("errDoReq: %s", errDoReq)
		iCheckResult.Fatal(errDoReq.Error())
		return
	}

	// MARK: -Do tests
	for _, checkLabel := range cfg.Checks {
		if fn, exist := checkFns[checkLabel]; exist {
			if !fn(resp, nil) {
				if iCheckResult.Fail(checkLabel) {
					break
				}
				continue
			}
			if iCheckResult.Ok(checkLabel) {
				break
			}
		} else {
			log.Printf("Warning! Check type %s does not exist", checkLabel)
		}
	}

	// MARK: - Notify on status change
	result, errGetLastStatus := c.saver.GetLastStatus(iCheckResult.Url)
	if errGetLastStatus != nil {
		log.Printf("errGetLastStatus: %s", errGetLastStatus)
		return
	}
	if result == nil {
		return
	}
	if *result != iCheckResult.isPassed() {
		go c.notifyOnStatusChanged(ctx, iCheckResult.Url, iCheckResult.isPassed())
	}
}

func (c *checker) notifyOnStatusChanged(ctx context.Context, url string, state bool) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://httpbin.org/post", nil)
	if err != nil {
		log.Printf("notify state change: %s", err)
		return
	}
	// @note тут была бы куча кода с ретраями, но в задании такого не было. Я думаю, смысл ясен.
	c.cl.Do(req)
}

func (c *checker) save(chr *CheckResult) {
	if errSaveResult := c.saver.Save(
		chr.Url,
		chr.isPassed(),
		chr.checksFailed,
		chr.errorMessage,
	); errSaveResult != nil {
		log.Printf("errSaveResult: %s", errSaveResult)
	}

}

func NewChecker(httpTimeoutSec int, s saver.Saver) *checker {
	var checkr checker
	if reflect.ValueOf(s).IsNil() {
		checkr.saver = saver.StdoutSaver{}
	} else {
		checkr.saver = s
	}
	checkr.cl = http.Client{
		Timeout: time.Duration(httpTimeoutSec * int(time.Second)),
	}

	return &checkr
}
