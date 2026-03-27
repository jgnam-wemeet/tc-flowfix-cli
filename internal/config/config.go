package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type SSHTunnel struct {
	Enabled bool   `yaml:"enabled"`
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	User    string `yaml:"user"`
	KeyPath string `yaml:"key_path"`
}

type Environment struct {
	Host      string    `yaml:"host"`
	Port      int       `yaml:"port"`
	User      string    `yaml:"user"`
	Password  string    `yaml:"password"`
	Database  string    `yaml:"database"`
	SSHTunnel SSHTunnel `yaml:"ssh_tunnel"`
}

type DockerConfig struct {
	ContainerName string `yaml:"container_name"`
	Image         string `yaml:"image"`
	Port          int    `yaml:"port"`
	RootPassword  string `yaml:"root_password"`
	Network       string `yaml:"network"`
}

type LocalConfig struct {
	Docker DockerConfig `yaml:"docker"`
}

type DumpConfig struct {
	OutputDir string `yaml:"output_dir"`
	Compress  bool   `yaml:"compress"`
}

type Config struct {
	Environments map[string]Environment `yaml:"environments"`
	Local        LocalConfig            `yaml:"local"`
	Dump         DumpConfig             `yaml:"dump"`
}

func ConfigDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "경고: 홈 디렉토리를 찾을 수 없습니다: %v\n", err)
		return ".flowfix"
	}
	return filepath.Join(home, ".flowfix")
}

func ConfigPath() string {
	return filepath.Join(ConfigDir(), "config.yaml")
}

func DumpsDir() string {
	return filepath.Join(ConfigDir(), "dumps")
}

func MustConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("홈 디렉토리를 확인할 수 없습니다: %w", err)
	}
	return filepath.Join(home, ".flowfix"), nil
}

func DefaultConfig() *Config {
	return &Config{
		Environments: map[string]Environment{
			"production": {
				Host:     "",
				Port:     3306,
				User:     "",
				Password: "",
				Database: "tc_logistics",
				SSHTunnel: SSHTunnel{
					Enabled: false,
					Host:    "",
					Port:    22,
					User:    "",
					KeyPath: "",
				},
			},
			"staging": {
				Host:     "",
				Port:     3306,
				User:     "",
				Password: "",
				Database: "tc_logistics",
				SSHTunnel: SSHTunnel{
					Enabled: false,
					Host:    "",
					Port:    22,
					User:    "",
					KeyPath: "",
				},
			},
		},
		Local: LocalConfig{
			Docker: DockerConfig{
				ContainerName: "flowfix-mysql",
				Image:         "mysql:8.0",
				Port:          13306,
				RootPassword:  "flowfix2025!",
				Network:       "tc-flowfix-platform",
			},
		},
		Dump: DumpConfig{
			OutputDir: DumpsDir(),
			Compress:  true,
		},
	}
}

func Load() (*Config, error) {
	path := ConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("설정 파일을 읽을 수 없습니다 (%s): %w\n'flowfix config init'으로 설정을 초기화하세요", path, err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("설정 파일 파싱 실패: %w", err)
	}

	// resolve dump output dir
	if cfg.Dump.OutputDir == "" {
		cfg.Dump.OutputDir = DumpsDir()
	}

	return cfg, nil
}

func Save(cfg *Config) error {
	dir := ConfigDir()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("설정 디렉토리 생성 실패: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("설정 직렬화 실패: %w", err)
	}

	path := ConfigPath()
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("설정 파일 저장 실패: %w", err)
	}

	return nil
}

func (e *Environment) Validate() error {
	if e.Host == "" {
		return fmt.Errorf("DB 호스트가 설정되지 않았습니다")
	}
	if e.User == "" {
		return fmt.Errorf("DB 사용자가 설정되지 않았습니다")
	}
	if e.Password == "" {
		return fmt.Errorf("DB 비밀번호가 설정되지 않았습니다")
	}
	if e.Database == "" {
		return fmt.Errorf("데이터베이스 이름이 설정되지 않았습니다")
	}
	return nil
}