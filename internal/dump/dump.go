package dump

import (
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
)

type DumpInfo struct {
	Environment string
	Filename    string
	Path        string
	Size        int64
	CreatedAt   time.Time
	Compressed  bool
}

func ExecuteDump(env string, envCfg config.Environment, dumpCfg config.DumpConfig, sshLocalPort int) (string, error) {
	if err := os.MkdirAll(dumpCfg.OutputDir, 0700); err != nil {
		return "", fmt.Errorf("덤프 디렉토리 생성 실패: %w", err)
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s_%s.sql", envCfg.Database, env, timestamp)
	if dumpCfg.Compress {
		filename += ".gz"
	}
	dumpPath := filepath.Join(dumpCfg.OutputDir, filename)

	host := envCfg.Host
	port := fmt.Sprintf("%d", envCfg.Port)
	if sshLocalPort > 0 {
		host = "127.0.0.1"
		port = fmt.Sprintf("%d", sshLocalPort)
	}

	args := []string{
		"-h", host,
		"-P", port,
		"-u", envCfg.User,
		fmt.Sprintf("-p%s", envCfg.Password),
		"--single-transaction",
		"--routines",
		"--triggers",
		"--set-gtid-purged=OFF",
		"--column-statistics=0",
		"--no-tablespaces",
		envCfg.Database,
	}

	fmt.Printf("  mysqldump 실행 중 (%s:%s/%s)...\n", host, port, envCfg.Database)

	cmd := exec.Command("mysqldump", args...)
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("mysqldump stdout 파이프 생성 실패: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("mysqldump 실행 실패: %w\nmysqldump이 설치되어 있는지 확인하세요 (brew install mysql-client)", err)
	}

	outFile, err := os.Create(dumpPath)
	if err != nil {
		cmd.Process.Kill()
		return "", fmt.Errorf("덤프 파일 생성 실패: %w", err)
	}
	defer outFile.Close()

	var writer io.Writer = outFile
	if dumpCfg.Compress {
		gzWriter := gzip.NewWriter(outFile)
		defer gzWriter.Close()
		writer = gzWriter
	}

	if _, err := io.Copy(writer, stdout); err != nil {
		cmd.Process.Kill()
		return "", fmt.Errorf("덤프 데이터 저장 실패: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		os.Remove(dumpPath)
		return "", fmt.Errorf("mysqldump 실패: %w", err)
	}

	return dumpPath, nil
}

func ImportToDocker(dumpPath string, containerName string, rootPassword string, database string, compressed bool) error {
	fmt.Printf("  로컬 Docker MySQL에 임포트 중...\n")

	// Drop and recreate database for clean import
	dropSQL := fmt.Sprintf("DROP DATABASE IF EXISTS `%s`; CREATE DATABASE `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", database, database)
	dropCmd := exec.Command("docker", "exec", "-i", containerName,
		"mysql", fmt.Sprintf("-p%s", rootPassword), "-e", dropSQL)
	if out, err := dropCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("데이터베이스 초기화 실패: %s\n%s", err, string(out))
	}

	var reader io.Reader
	f, err := os.Open(dumpPath)
	if err != nil {
		return fmt.Errorf("덤프 파일 열기 실패: %w", err)
	}
	defer f.Close()

	if compressed {
		gzReader, err := gzip.NewReader(f)
		if err != nil {
			return fmt.Errorf("gzip 압축 해제 실패: %w", err)
		}
		defer gzReader.Close()
		reader = gzReader
	} else {
		reader = f
	}

	importCmd := exec.Command("docker", "exec", "-i", containerName,
		"mysql", fmt.Sprintf("-p%s", rootPassword), database)
	importCmd.Stdin = reader
	importCmd.Stderr = os.Stderr

	if err := importCmd.Run(); err != nil {
		return fmt.Errorf("임포트 실패: %w", err)
	}

	return nil
}

func ListDumps(outputDir string) ([]DumpInfo, error) {
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		return nil, nil
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		return nil, fmt.Errorf("덤프 디렉토리 읽기 실패: %w", err)
	}

	var dumps []DumpInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") && !strings.HasSuffix(name, ".sql.gz") {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		env := "unknown"
		parts := strings.Split(name, "_")
		if len(parts) >= 3 {
			// format: database_env_timestamp.sql(.gz)
			env = parts[1]
		}

		dumps = append(dumps, DumpInfo{
			Environment: env,
			Filename:    name,
			Path:        filepath.Join(outputDir, name),
			Size:        info.Size(),
			CreatedAt:   info.ModTime(),
			Compressed:  strings.HasSuffix(name, ".gz"),
		})
	}

	sort.Slice(dumps, func(i, j int) bool {
		return dumps[i].CreatedAt.After(dumps[j].CreatedAt)
	})

	return dumps, nil
}

func FormatSize(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}