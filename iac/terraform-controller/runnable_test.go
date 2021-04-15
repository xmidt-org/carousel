package terraform_controller

import "errors"

type simplerunnable struct {
	Name string
	Data []byte
}

func (s simplerunnable) Output() ([]byte, error) {
	if s.Data == nil {
		return []byte{}, errors.New("no data")
	}
	return s.Data, nil
}

func (s simplerunnable) String() string {
	return s.Name
}
