// Copyright 2020 The Codefresh Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package codefresh

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/task"
)

const (
	defaultHost = "https://g.codefresh.io"
)

type (
	// Codefresh API client
	Codefresh interface {
		Tasks() ([]task.Task, error)
		ReportStatus(status AgentStatus) error
		Host() string
	}

	// RequestDoer interface
	RequestDoer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// Options for codefresh
	Options struct {
		Host       string
		Token      string
		AgentID    string
		Logger     logger.Logger
		HTTPClient RequestDoer
	}

	cf struct {
		host       string
		token      string
		agentID    string
		logger     logger.Logger
		httpClient interface {
			Do(*http.Request) (*http.Response, error)
		}
	}
)

// New build Codefresh client from options
func New(opt Options) Codefresh {
	host := opt.Host
	if host == "" {
		host = defaultHost
	}

	httpClient := opt.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{}
	}

	return &cf{
		agentID:    opt.AgentID,
		httpClient: httpClient,
		host:       host,
		logger:     opt.Logger,
		token:      opt.Token,
	}
}

// Tasks get from Codefresh all latest tasks
func (c cf) Tasks() ([]task.Task, error) {
	c.logger.Debug("Requesting tasks")
	res, err := c.doRequest("GET", path.Join("api", "agent", c.agentID, "tasks"), nil)
	if err != nil {
		return nil, err
	}
	tasks, err := task.UnmarshalTasks(res)
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

// Host returns the host
func (c cf) Host() string {
	return c.host
}

// ReportStatus updates the agent entity with given status
func (c cf) ReportStatus(status AgentStatus) error {
	c.logger.Debug("Reporting status")
	s, err := status.Marshal()
	if err != nil {
		return err
	}
	_, err = c.doRequest("PUT", path.Join("api", "agent", c.agentID, "status"), bytes.NewBuffer(s))
	if err != nil {
		return err
	}
	return nil
}

func (c cf) buildErrorFromResponse(status int, body []byte) error {
	return Error{
		APIStatusCode: status,
		Message:       string(body),
	}
}

func (c cf) prepareURL(path string) (*url.URL, error) {
	u, err := url.Parse(c.host)
	if err != nil {
		return nil, err
	}
	u.Path = path
	return u, nil
}

func (c cf) prepareRequest(method string, api string, data io.Reader) (*http.Request, error) {
	u, err := c.prepareURL(api)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(method, u.String(), data)
	if err != nil {
		return nil, err
	}
	if c.token != "" {
		req.Header.Add("Authorization", c.token)
	}
	req.Header.Add("Codefresh-Agent-Version", "")
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}

func (c cf) doRequest(method string, api string, body io.Reader) ([]byte, error) {
	req, err := c.prepareRequest(method, api, body)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, c.buildErrorFromResponse(resp.StatusCode, data)
	}
	return data, nil
}
