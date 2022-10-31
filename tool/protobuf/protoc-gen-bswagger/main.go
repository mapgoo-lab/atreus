package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/mapgoo-lab/atreus/tool/protobuf/pkg/gen"
	"github.com/mapgoo-lab/atreus/tool/protobuf/pkg/generator"
)

func main() {
	versionFlag := flag.Bool("version", false, "print version and exit")
	isSimpleJson := flag.Bool("simplejson", false, "Simple json export")
	flag.Parse()
	if *versionFlag {
		fmt.Println(generator.Version)
		os.Exit(0)
	}

	g := NewSwaggerGenerator(*isSimpleJson)
	gen.Main(g)
}
