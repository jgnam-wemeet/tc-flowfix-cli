package cmd

import (
	"fmt"
	"os"

	"github.com/namjeonggil/tc-flowfix-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "설정 관리 명령어",
}

var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "설정 파일 초기화",
	Long:  "~/.flowfix/config.yaml 설정 파일을 기본값으로 생성합니다.",
	RunE:  runConfigInit,
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "현재 설정 표시",
	RunE:  runConfigShow,
}

func init() {
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigInit(cmd *cobra.Command, args []string) error {
	path := config.ConfigPath()

	if _, err := os.Stat(path); err == nil {
		fmt.Printf("설정 파일이 이미 존재합니다: %s\n", path)
		fmt.Println("덮어쓰려면 파일을 삭제 후 다시 실행하세요.")
		return nil
	}

	cfg := config.DefaultConfig()
	if err := config.Save(cfg); err != nil {
		return err
	}

	fmt.Printf("설정 파일이 생성되었습니다: %s\n\n", path)
	fmt.Println("다음 단계:")
	fmt.Println("  1. 설정 파일을 열어 DB 접속 정보를 입력하세요")
	fmt.Printf("     vi %s\n", path)
	fmt.Println("  2. 덤프를 실행하세요")
	fmt.Println("     flowfix db dump staging")

	return nil
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// mask passwords for display
	displayCfg := *cfg
	for name, env := range displayCfg.Environments {
		if env.Password != "" {
			env.Password = "****"
		}
		if env.SSHTunnel.KeyPath != "" {
			env.SSHTunnel.KeyPath = maskPath(env.SSHTunnel.KeyPath)
		}
		displayCfg.Environments[name] = env
	}
	displayCfg.Local.Docker.RootPassword = "****"

	data, err := yaml.Marshal(displayCfg)
	if err != nil {
		return fmt.Errorf("설정 표시 실패: %w", err)
	}

	fmt.Printf("설정 파일: %s\n\n", config.ConfigPath())
	fmt.Println(string(data))
	return nil
}

func maskPath(path string) string {
	if len(path) <= 10 {
		return path
	}
	return path[:5] + "..." + path[len(path)-5:]
}
