package cmd

import (
	"fmt"
	"os"

	"axfrcheck/pkg"

	"github.com/VintageOps/dns-zone-compare/pkg/utils"
	"github.com/urfave/cli/v2"
)

func Execute() {
	app := &cli.App{
		Name:            "axfrcheck",
		Usage:           "check the master zones defined as slave in a given configuration",
		HideHelpCommand: true,
		Action: func(c *cli.Context) error {
			if c.NArg() != 1 {
				fmt.Printf("ERROR: Needs one argument, the nameds/pdns configuration file, %d provided", c.NArg())
				cli.ShowAppHelpAndExit(c, 1)
			}
			return pkg.CheckMasters()
		},
	}
	err := app.Run(os.Args)
	utils.FatalOnErr(err)
}
