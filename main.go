package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var (
	version string = "0.0.1"
)

func main() {
	newApp().Run(os.Args)
}

func newApp() *cli.App {
	app := cli.NewApp()
	app.Name = "recodesign"
	app.Usage = "Resign your ipa"
	app.Version = version
	app.Author = "Seiji Toyama"
	app.Email = "toyama.seiji@gmail.com"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "identifier, i",
			Usage: "identifier for resigning",
		},
		cli.StringFlag{
			Name:  "provision, p",
			Usage: "provision path for resigning",
		},
		cli.StringFlag{
			Name:  "sign, s",
			Usage: "codesign for resigning (e.g. 'iPhone Distribution: XXXXX XXXXX (XXXXXXXXXX)')# security find-identity -p codesigning -v",
		},
	}

	app.Action = func(c *cli.Context) {
		identifier := c.String("identifier")
		provisionPath := c.String("provision")
		sign := c.String("sign")

		if len(c.Args()) <= 0 || len(provisionPath) <= 0 || len(sign) <= 0 {
			cli.ShowAppHelp(c)
			return
		}

		ipaPath := c.Args()[0]
		err := isFile(ipaPath)
		perror(err)
		err = isFile(provisionPath)
		perror(err)

		ipa := NewIpa(ipaPath)
		defer ipa.Close()
		ipa.ReplaceProvision(provisionPath)
		ipa.CodeSign(sign, identifier)
	}
	return app
}
