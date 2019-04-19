package main

import (
	"flag"
	"fmt"
	"os"
)

type listFlags []string

func (l *listFlags) String() string {
	return "plugin"
}

func (l *listFlags) Set(value string) error {
	*l = append(*l, value)
	return nil
}

var pluginList listFlags

func main() {
	binPath := flag.String("path", "", "bin path")
	outDir := flag.String("out", "", "output directory")
	qtPlugin := flag.Bool("qt", false, "enable qt plugin")
	flag.Var(&pluginList, "plugin", "plugin dir or file, like: \n/usr/lib/x86_64-linux-gnu:/usr/lib/x86_64-linux-gnu/nss \n/usr/lib/x86_64-linux-gnu:/usr/lib/x86_64-linux-gnu/nss/libfreebl3.so")

	flag.Parse()

	dep := NewDepends(*binPath)
	err := dep.Install(*outDir, *qtPlugin, pluginList)
	if nil != err {
		fmt.Println(err)
		os.Exit(-1)
	}
}
