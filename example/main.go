package main

import (
	"flag"
	"strings"

	"github.com/justlorain/ezcache"
)

func main() {
	addrs := ParseFlags()
	e := ezcache.NewEngine(ezcache.WithAddr(addrs[0]))
	e.RegisterNodes(addrs...)
	if err := e.Run(); err != nil {
		return
	}
}

// ParseFlags e.g.
// go run . --addrs=http://localhost:7246,http://localhost:7247,http://localhost:7248
// go run . --addrs=http://localhost:7247,http://localhost:7248,http://localhost:7246
// go run . --addrs=http://localhost:7248,http://localhost:7246,http://localhost:7247
// NOTE: first addr is local node
func ParseFlags() []string {
	var addrsFlag string
	flag.StringVar(&addrsFlag, "addrs", "", "nodes addresses")
	flag.Parse()
	return strings.Split(addrsFlag, ",")
}
