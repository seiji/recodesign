package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func perror(err error) {
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func isFile(path string) error {
	var err error = nil
	if fileInfo, e := os.Stat(path); e != nil || fileInfo.IsDir() {
		err = errors.New(fmt.Sprintf("Invalid path %s", path))
	}
	return err
}

type Ipa struct {
	Dir        string
	FilePath   string
	FileName   string
	AppPath    string
	AppName    string
	Identifier string
}

func (i *Ipa) ReplaceProvision(src string) error {
	dst := filepath.Join(i.AppPath, "embedded.mobileprovision")
	_, err := CopyFile(src, dst)
	return err
}

func (i *Ipa) CodeSign(sign, identifier string) error {
	if len(identifier) <= 0 {
		identifier = i.Identifier
	}

	path, err := exec.LookPath("codesign")
	perror(err)

	curDir, _ := filepath.Abs(".")
	defer os.Chdir(curDir)

	os.Chdir(i.Dir)
	err = SystemRun(path,
		"--force",
		"--sign", sign,
		"--identifier", i.Identifier,
		filepath.Join("Payload", i.AppName))
	if err != nil {
		return err
	}

	err = SystemRun("zip", "-ry", filepath.Join(curDir, i.FileName), "Payload")
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
	SystemRun("unzip", "-o", "-d", dst, path)

	var appPath string
	walkFn := func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".app" {
			appPath = path
			return filepath.SkipDir
		}
		return nil
	}
	err := filepath.Walk(dst, walkFn)
	perror(err)

	var appName string
	if lastIndex := strings.LastIndex(appPath, string(os.PathSeparator)); lastIndex > -1 {
		appName = appPath[lastIndex+1:]
	}

	identifier, err := SystemOutput("/usr/libexec/PlistBuddy", "-c", "Print :CFBundleIdentifier", filepath.Join(appPath, "Info.plist"))
	perror(err)

	i := &Ipa{dst, path, name, appPath, appName, identifier}

	return i
}

func SystemRun(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	fmt.Printf("%s\n", out.String())

	return err
}

func SystemOutput(name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	fmt.Printf("$ %s\n", strings.Join(cmd.Args, " "))
	var out bytes.Buffer
	cmd.Stderr = &out
	b, err := cmd.Output()
	fmt.Printf("%s\n", out.String())

	return string(b[:]), err
}

func CopyFile(src, dst string) (int64, error) {
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
