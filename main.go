package main

import (
	"errors"
	"fmt"
	"github.com/codegangsta/cli"
	"os"
)

var Version string = "HEAD"

var Flags = []cli.Flag{
	flagIpa,
	flagProvision,
	flagSign,
}
var flagIpa = cli.StringFlag{
	Name:  "ipa, i",
	Value: "./xxxxx.ipa",
	Usage: "ipa path for resigning",
}
var flagProvision = cli.StringFlag{
	Name:  "provision, p",
	Value: "./xxxxx.mobileprovision",
	Usage: "provision path for resigning",
}
var flagSign = cli.StringFlag{
	Name:  "sign, s",
	Value: "iPhone Distribution: XXXXX XXXXX (XXXXXXXXXX)",
	Usage: "codesign for resigning # security find-identity -p codesigning -v",
}

func main() {
	newApp().Run(os.Args)
}

func perror(err error) {
	if err != nil {
		panic(err)
	}
}

func isFile(path string) error {
	var err error = nil
	if fileInfo, e := os.Stat(path); e != nil || fileInfo.IsDir() {
		err = errors.New(fmt.Sprintf("Invalid path %s", path))
	}
	return err
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "recodesign"
	app.Usage = "Resign your ipa"
	app.Version = Version
	app.Author = "Seiji Toyama"
	app.Email = "toyama.seiji@gmail.com"
	app.Flags = Flags

	app.Action = func(c *cli.Context) {
		ipaPath := c.String("ipa")
		provisionPath := c.String("provision")
		sign := c.String("sign")

		err := isFile(ipaPath)
		perror(err)
		err = isFile(provisionPath)
		perror(err)

		ipa := NewIpa(ipaPath)
		defer ipa.Close()
		ipa.ReplaceProvision(provisionPath)
		ipa.CodeSign(sign)
	}
	return app
}
