/*
 * Copyright 2019 Big Switch Networks, Inc.
 */

package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var (
	logFlags    = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile
	callDepth   = 3
	logger      = log.New(os.Stderr, "", logFlags)
	loggerLevel = DEBUG
)

type LogLevelInt uint8

const (
	DEBUG LogLevelInt = iota + 1
	INFO
	WARN
	ERROR
	FATAL
)

type LogLevel struct {
	strValue string
	intValue LogLevelInt
}

var LogLevelMap = map[string]LogLevel{
	"DEBUG": {"DEBUG ", DEBUG},
	"INFO":  {"INFO  ", INFO},
	"WARN":  {"WARN  ", WARN},
	"ERROR": {"ERROR ", ERROR},
	"FATAL": {"FATAL ", FATAL},
}

func setOutput(w io.Writer) {
	logger.SetOutput(w)
}

func SetLogLevel(logLevelStr string) {
	if level, valid := LogLevelMap[strings.ToUpper(logLevelStr)]; valid {
		loggerLevel = level.intValue
	}
}

func SetLogFile(fName string) error {
	fp, err := os.OpenFile(fName, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		logger.Printf("Log file creation failed: %+v, Defaulting to STDOUT\n", err)
		return err
	}
	setOutput(fp)
	return nil
}

func logThis(logLevel LogLevel, logMsg string) {
	if logLevel.intValue < loggerLevel {
		return
	}
	logger.SetPrefix(logLevel.strValue)
	logger.Output(callDepth, logMsg)
}

func Debugf(format string, v ...interface{}) {
	logThis(LogLevelMap["DEBUG"], fmt.Sprintf(format, v...))
}

func Debug(v ...interface{}) {
	logThis(LogLevelMap["DEBUG"], fmt.Sprintln(v...))
}

func Infof(format string, v ...interface{}) {
	logThis(LogLevelMap["INFO"], fmt.Sprintf(format, v...))
}

func Info(v ...interface{}) {
	logThis(LogLevelMap["INFO"], fmt.Sprintln(v...))
}

func Warnf(format string, v ...interface{}) {
	logThis(LogLevelMap["WARN"], fmt.Sprintf(format, v...))
}

func Warn(v ...interface{}) {
	logThis(LogLevelMap["WARN"], fmt.Sprintln(v...))
}

func Errorf(format string, v ...interface{}) {
	logThis(LogLevelMap["ERROR"], fmt.Sprintf(format, v...))
}

func Error(v ...interface{}) {
	logThis(LogLevelMap["ERROR"], fmt.Sprintln(v...))
}

func Fatalf(format string, v ...interface{}) {
	logThis(LogLevelMap["FATAL"], fmt.Sprintf(format, v...))
}

func Fatal(v ...interface{}) {
	logThis(LogLevelMap["FATAL"], fmt.Sprintln(v...))
}
