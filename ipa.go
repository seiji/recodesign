package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type Ipa struct {
	Dir      string
	FilePath string
	FileName string
	AppPath  string
	AppName  string
}

func (i *Ipa) ReplaceProvision(path string) error {
	dst := filepath.Join(i.AppPath, "embedded.mobileprovision")
	_, err := CopyFile(path, dst)
	return err
}

func (i *Ipa) CodeSign(sign string) error {
	path, err := exec.LookPath("codesign")
	if err != nil {
		log.Fatal(err)
	}
	curDir, _ := filepath.Abs(".")
	defer os.Chdir(curDir)

	os.Chdir(i.Dir)
	err = SystemCommand(path,
		"--force",
		"--sign", sign,
		"--resource-rules",
		filepath.Join("Payload", i.AppName, "ResourceRules.plist"),
		filepath.Join("Payload", i.AppName))
	if err != nil {
		return err
	}

	err = SystemCommand("zip", "-ry", filepath.Join(curDir, i.FileName), "Payload")
	return nil
}

func (i *Ipa) Close() {
	os.RemoveAll(i.Dir)
}

func NewIpa(path string) *Ipa {
	var name string
	if lastIndex := strings.LastIndex(path, string(os.PathSeparator)); lastIndex > -1 {
		name = path[lastIndex+1:]
	}

	t := time.Now()
	dst := fmt.Sprintf("%s-%d", name, t.Unix())
	SystemCommand("unzip", "-o", "-d", dst, path)

	var appPath string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".app" {
			appPath = path
			return filepath.SkipDir
		}
		return nil
	}
	err := filepath.Walk(dst, walkFn)
	if err != nil {
		log.Fatal(err)
	}

	var appName string
	if lastIndex := strings.LastIndex(appPath, string(os.PathSeparator)); lastIndex > -1 {
		appName = appPath[lastIndex+1:]
	}
	i := &Ipa{dst, path, name, appPath, appName}

	return i
}

func SystemCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	fmt.Printf("%s\n", out.String())
	if err != nil {
		log.Fatal(err)
		return err
	}
	return nil
}

func CopyFile(dst, src string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}
