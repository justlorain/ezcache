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
	"fmt"
	"sort"
)

type Hash struct {
	ring              []uint32
	nodes             map[uint32]string
	replicationFactor int
	hashFunc          HashFunc
}

type HashFunc func(data []byte) uint32

// NewHash consistent hash
func NewHash(factor int, fn HashFunc) *Hash {
	return &Hash{
		ring:              make([]uint32, 0),
		nodes:             make(map[uint32]string),
		replicationFactor: factor,
		hashFunc:          fn,
	}
}

func (h *Hash) Add(nodes ...string) {
	for _, node := range nodes {
		for i := 0; i < h.replicationFactor; i++ {
			hash := h.hashFunc([]byte(fmt.Sprintf("%s%d", node, i)))
			h.nodes[hash] = node
			h.ring = append(h.ring, hash)
		}
	}
	sort.Slice(h.ring, func(i, j int) bool {
		return h.ring[i] < h.ring[j]
	})
}

func (h *Hash) Get(key string) string {
	if len(h.nodes) == 0 {
		return ""
	}
	// same key will get same hash so this ensures that a picked node won't pick another node
	hash := h.hashFunc([]byte(key))
	idx := sort.Search(len(h.ring), func(i int) bool {
		return h.ring[i] >= hash
	})
	// handle case idx == len(h.ring), which will choose h.ring[0]
	if idx >= len(h.ring) {
		idx = 0
	}
	return h.nodes[h.ring[idx]]
}

func (h *Hash) Remove(key string) {
	if key == "" {
		return
	}
	for i := 0; i < h.replicationFactor; i++ {
		hash := h.hashFunc([]byte(fmt.Sprintf("%s%d", key, i)))
		delete(h.nodes, hash)
		idx := SearchUint32s(h.ring, hash)
		if idx == -1 {
			return
		}
		copy(h.ring[idx:], h.ring[idx+1:])
		h.ring = h.ring[:len(h.ring)-1]
	}
}

func SearchUint32s(s []uint32, target uint32) int {
	left, right := 0, len(s)-1
	for left <= right {
		mid := left + ((right - left) >> 1)
		if s[mid] == target {
			return mid
		} else if s[mid] < target {
			left = mid + 1
		} else {
			right = mid - 1
		}
	}
	return -1
}
