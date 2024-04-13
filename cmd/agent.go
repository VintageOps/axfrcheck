package cmd

import (
	pkg "axfrcheck/pkg"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "axfrcheck [config file]",
	Short: "AXFR check utility",
	Long:  `AXFR check utility checks the master zones defined as slave in a given configuration`,
	Args:  cobra.ExactArgs(1), // Requires exactly one argument
	Run: func(cmd *cobra.Command, args []string) {
		configFile := args[0]
		zones, err := pkg.ParseNamedConf(configFile)
		if err != nil {
			fmt.Printf("Could not parse %s", configFile)
		}
		viper.Unmarshal(zones)
		if err != nil {
			fmt.Printf("Unable to decode into config struct, %v", err)
		}

		pkg.CheckMasters(zones)
	},
}

func Execute() error {
	return rootCmd.Execute()
}
