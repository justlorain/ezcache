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
	"log/slog"
	"net/http"
	"strings"
	"sync"
)

var (
	ErrEmptyParam = errors.New("error empty param")
	ErrPickNote   = errors.New("error pick node")
)

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
	slog.Info("EZCache: engine starts running", "addr", e.options.Addr)
	slog.Info("EZCache: node list", "addrs", e.addrs)
	return srv.ListenAndServe()
}

func (e *Engine) Set(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	value := r.PathValue("value")
	if key == "" || value == "" {
		http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}

	e.mu.RLock()
	pickedNodeAddr := e.nodes.Get(key)
	e.mu.RUnlock()
	slog.Info("EZCache: pick node", "addr", pickedNodeAddr)

	if pickedNodeAddr == "" {
		http.Error(w, ErrPickNote.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(pickedNodeAddr, e.options.Addr) {
		e.cache.Set(key, ByteView{
			B: []byte(value),
		})
		slog.Info("EZCache: set", "key", key, "value", value)
		return
	}

	url := fmt.Sprintf("%v%v/%v/%v", pickedNodeAddr, e.options.BasePath, key, value)
	if err := DoHTTPRequest(http.MethodPost, url, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("EZCache: node redirect set request", "from", e.options.Addr, "to", pickedNodeAddr)
}

func (e *Engine) Get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}

	e.mu.RLock()
	pickedNodeAddr := e.nodes.Get(key)
	e.mu.RUnlock()
	slog.Info("EZCache: pick node", "addr", pickedNodeAddr)

	if pickedNodeAddr == "" {
		http.Error(w, ErrPickNote.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(pickedNodeAddr, e.options.Addr) {
		if v, ok := e.cache.Get(key); ok {
			_, _ = w.Write(v.ByteSlice())
			slog.Info("EZCache: get", "key", key, "value", v.String())
			return
		}
		slog.Info("EZCache: key does not exist", "key", key)
		return
	}

	url := fmt.Sprintf("%v%v/%v", pickedNodeAddr, e.options.BasePath, key)
	slog.Info("EZCache: node redirect get request", "from", e.options.Addr, "to", pickedNodeAddr)
	http.Redirect(w, r, url, http.StatusFound)
}

func (e *Engine) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		http.Error(w, ErrEmptyParam.Error(), http.StatusBadRequest)
		return
	}

	e.mu.RLock()
	pickedNodeAddr := e.nodes.Get(key)
	e.mu.RUnlock()
	slog.Info("EZCache: pick node", "addr", pickedNodeAddr)

	if pickedNodeAddr == "" {
		http.Error(w, ErrPickNote.Error(), http.StatusInternalServerError)
		return
	}

	if strings.Contains(pickedNodeAddr, e.options.Addr) {
		e.cache.Delete(key)
		slog.Info("EZCache: delete", "key", key)
		return
	}

	url := fmt.Sprintf("%v%v/%v", pickedNodeAddr, e.options.BasePath, key)
	if err := DoHTTPRequest(http.MethodDelete, url, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	slog.Info("EZCache: node redirect delete request", "from", e.options.Addr, "to", pickedNodeAddr)
}

func DoHTTPRequest(method, url string, body io.Reader) error {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return err
	}
	_, err = http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	return nil
}
