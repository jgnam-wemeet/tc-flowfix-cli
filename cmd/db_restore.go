package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
	"github.com/namjeonggil/tc-flowfix-cli/internal/docker"
	"github.com/namjeonggil/tc-flowfix-cli/internal/dump"
	"github.com/spf13/cobra"
)

var dbRestoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "덤프 파일을 선택하여 로컬 Docker MySQL에 복원",
	Long: `저장된 덤프 파일 목록을 보여주고, 번호로 선택하여 로컬 Docker MySQL에 복원합니다.

Examples:
  flowfix db restore`,
	RunE: runDBRestore,
}

func runDBRestore(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
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
	fmt.Printf("  %-4s %-12s %-50s %-10s %s\n", "#", "환경", "파일명", "크기", "생성일시")
	fmt.Println("  " + strings.Repeat("─", 90))

	for i, d := range dumps {
		fmt.Printf("  %-4d %-12s %-50s %-10s %s\n",
			i+1,
			d.Environment,
			d.Filename,
			dump.FormatSize(d.Size),
			d.CreatedAt.Format("2006-01-02 15:04:05"),
		)
	}

	fmt.Print("\n복원할 파일 번호를 선택하세요: ")
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return fmt.Errorf("입력을 읽을 수 없습니다")
	}

	input := strings.TrimSpace(scanner.Text())
	num, err := strconv.Atoi(input)
	if err != nil || num < 1 || num > len(dumps) {
		return fmt.Errorf("잘못된 번호입니다: %s (1~%d 사이의 번호를 입력하세요)", input, len(dumps))
	}

	selected := dumps[num-1]
	fmt.Println()

	fmt.Println("[1/2] 로컬 Docker MySQL 확인 중...")
	if err := docker.EnsureContainer(cfg.Local.Docker); err != nil {
		return err
	}
	fmt.Printf("  ✓ 컨테이너 실행 중 (%s)\n", cfg.Local.Docker.ContainerName)

	fmt.Println("[2/2] 데이터베이스 복원 중...")
	startTime := time.Now()

	database := extractDatabase(selected.Filename)
	if err := dump.ImportToDocker(
		selected.Path,
		cfg.Local.Docker.ContainerName,
		cfg.Local.Docker.RootPassword,
		database,
		selected.Compressed,
	); err != nil {
		return err
	}

	fmt.Printf("  ✓ 복원 완료: %s\n", selected.Filename)
	fmt.Printf("  소요 시간: %s\n", time.Since(startTime).Round(time.Second))
	return nil
}

func extractDatabase(filename string) string {
	// format: {database}_{env}_{timestamp}.sql(.gz)
	// e.g. tc_logistics_staging_20260420_153022.sql.gz
	name := strings.TrimSuffix(filename, ".gz")
	name = strings.TrimSuffix(name, ".sql")

	// Find the environment part by scanning from the end
	// timestamp is always YYYYMMDD_HHMMSS (15 chars)
	// so we strip that and the env name
	parts := strings.Split(name, "_")
	if len(parts) < 4 {
		return "tc_logistics"
	}

	// last two parts are timestamp (date_time), one before that is env
	// everything before env is the database name
	envIdx := len(parts) - 3
	if envIdx < 1 {
		return "tc_logistics"
	}

	return strings.Join(parts[:envIdx], "_")
}
