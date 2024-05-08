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

import "sync"

type Cache struct {
	mu sync.RWMutex
	kv map[string]any
}

func NewCache() *Cache {
	return &Cache{
		kv: make(map[string]any),
	}
}

func (c *Cache) Set(key string, value ByteView) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.kv[key] = value
}

func (c *Cache) Get(key string) (ByteView, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if value, ok := c.kv[key]; ok {
		return value.(ByteView), true
	}
	return ByteView{}, false
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.kv, key)
}
