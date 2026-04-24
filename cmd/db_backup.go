package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
	"github.com/namjeonggil/tc-flowfix-cli/internal/docker"
	"github.com/namjeonggil/tc-flowfix-cli/internal/dump"
	"github.com/spf13/cobra"
)

var dbBackupCmd = &cobra.Command{
	Use:   "backup",
	Short: "로컬 Docker MySQL 데이터베이스를 파일로 백업",
	Long: `로컬 Docker MySQL 컨테이너의 데이터베이스를 .sql 파일로 백업합니다.

Examples:
  flowfix db backup`,
	RunE: runDBBackup,
}

func runDBBackup(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	fmt.Println("[1/2] 로컬 Docker MySQL 확인 중...")
	if !docker.IsContainerRunning(cfg.Local.Docker.ContainerName) {
		return fmt.Errorf("Docker 컨테이너 '%s'이(가) 실행 중이 아닙니다.\n'flowfix db dump <env>'를 먼저 실행하세요", cfg.Local.Docker.ContainerName)
	}
	fmt.Printf("  ✓ 컨테이너 실행 중 (%s)\n", cfg.Local.Docker.ContainerName)

	database := "tc_logistics"
	for _, envCfg := range cfg.Environments {
		if envCfg.Database != "" {
			database = envCfg.Database
			break
		}
	}

	fmt.Println("[2/2] 데이터베이스 백업 중...")
	startTime := time.Now()
	dumpPath, err := dump.BackupLocal(
		cfg.Local.Docker.ContainerName,
		cfg.Local.Docker.RootPassword,
		database,
		cfg.Dump,
	)
	if err != nil {
		return err
	}

	if fi, err := os.Stat(dumpPath); err == nil {
		fmt.Printf("  ✓ 백업 완료: %s (%s)\n", dumpPath, dump.FormatSize(fi.Size()))
	} else {
		fmt.Printf("  ✓ 백업 완료: %s\n", dumpPath)
	}

	fmt.Printf("  소요 시간: %s\n", time.Since(startTime).Round(time.Second))
	return nil
}
