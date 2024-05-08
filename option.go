// Copyright 2024 justlorain
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

package ezcache

import "strings"

type Option func(o *Options)

type Options struct {
	BasePath   string
	Addr       string
	RetryTimes int
}

var defaultOptions = Options{
	BasePath:   "/_ezcache",
	Addr:       ":8080",
	RetryTimes: 3,
}

func newOptions(opts ...Option) *Options {
	options := &Options{
		BasePath:   defaultOptions.BasePath,
		Addr:       defaultOptions.Addr,
		RetryTimes: defaultOptions.RetryTimes,
	}
	options.apply(opts...)
	return options
}

func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}
}

// WithBasePath used to define base path of server
func WithBasePath(path string) Option {
	return func(o *Options) {
		o.BasePath = path
	}
}

// WithAddr used to define host addr server listens
func WithAddr(addr string) Option {
	addr = StandardizeAddr(addr)
	return func(o *Options) {
		o.Addr = addr
	}
}

// WithRetryTimes used to define retry times for each node
func WithRetryTimes(times int) Option {
	return func(o *Options) {
		o.RetryTimes = times
	}
}

func StandardizeAddr(addr string) string {
	segments := strings.Split(addr, "://")
	length := len(segments)
	if length == 1 {
		return segments[0]
	}
	if length == 2 {
		return segments[1]
	}
	return ""
}
