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
	"net/http"
	"net/url"
	"testing"

	"github.com/codefresh-io/go/venona/pkg/logger"
	"github.com/codefresh-io/go/venona/pkg/mocks"
	"github.com/stretchr/testify/assert"
)

func buildFakeMock() *mocks.Logger {
	l := &mocks.Logger{}
	return l
}

func mustURL(u string) *url.URL {
	r, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return r
}

func TestNew(t *testing.T) {
	type args struct {
		opt Options
	}
	tests := []struct {
		name string
		args args
		want Codefresh
	}{
		{
			name: "Build client with default host",
			args: args{},
			want: &cf{
				host:       defaultHost,
				httpClient: buildClient(false),
			},
		},
		{
			name: "Build client with given host",
			args: args{
				Options{
					Host: "http://host.com",
				},
			},
			want: &cf{
				host:       "http://host.com",
				httpClient: buildClient(false),
			},
		},
		{
			name: "Build client with insecure mode",
			args: args{
				Options{
					Host:     "http://host.com",
					Insecure: true,
				},
			},
			want: &cf{
				host:       "http://host.com",
				httpClient: buildClient(true),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, New(tt.args.opt))
		})
	}
}

func Test_cf_prepareURL(t *testing.T) {
	type fields struct {
		host       string
		token      string
		agentID    string
		logger     logger.Logger
		httpClient interface {
			Do(*http.Request) (*http.Response, error)
		}
	}
	type args struct {
		path string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *url.URL
		wantErr bool
	}{
		{
			name: "Reject when parsing the URL faile",
			args: args{},
			fields: fields{
				host: "123://sdd",
			},
			wantErr: true,
		},
		{
			name: "Append path to the host",
			args: args{
				path: "/123/123",
			},
			fields: fields{
				host: "http://url",
			},
			wantErr: false,
			want:    mustURL("http://url/123/123"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cf{
				host:       tt.fields.host,
				token:      tt.fields.token,
				agentID:    tt.fields.agentID,
				logger:     tt.fields.logger,
				httpClient: tt.fields.httpClient,
			}
			_, err := c.prepareURL(tt.args.path)
			if tt.wantErr {
				assert.Error(t, err)
			}
		})
	}
}
