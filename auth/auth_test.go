package auth

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
)

type MockClient struct{}

var (
	DoFunc func(req *http.Request) (*http.Response, error)
)

func (m *MockClient) Do(req *http.Request) (*http.Response, error) {
	return DoFunc(req)
}

func TestMain(m *testing.M) {
	client = &MockClient{}
	os.Exit(m.Run())
}

func TestAuthStartError(t *testing.T) {
	DoFunc = func(*http.Request) (*http.Response, error) {
		return nil, errors.New(
			"Wanted error from mock web server",
		)
	}

	data, err := StartDeviceAuth("", "0")
	if err == nil {
		t.Error(err)
	}

	assert.Equal(t, data, DeviceCode{})
	assert.NotNil(t, err)
}

func TestAuthStartData(t *testing.T) {
	d := DeviceCode{
		DeviceCode:              "0",
		UserCode:                "test",
		VerificationURI:         "test",
		ExpiresIn:               1,
		Interval:                1,
		VerificationURIComplete: "test11",
	}
	DoFunc = func(req *http.Request) (*http.Response, error) {
		header := req.Header.Values("content-type")
		if len(header) != 1 {
			return nil, errors.New("content-type len is wrong")
		}

		if header[0] != "application/x-www-form-urlencoded" {
			return nil, errors.New("content-type is wrong")
		}

		if req.Method != "POST" {
			return nil, errors.New("Method is wrong")
		}

		if !strings.Contains(req.URL.Path, "/oauth/device/code") {
			return nil, errors.New("url is wrong")
		}

		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		buf := new(strings.Builder)
		_, err = io.Copy(buf, body)
		if err != nil {
			return nil, err
		}

		if !strings.Contains(buf.String(), "client_id=") ||
			!strings.Contains(buf.String(), "&audience=https://api.arduino.cc") {
			return nil, errors.New("Payload is wrong")
		}

		data, err := json.Marshal(d)
		if err != nil {
			return nil, err
		}

		return &http.Response{
			Body: ioutil.NopCloser(bytes.NewBufferString(string(data))),
		}, nil
	}

	data, err := StartDeviceAuth("", "0")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, data, d)
}

func TestAuthCheck(t *testing.T) {
	type AuthAccess struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
		TokenType   string `json:"token_type"`
	}

	DoFunc = func(req *http.Request) (*http.Response, error) {
		header := req.Header.Values("content-type")
		if len(header) != 1 {
			return nil, errors.New("content-type len is wrong")
		}

		if header[0] != "application/x-www-form-urlencoded" {
			return nil, errors.New("content-type is wrong")
		}

		if req.Method != "POST" {
			return nil, errors.New("Method is wrong")
		}

		if !strings.Contains(req.URL.Path, "/oauth/token") {
			return nil, errors.New("url is wrong")
		}

		body, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		buf := new(strings.Builder)
		_, err = io.Copy(buf, body)
		if err != nil {
			return nil, err
		}

		if !strings.Contains(buf.String(), "client_id=") ||
			!strings.Contains(buf.String(), "grant_type=urn%3Aietf%3Aparams%3Aoauth%3Agrant-type%3Adevice_code&device_code=") {
			return nil, errors.New("Payload is wrong")
		}

		data, err := json.Marshal(AuthAccess{
			AccessToken: "asdf",
			ExpiresIn:   999,
			TokenType:   "testType",
		})
		if err != nil {
			return nil, err
		}

		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(string(data))),
		}, nil
	}

	token, err := CheckDeviceAuth("", "0", "")
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "asdf", token)
}
