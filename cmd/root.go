package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "flowfix",
	Short: "tc-flowfix-platform 비즈니스 유틸리티 CLI",
	Long:  "tc-flowfix-platform 프로젝트의 운영/개발 유틸리티 도구입니다.",
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(dbCmd)
	rootCmd.AddCommand(configCmd)
}