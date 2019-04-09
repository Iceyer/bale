package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func copyFile(src, dest string) error {
	s, err := os.Open(src)
	if nil != err {
		return err
	}
	defer s.Close()

	fi, _ := s.Stat()
	d, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE, fi.Mode())
	if nil != err {
		return err
	}
	defer d.Close()
	_, err = io.Copy(d, s)
	return err
}

var blockList = []string{
	"ld-linux.so.2",
	"ld-linux-x86-64.so.2",
	"libanl.so.1",
	"libasound.so.2",
	"libBrokenLocale.so.1",
	"libcidn.so.1",
	"libcom_err.so.2",
	"libcrypt.so.1",
	"libc.so.6",
	"libdl.so.2",
	"libdrm.so.2",
	"libexpat.so.1",
	"libfontconfig.so.1",
	"libfreetype.so.6",
	"libgcc_s.so.1",
	"libgdk_pixbuf-2.0.so.0",
	"libgio-2.0.so.0",
	"libglapi.so.0",
	"libglib-2.0.so.0",
	"libGL.so.1",
	"libgobject-2.0.so.0",
	"libgpg-error.so.0",
	"libharfbuzz.so.0",
	"libICE.so.6",
	"libjack.so.0",
	"libkeyutils.so.1",
	"libm.so.6",
	"libmvec.so.1",
	"libnsl.so.1",
	"libnss_compat.so.2",
	"libnss_db.so.2",
	"libnss_dns.so.2",
	"libnss_files.so.2",
	"libnss_hesiod.so.2",
	"libnss_nisplus.so.2",
	"libnss_nis.so.2",
	"libp11-kit.so.0",
	"libpango-1.0.so.0",
	"libpangocairo-1.0.so.0",
	"libpangoft2-1.0.so.0",
	"libpthread.so.0",
	"libresolv.so.2",
	"librt.so.1",
	"libSM.so.6",
	"libstdc++.so.6",
	"libthai.so.0",
	"libthread_db.so.1",
	"libusb-1.0.so.0",
	"libutil.so.1",
	"libuuid.so.1",
	"libX11.so.6",
	"libxcb.so.1",
	"libz.so.1",
}

var bashTemplate = `#!/bin/sh
HERE="$(dirname "$(readlink -f "${0}")")"
export LD_LIBRARY_PATH="${HERE}"/libs
export QT_PLUGIN_PATH="${HERE}"/plugins 
export QT_QPA_PLATFORM_PLUGIN_PATH="${HERE}"/plugins/platforms
exec "${HERE}"/%v $@
`

// Depends of binPath
type Depends struct {
	binPath string
	libs    map[string]string
	black   map[string]string
}

// NewDepends for binary file
func NewDepends(binPath string) *Depends {
	dep := &Depends{
		binPath: binPath,
		libs:    make(map[string]string),
		black:   make(map[string]string),
	}
	for _, v := range blockList {
		dep.black[v] = v
	}

	dep.getSharedLibraryDependencies(binPath)
	return dep
}

func (d *Depends) dependencies() (list []string) {
	for _, v := range d.libs {
		list = append(list, v)
	}
	return
}

// Install libraries to outDir
func (d *Depends) Install(outDir string, qtPlugin bool) error {

	err := os.MkdirAll(outDir, 0755)
	if nil != err {
		return err
	}
	err = os.MkdirAll(outDir+"/libs", 0755)
	if nil != err {
		return err
	}
	err = os.MkdirAll(outDir+"/plugins", 0755)
	if nil != err {
		return err
	}

	if qtPlugin {
		qtPluginRoot, qtPluginList := getQtPluginFileList()
		for _, so := range qtPluginList {
			d.getSharedLibraryDependencies(so)
		}

		// install qt plugin
		for _, so := range qtPluginList {
			rel, _ := filepath.Rel(qtPluginRoot, so)
			dest := filepath.Join(outDir, "plugins", rel)
			dir := filepath.Dir(dest)
			os.MkdirAll(dir, 0755)
			err = copyFile(so, dest)
			if nil != err {
				return err
			}
		}
	}

	for _, v := range d.dependencies() {
		f := strings.Split(v, "/")
		filename := outDir + "/libs/" + f[len(f)-1]
		err = copyFile(v, filename)
		if nil != err {
			return err
		}
	}
	f := strings.Split(d.binPath, "/")
	filename := outDir + "/" + f[len(f)-1]
	err = copyFile(d.binPath, filename)
	if nil != err {
		return err
	}

	bashFilename := filename + ".bash"
	bashFileContent := fmt.Sprintf(bashTemplate, f[len(f)-1])
	bashFile, err := os.OpenFile(bashFilename, os.O_CREATE|os.O_WRONLY, 0755)
	if nil != err {
		return err
	}
	_, err = bashFile.Write([]byte(bashFileContent))
	if nil != err {
		return err
	}
	return bashFile.Close()
}

func (d *Depends) getSharedLibraryDependencies(binPath string) {
	cmd := exec.Command("ldd", binPath)
	data, err := cmd.Output()
	if nil != err {
		fmt.Println(err)
		return
	}

	list := []string{}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.Index(line, "=>") > 0 {
			sos := strings.Split(line, " ")
			for _, v := range sos {
				if strings.Index(v, "/") == 0 {
					filename := filepath.Base(v)
					_, hasFound := d.libs[v]
					_, isBlackList := d.black[filename]
					if !hasFound && !isBlackList {
						list = append(list, v)
					}
				}
			}
		}
	}

	if len(list) > 0 {
		for _, v := range list {
			d.libs[v] = v
		}
		for _, v := range list {
			d.getSharedLibraryDependencies(v)
		}
	}
}
