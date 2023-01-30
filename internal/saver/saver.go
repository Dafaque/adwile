package saver

type Saver interface {
	Save(url string, passed bool, failed []string, errMsg string) error
	GetLastStatus(url string) (*bool, error)
}
