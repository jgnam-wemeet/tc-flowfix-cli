package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
	"github.com/namjeonggil/tc-flowfix-cli/internal/docker"
	"github.com/namjeonggil/tc-flowfix-cli/internal/dump"
	"github.com/spf13/cobra"
)

var skipImport bool

var dbDumpCmd = &cobra.Command{
	Use:   "dump <environment>",
	Short: "원격 DB를 로컬 Docker MySQL로 덤프",
	Long: `운영(production) 또는 스테이징(staging) 환경의 MySQL DB를
로컬 Docker 컨테이너로 덤프하고 복원합니다.

Examples:
  flowfix db dump staging
  flowfix db dump production
  flowfix db dump staging --skip-import  # 덤프만 수행 (임포트 생략)`,
	Args: cobra.ExactArgs(1),
	RunE: runDBDump,
}

func init() {
	dbDumpCmd.Flags().BoolVar(&skipImport, "skip-import", false, "덤프만 수행하고 로컬 임포트는 생략")
}

func runDBDump(cmd *cobra.Command, args []string) error {
	env := args[0]
	if env != "production" && env != "staging" {
		return fmt.Errorf("지원하지 않는 환경: %s (production 또는 staging만 가능)", env)
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	envCfg, ok := cfg.Environments[env]
	if !ok {
		return fmt.Errorf("환경 '%s'의 설정을 찾을 수 없습니다", env)
	}

	if err := envCfg.Validate(); err != nil {
		return fmt.Errorf("[%s] %w\n설정 파일을 확인하세요: %s", env, err, config.ConfigPath())
	}

	fmt.Printf("\n=== %s DB 덤프 시작 ===\n\n", env)
	startTime := time.Now()

	// Step 1: SSH tunnel if needed
	var sshLocalPort int
	var sshCmd *exec.Cmd
	if envCfg.SSHTunnel.Enabled {
		fmt.Println("[1/4] SSH 터널 설정 중...")
		localPort, cmd, err := setupSSHTunnel(envCfg)
		if err != nil {
			return fmt.Errorf("SSH 터널 설정 실패: %w", err)
		}
		sshLocalPort = localPort
		sshCmd = cmd
		defer func() {
			if sshCmd.Process != nil {
				sshCmd.Process.Kill()
				sshCmd.Wait()
				fmt.Println("  SSH 터널 종료")
			}
		}()
		fmt.Printf("  SSH 터널 연결됨 (localhost:%d -> %s:%d)\n", localPort, envCfg.Host, envCfg.Port)
	} else {
		fmt.Println("[1/4] SSH 터널 불필요 (직접 연결)")
	}

	// Step 2: Dump
	fmt.Println("\n[2/4] 데이터베이스 덤프 중...")
	dumpPath, err := dump.ExecuteDump(env, envCfg, cfg.Dump, sshLocalPort)
	if err != nil {
		return err
	}
	if fi, err := os.Stat(dumpPath); err == nil {
		fmt.Printf("  덤프 완료: %s (%s)\n", dumpPath, dump.FormatSize(fi.Size()))
	} else {
		fmt.Printf("  덤프 완료: %s\n", dumpPath)
	}

	if skipImport {
		fmt.Printf("\n=== 덤프 완료 (임포트 생략) ===\n")
		fmt.Printf("  파일: %s\n", dumpPath)
		fmt.Printf("  소요 시간: %s\n", time.Since(startTime).Round(time.Second))
		return nil
	}

	// Step 3: Ensure local Docker MySQL
	fmt.Println("\n[3/4] 로컬 Docker MySQL 준비 중...")
	if err := docker.EnsureContainer(cfg.Local.Docker); err != nil {
		return err
	}

	// Step 4: Import
	fmt.Println("\n[4/4] 덤프 데이터 임포트 중...")
	compressed := cfg.Dump.Compress
	if err := dump.ImportToDocker(dumpPath, cfg.Local.Docker.ContainerName, cfg.Local.Docker.RootPassword, envCfg.Database, compressed); err != nil {
		return err
	}

	elapsed := time.Since(startTime).Round(time.Second)
	fmt.Printf("\n=== %s DB 덤프 & 임포트 완료 ===\n", env)
	fmt.Printf("  소요 시간: %s\n", elapsed)
	fmt.Printf("  덤프 파일: %s\n", dumpPath)
	fmt.Printf("  로컬 접속: mysql -h 127.0.0.1 -P %d -u root -p'<password>' %s\n",
		cfg.Local.Docker.Port, envCfg.Database)

	return nil
}

func setupSSHTunnel(envCfg config.Environment) (int, *exec.Cmd, error) {
	tunnel := envCfg.SSHTunnel

	localPort, err := findFreePort()
	if err != nil {
		return 0, nil, fmt.Errorf("사용 가능한 포트를 찾을 수 없습니다: %w", err)
	}

	args := []string{
		"-N", "-L",
		fmt.Sprintf("%d:%s:%d", localPort, envCfg.Host, envCfg.Port),
		"-p", strconv.Itoa(tunnel.Port),
		fmt.Sprintf("%s@%s", tunnel.User, tunnel.Host),
		"-o", "ServerAliveInterval=60",
	}

	if tunnel.KeyPath != "" {
		// ~ 경로 확장
		keyPath := tunnel.KeyPath
		if len(keyPath) > 0 && keyPath[0] == '~' {
			home, _ := os.UserHomeDir()
			keyPath = home + keyPath[1:]
		}

		info, err := os.Stat(keyPath)
		if err != nil {
			return 0, nil, fmt.Errorf("SSH 키 파일을 찾을 수 없습니다: %s", keyPath)
		}

		// 권한 체크 (600 또는 400이어야 함)
		perm := info.Mode().Perm()
		if perm&0077 != 0 {
			return 0, nil, fmt.Errorf("SSH 키 파일 권한이 너무 개방적입니다: %s (%04o)\n  수정: chmod 600 %s", keyPath, perm, keyPath)
		}

		args = append([]string{"-i", keyPath}, args...)
	}

	cmd := exec.Command("ssh", args...)
	if err := cmd.Start(); err != nil {
		return 0, nil, err
	}

	// wait for tunnel to be ready
	for i := 0; i < 10; i++ {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", localPort), time.Second)
		if err == nil {
			conn.Close()
			return localPort, cmd, nil
		}
		time.Sleep(500 * time.Millisecond)
	}

	cmd.Process.Kill()
	cmd.Wait()
	return 0, nil, fmt.Errorf("SSH 터널이 시간 내에 준비되지 않았습니다")
}

func findFreePort() (int, error) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
