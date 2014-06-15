package yeller

import (
	"log"
	"os"
)

type LogErrorHandler struct {
	logger *log.Logger
}

type SilentErrorHandler struct{}

func (l *LogErrorHandler) HandleIOError(e error) error {
	l.logger.Println(e)
	return nil
}

func (l *LogErrorHandler) HandleAuthError(e error) error {
	l.logger.Println(e)
	return nil
}

func NewLogErrorHandler(l *log.Logger) YellerErrorHandler {
	return &LogErrorHandler{
		logger: l,
	}
}

func NewStdErrErrorHandler() YellerErrorHandler {
	return NewLogErrorHandler(log.New(os.Stderr, "yeller", log.Flags()))
}

func NewSilentErrorHandler() YellerErrorHandler {
	return &SilentErrorHandler{}
}

func (l *SilentErrorHandler) HandleIOError(e error) error {
	return nil
}

func (l *SilentErrorHandler) HandleAuthError(e error) error {
	return nil
}
