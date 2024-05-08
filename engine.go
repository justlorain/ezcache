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
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
)

var (
	ErrEmptyParam        = errors.New("error empty param")
	ErrPickNote          = errors.New("error pick node")
	ErrFormRequest       = errors.New("error form request")
	ErrReadBody          = errors.New("error read body")
	ErrReachedRetryTimes = errors.New("error reached retry times")
)

const _internalFlag = "?type=internal"

type Engine struct {
	mu      sync.RWMutex
	options *Options
	nodes   *Hash
	cache   *Cache
	addrs   []string
}

func NewEngine(opts ...Option) *Engine {
	options := newOptions(opts...)
	return &Engine{
		options: options,
		nodes:   NewHash(0, nil),
		cache:   NewCache(),
		addrs:   make([]string, 0),
	}
}

func (e *Engine) RegisterNodes(addrs ...string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.nodes.Add(addrs...)
	e.addrs = append(e.addrs, addrs...)
}

func (e *Engine) Run() error {
	mux := http.NewServeMux()
	// to keep it simple, we used url path value here rather than http body
	mux.HandleFunc("POST "+e.options.BasePath+"/{key}/{value}", e.Set)
	mux.HandleFunc("GET "+e.options.BasePath+"/{key}", e.Get)
	mux.HandleFunc("DELETE "+e.options.BasePath+"/{key}", e.Delete)
	srv := http.Server{
		Addr:    e.options.Addr,
		Handler: mux,
	}
	return srv.ListenAndServe()
}

func (e *Engine) Set(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value := r.PathValue("value")
	if key == "" || value == "" {
		http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}

	// set locally
	e.cache.Set(key, ByteView{
		B: []byte(value),
	})

	// drop if from internal node
	if r.URL.Query().Get("type") == "internal" {
		return
	}

	// set distantly
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, addr := range e.addrs {
		// skip node itself
		if addr == e.options.Addr {
			continue
		}

		url := fmt.Sprintf("%v%v/%v/%v%v", addr, e.options.BasePath, key, value, _internalFlag)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			http.Error(w, ErrFormRequest.Error(), http.StatusInternalServerError)
			return
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, ErrFormRequest.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func (e *Engine) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}

	// get locally
	if value, ok := e.cache.Get(key); ok {
		_, _ = w.Write(value.ByteSlice())
		return
	}

	// get distantly
	var times uint32 = 0
	for {
		atomic.AddUint32(&times, 1)
		// reached max retry times
		if int(atomic.LoadUint32(&times)) == e.options.RetryTimes {
			http.Error(w, ErrReachedRetryTimes.Error(), http.StatusInternalServerError)
			return
		}

		e.mu.RLock()
		pickedNodeAddr := e.nodes.Get(key)
		e.mu.RUnlock()

		if pickedNodeAddr == "" {
			http.Error(w, ErrPickNote.Error(), http.StatusInternalServerError)
			return
		}
		// retry
		if pickedNodeAddr == e.options.Addr {
			continue
		}

		url := fmt.Sprintf("%v%v/%v", pickedNodeAddr, e.options.BasePath, key)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			http.Error(w, ErrFormRequest.Error(), http.StatusInternalServerError)
			return
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			_ = resp.Body.Close()
			http.Error(w, ErrFormRequest.Error(), http.StatusInternalServerError)
			return
		}

		// retry
		if resp.StatusCode != http.StatusOK {
			_ = resp.Body.Close()
			continue
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			http.Error(w, ErrReadBody.Error(), http.StatusInternalServerError)
			return
		}
		// populate cache
		e.cache.Set(key, ByteView{
			B: CopyBytes(body),
		})

		_, _ = w.Write(body)
		return
	}
}

func (e *Engine) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}

	// delete locally
	e.cache.Delete(key)

	// drop if from internal node
	if r.URL.Query().Get("type") == "internal" {
		return
	}

	// delete distantly
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, addr := range e.addrs {
		if addr == e.options.Addr {
			continue
		}
		url := fmt.Sprintf("%v%v/%v%v", addr, e.options.BasePath, key, _internalFlag)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			http.Error(w, ErrFormRequest.Error(), http.StatusInternalServerError)
			return
		}
		_, err = http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, ErrFormRequest.Error(), http.StatusInternalServerError)
			return
		}
	}

}
