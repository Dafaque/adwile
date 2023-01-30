package checker

import (
	"healthcheck/internal/config"
)

type CheckResult struct {
	config.ConfigUrl
	checksOkCount int
	checksFailed  []string
	checksOk      []string
	errorMessage  string
}

func (chr *CheckResult) isPassed() bool {
	return chr.checksOkCount == chr.MinChecksCount
}

func (chr *CheckResult) isShouldStop() bool {
	return chr.isPassed() || len(chr.Checks)-len(chr.checksFailed) < chr.MinChecksCount
}

func (chr *CheckResult) Ok(label string) bool {
	chr.checksOk = append(chr.checksOk, label)
	chr.checksOkCount++
	return chr.isShouldStop()
}

func (chr *CheckResult) Fail(label string) bool {
	chr.checksFailed = append(chr.checksFailed, label)
	return chr.isShouldStop()
}

func (chr *CheckResult) Fatal(errMsg string) {
	chr.errorMessage = errMsg
}

func newCheckResult(cfg config.ConfigUrl) *CheckResult {
	var chr CheckResult
	chr.ConfigUrl = cfg
	chr.checksFailed = make([]string, 0)
	chr.checksOk = make([]string, 0)
	return &chr
}
