/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"fmt"
	"os"

	"github.com/inconshreveable/mousetrap"
	"github.com/sdir/moba/moba"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "moba",
	Short: "Moba parse MobaXterm config",
	Args:  cobra.MaximumNArgs(1),

	RunE: func(cmd *cobra.Command, args []string) error {
		defaultIni := "MobaXterm.ini"
		if len(args) == 1 {
			defaultIni = args[0]
		}

		moba, err := moba.NewMoba(defaultIni)
		if err != nil {
			return err
		}
		moba.ShowPasswords()
		moba.ShowCredentials()
		return nil
	},

	PostRun: func(cmd *cobra.Command, args []string) {
		if mousetrap.StartedByExplorer() {
			cmd.Println("Press return to continue...")
			fmt.Scanln()
			os.Exit(1)
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {

	cobra.MousetrapHelpText = ""
}
