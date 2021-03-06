package model

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestErrors(t *testing.T) {
	assert := assert.New(t)
	errors := []error{errors.New("error1"), errors.New("testing error list"), errors.New("test"), Errors([]error{errors.New("inner list test"), errors.New("")})}
	errorsString := "multiple errors: [error1, testing error list, test, multiple errors: [inner list test, ]]"
	assert.Equal(errorsString, Errors(errors).Error())
	assert.Equal(errors, Errors(errors).Errors())
}
