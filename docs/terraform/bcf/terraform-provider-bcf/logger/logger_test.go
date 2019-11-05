/*
 * Copyright 2019 Big Switch Networks, Inc.
 */

package logger

import (
	"bytes"
	"math/rand"
	"os"
	"strings"
	"testing"
)

var (
	letter = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func Test_AllLevels(t *testing.T) {
	text := "This is the test string"
	textf := "This is the test %s"
	textt := "string"

	buf := new(bytes.Buffer)
	setOutput(buf)

	Debug(text)
	line := buf.String()
	if strings.Contains(line, LogLevelMap["DEBUG"].strValue) == false {
		t.Error("Unexpected failure")
	}
	Debugf(textf, textt)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["DEBUG"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Infof(textf, textt)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["INFO"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Info(text)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["INFO"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Warn(text)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["WARN"].strValue) == false {
		t.Error("Unexpected failure")
	}
	Warnf(textf, textt)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["WARN"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Errorf(textf, textt)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["ERROR"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Error(text)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["ERROR"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Fatalf(textf, textt)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["FATAL"].strValue) == false {
		t.Error("Unexpected failure")
	}

	Fatal(text)
	line = buf.String()
	if strings.Contains(line, LogLevelMap["FATAL"].strValue) == false {
		t.Error("Unexpected failure")
	}
}

func Test_OutLogFile(t *testing.T) {
	buf := make([]rune, 8)
	for i := range buf {
		buf[i] = letter[rand.Intn(len(letter))]
	}

	outFileName := "/tmp/testing" + string(buf) + ".tmp"
	err := SetLogFile(outFileName)
	if err != nil {
		t.Error("Unexpected failure", err)
	}

	text := "This is the test string"
	Info(text)

	file, err := os.Open(outFileName)
	if err != nil {
		t.Error("Unexpected failure", err)
	}
	defer file.Close()

	data := make([]byte, 100)
	_, err = file.Read(data)
	if err != nil {
		t.Error("Unexpected failure", err)
	}

	if strings.Contains(string(data), text) == false {
		t.Error("Unexpected failure")
	}
	os.Remove(outFileName)
}

func Test_LogLevel(t *testing.T) {
	buf := make([]rune, 8)
	for i := range buf {
		buf[i] = letter[rand.Intn(len(letter))]
	}

	outFileName := "/tmp/testing" + string(buf) + ".tmp"
	err := SetLogFile(outFileName)
	if err != nil {
		t.Error("Unexpected failure", err)
	}

	text := "This is the test string"

	SetLogLevel("INFO")
	Debug(text)
	file, err := os.Open(outFileName)
	if err != nil {
		t.Error("Unexpected failure", err)
	}
	defer file.Close()

	data := make([]byte, 100)

	_, err = file.Read(data)
	if err == nil {
		// Should have been empty!
		t.Error("Unexpected failure", err)
	}

	SetLogLevel("DEBUG")
	Debug(text)
	_, err = file.Read(data)
	if err != nil {
		t.Error("Unexpected failure", err)
	}

	if strings.Contains(string(data), text) == false {
		t.Error("Unexpected failure")
	}

	SetLogLevel("INVALID_LEVEL")

	os.Remove(outFileName)
}
