package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func getQtPluginFileList() (pluginRoot string, list []string) {
	cmd := exec.Command("qmake", "-query", "QT_INSTALL_PLUGINS")
	data, err := cmd.Output()
	if nil != err {
		fmt.Println(err)
		return
	}
	pluginRoot = strings.TrimSpace(string(data))
	pluginDirs := []string{
		"iconengines",
		"imageformats",
		"platforminputcontexts",
		"platforms",
		"platformthemes",
		"styles",
		"xcbglintegrations",
	}

	for _, dir := range pluginDirs {
		filepath.Walk(filepath.Join(pluginRoot, dir),
			func(path string, info os.FileInfo, err error) error {
				if filepath.Ext(path) != ".so" {
					return nil
				}
				list = append(list, path)
				return nil
			})
	}
	return
}
