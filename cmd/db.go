package cmd

import (
	"github.com/spf13/cobra"
)

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "데이터베이스 관련 명령어",
}

func init() {
	dbCmd.AddCommand(dbDumpCmd)
	dbCmd.AddCommand(dbListCmd)
}