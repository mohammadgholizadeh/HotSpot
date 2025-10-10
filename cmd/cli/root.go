package main

import (
	"github.com/spf13/cobra"
)

var (
	cfgPath   string
	portFlag  string
	bodyLimit int64
)

var rootCmd = &cobra.Command{
	Use:   "hotspot",
	Short: "HotSpot service CLI",
	RunE:  func(cmd *cobra.Command, args []string) error { return cmd.Help() },
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgPath, "config", "c", "configs/config.yaml", "Path to config file")
	rootCmd.PersistentFlags().StringVar(&portFlag, "port", "", "Override HTTP listen port (optional)")
	rootCmd.PersistentFlags().Int64Var(&bodyLimit, "body-limit-bytes", 1048576, "Maximum request body size in bytes")
}

func Execute() error { return rootCmd.Execute() }

func CfgPath() string { return cfgPath }

func PortOverride() string { return portFlag }

func BodyLimit() int64 { return bodyLimit }
