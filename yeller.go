package yeller

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type ErrorNotification struct {
	Type          string                 `json:"type"`
	Message       string                 `json:"message"`
	StackTrace    []StackFrame           `json:"stacktrace"`
	Url           string                 `json:"url"`
	Host          string                 `json:"host"`
	Environment   string                 `json:"application-environment"`
	CustomData    map[string]interface{} `json:"custom-data"`
	Location      string                 `json:"location"`
	ClientVersion string                 `json:"client-version"`
}

type StackFrame struct {
	Filename     string
	LineNumber   string
	FunctionName string
}

const (
	MAX_STACK_DEPTH = 256
)

var client *Client

func Start(apiKey string) {
	client = NewClient(apiKey, "production")
}

func StartEnv(apiKey string, env string) {
	client = NewClient(apiKey, env)
}

func Notify(appErr error) {
	NotifyInfo(appErr, make(map[string]interface{}))
}

func NotifyInfo(appErr error, info map[string]interface{}) {
	notification := newErrorNotification(appErr, info)
	err := client.Notify(notification)
	if err != nil {
		log.Println(err)
	}
}

func NotifyPanic(panicErr interface{}) {
	switch v := panicErr.(type) {
	case error:
		Notify(v)
	case string:
		Notify(errors.New(v))
	default:
		Notify(errors.New(fmt.Sprint(panicErr)))
	}
}

func NotifyPanicInfo(panicErr interface{}, info map[string]interface{}) {
	switch v := panicErr.(type) {
	case error:
		NotifyInfo(v, info)
	case string:
		NotifyInfo(errors.New(v), info)
	default:
		NotifyInfo(errors.New(fmt.Sprint(panicErr)), info)
	}
}

func (f StackFrame) MarshalJSON() ([]byte, error) {
	fields := []string{f.Filename, f.LineNumber, f.FunctionName}
	return json.Marshal(fields)
}

func newErrorNotification(err error, info map[string]interface{}) *ErrorNotification {
	return &ErrorNotification{
		Type:          "error",
		Message:       err.Error(),
		StackTrace:    applicationStackTrace(),
		Host:          applicationHostname(),
		Environment:   client.Environment,
		CustomData:    info,
		ClientVersion: client.Version,
	}
}

func applicationHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

func applicationStackTrace() (stackTrace []StackFrame) {
	for i := 1; i <= MAX_STACK_DEPTH+1; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}

		// Ignore all stack frames coming from this package
		if strings.Contains(file, "github.com/yeller/yeller-golang") {
			continue
		}

		frame := StackFrame{
			Filename:     file,
			LineNumber:   strconv.Itoa(line),
			FunctionName: functionName(pc),
		}
		stackTrace = append(stackTrace, frame)
	}
	return stackTrace
}

func functionName(pc uintptr) string {
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "???"
	}
	return fn.Name()
}
