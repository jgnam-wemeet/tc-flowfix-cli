package docker

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
)

func IsContainerRunning(name string) bool {
	out, err := exec.Command("docker", "inspect", "-f", "{{.State.Running}}", name).Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

func ContainerExists(name string) bool {
	err := exec.Command("docker", "inspect", name).Run()
	return err == nil
}

func EnsureContainer(cfg config.DockerConfig) error {
	if IsContainerRunning(cfg.ContainerName) {
		fmt.Printf("  Docker 컨테이너 '%s' 이미 실행 중\n", cfg.ContainerName)
		return nil
	}

	if ContainerExists(cfg.ContainerName) {
		fmt.Printf("  Docker 컨테이너 '%s' 시작 중...\n", cfg.ContainerName)
		cmd := exec.Command("docker", "start", cfg.ContainerName)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("컨테이너 시작 실패: %s\n%s", err, string(out))
		}
		return waitForMySQL(cfg)
	}

	fmt.Printf("  Docker 컨테이너 '%s' 생성 중...\n", cfg.ContainerName)

	// ensure network exists
	exec.Command("docker", "network", "create", cfg.Network).Run()

	args := []string{
		"run", "-d",
		"--name", cfg.ContainerName,
		"--network", cfg.Network,
		"-p", fmt.Sprintf("%d:3306", cfg.Port),
		"-e", fmt.Sprintf("MYSQL_ROOT_PASSWORD=%s", cfg.RootPassword),
		"-e", "MYSQL_DATABASE=tc_logistics",
		"-e", "MYSQL_ROOT_HOST=%",
		cfg.Image,
		"--default-authentication-plugin=mysql_native_password",
		"--character-set-server=utf8mb4",
		"--collation-server=utf8mb4_unicode_ci",
		"--log-bin-trust-function-creators=1",
	}

	cmd := exec.Command("docker", args...)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("컨테이너 생성 실패: %s\n%s", err, string(out))
	}

	return waitForMySQL(cfg)
}

func waitForMySQL(cfg config.DockerConfig) error {
	fmt.Print("  MySQL 준비 대기 중")
	for i := 0; i < 30; i++ {
		cmd := exec.Command("docker", "exec", cfg.ContainerName,
			"mysqladmin", "ping", "-h", "localhost",
			fmt.Sprintf("-p%s", cfg.RootPassword), "--silent")
		if err := cmd.Run(); err == nil {
			fmt.Println(" OK")
			return nil
		}
		fmt.Print(".")
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("\nMySQL이 시간 내에 준비되지 않았습니다 (60초 초과)")
}

func ExecSQL(containerName, rootPassword, sql string) error {
	cmd := exec.Command("docker", "exec", "-i", containerName,
		"mysql", fmt.Sprintf("-p%s", rootPassword), "-e", sql)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("SQL 실행 실패: %s\n%s", err, string(out))
	}
	return nil
}