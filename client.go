package yeller

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"
)

type Client struct {
	ApiKey          string
	Environment     string
	Version         string
	lastHostnameIdx int
	hostnames       []string
	httpClient      *http.Client
	errorHandler    YellerErrorHandler
}

type YellerErrorHandler interface {
	HandleIOError(error) error
	HandleAuthError(error) error
}

const CLIENT_VERSION = "yeller-golang: 0.0.1"

func NewClient(apiKey string, env string, errorHandler YellerErrorHandler) (client *Client) {
	yellerHostnames := []string{
		"collector1.yellerapp.com",
		"collector2.yellerapp.com",
		"collector3.yellerapp.com",
		"collector4.yellerapp.com",
		"collector5.yellerapp.com",
	}
	return NewClientHostnames(apiKey, env, errorHandler, yellerHostnames)
}

func NewClientHostnames(apiKey string, env string, errorHandler YellerErrorHandler, hostnames []string) (client *Client) {
	// Set a timeout of 1 second before moving on to a different host
	transport := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			timeout := time.Duration(1 * time.Second)
			return net.DialTimeout(network, addr, timeout)
		},
	}
	httpClient := http.Client{Transport: &transport}

	return &Client{
		ApiKey:          apiKey,
		Environment:     env,
		Version:         CLIENT_VERSION,
		lastHostnameIdx: randomHostnameIdx(hostnames),
		hostnames:       hostnames,
		httpClient:      &httpClient,
		errorHandler:    errorHandler,
	}
}

func (c *Client) Notify(note *ErrorNotification) error {
	json, err := json.Marshal(note)
	if err != nil {
		return err
	}

	for _ = range c.hostnames {
		err = c.tryNotifying(json)
		if err == nil {
			break
		} else {
			c.cycleHostname()
		}
	}

	if err != nil {
		c.errorHandler.HandleIOError(err)
		return err
	}
	return nil
}

type LogErrorHandler struct {
	logger *log.Logger
}

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
	// XXX
	return nil
}

func (c *Client) tryNotifying(json []byte) error {
	url := "http://" + c.hostname() + "/" + c.ApiKey
	response, err := c.httpClient.Post(url, "application/json", bytes.NewReader(json))
    if err != nil {
        panic(err)
    }
	if response.StatusCode == 401 {
		authError := errors.New("Could not authenticate yeller client. Check your API key and that your subscription is active")
		c.errorHandler.HandleAuthError(authError)
		return nil
	}
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return errors.New("Received a non 200 HTTP Code: " + response.Status)
	}
	return err
}

func (c *Client) hostname() string {
	return c.hostnames[c.lastHostnameIdx]
}

func (c *Client) cycleHostname() {
	c.lastHostnameIdx = (c.lastHostnameIdx + 1) % len(c.hostnames)
}

func randomHostnameIdx(hostnames []string) int {
	// Use a locally-scoped random source to avoid overwriting
	// global state.
	randSrc := rand.NewSource(time.Now().UTC().UnixNano())
	random := rand.New(randSrc)
	return random.Intn(len(hostnames))
}
