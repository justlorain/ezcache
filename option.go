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

import (
	"hash/crc32"
)

type Option func(o *Options)

type Options struct {
	BasePath          string
	ReplicationFactor int
	HashFunc          HashFunc
}

var defaultOptions = Options{
	BasePath:          "/_ezcache",
	ReplicationFactor: 10,
	HashFunc:          crc32.ChecksumIEEE,
}

func newOptions(opts ...Option) *Options {
	options := &Options{
		BasePath:          defaultOptions.BasePath,
		ReplicationFactor: defaultOptions.ReplicationFactor,
		HashFunc:          defaultOptions.HashFunc,
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

// WithReplicationFactor used to define consistent hash replication factor
func WithReplicationFactor(factor int) Option {
	return func(o *Options) {
		o.ReplicationFactor = factor
	}
}

// WithHashFunc used to define hash func used by consistent hash
func WithHashFunc(fn HashFunc) Option {
	return func(o *Options) {
		o.HashFunc = fn
	}
}
