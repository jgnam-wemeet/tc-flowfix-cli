package cmd

import (
	"fmt"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
	"github.com/namjeonggil/tc-flowfix-cli/internal/dump"
	"github.com/spf13/cobra"
)

var dbListCmd = &cobra.Command{
	Use:   "list",
	Short: "로컬에 저장된 덤프 파일 목록",
	RunE:  runDBList,
}

func runDBList(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		// config 없어도 기본 경로에서 시도
		cfg = config.DefaultConfig()
	}

	dumps, err := dump.ListDumps(cfg.Dump.OutputDir)
	if err != nil {
		return err
	}

	if len(dumps) == 0 {
		fmt.Println("저장된 덤프 파일이 없습니다.")
		fmt.Printf("덤프 디렉토리: %s\n", cfg.Dump.OutputDir)
		return nil
	}

	fmt.Printf("덤프 파일 목록 (%s)\n\n", cfg.Dump.OutputDir)
	fmt.Printf("%-12s %-50s %-10s %s\n", "환경", "파일명", "크기", "생성일시")
	fmt.Println("─────────────────────────────────────────────────────────────────────────────────────────")

	for _, d := range dumps {
		compressed := ""
		if d.Compressed {
			compressed = " (gz)"
		}
		fmt.Printf("%-12s %-50s %-10s %s%s\n",
			d.Environment,
			d.Filename,
			dump.FormatSize(d.Size),
			d.CreatedAt.Format("2006-01-02 15:04:05"),
			compressed,
		)
	}

	fmt.Printf("\n총 %d개 파일\n", len(dumps))
	return nil
}
