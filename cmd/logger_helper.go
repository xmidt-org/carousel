package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"github.com/mitchellh/cli"
)

type UILogger struct {
	cli.Ui
}

var errMissingValue = errors.New("(MISSING)")

func (c *UILogger) Log(keyvals ...interface{}) error {
	fields := LogLine{}
	for i := 0; i < len(keyvals); i += 2 {
		if i+1 < len(keyvals) {
			fields[fmt.Sprint(keyvals[i])] = keyvals[i+1]
		} else {
			fields[fmt.Sprint(keyvals[i])] = errMissingValue
		}
	}
	switch fields[fmt.Sprint(level.Key())] {
	case level.DebugValue():
		// TODO:// verbose logging
	case level.InfoValue():
		c.Output(fields.String())
	case level.WarnValue():
		c.Warn(fields.String())
	case level.ErrorValue():
		c.Error(fields.String())
	default:
		c.Output(fields.String())
	}
	return nil
}

type LogLine map[string]interface{}

func (l LogLine) String() string {
	buf := &bytes.Buffer{}
	if message, ok := l["msg"]; ok {
		buf.WriteString(fmt.Sprintf("%s", message))
	}

	for key, value := range l {
		if key != fmt.Sprint(level.Key()) && key != "msg" {
			buf.WriteString(fmt.Sprintf("%s=%s ", key, value))
		}
	}
	return buf.String()
}
