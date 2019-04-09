package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	binPath := flag.String("path", "", "bin path")
	outDir := flag.String("out", "", "output directory")
	qtPlugin := flag.Bool("qt", false, "enable qt plugin")

	flag.Parse()

	dep := NewDepends(*binPath)
	err := dep.Install(*outDir, *qtPlugin)
	if nil != err {
		fmt.Println(err)
		os.Exit(-1)
	}
}
