# tc-flowfix-cli

tc-flowfix-platform 비즈니스 유틸리티 CLI

원격 DB(production/staging)를 로컬 Docker MySQL로 덤프 & 복원하는 도구입니다.

## 설치

```bash
# 빌드
go build -o flowfix .

# 또는 글로벌 설치
go install .
ln -s ~/go/bin/tc-flowfix-cli ~/go/bin/flowfix
```

## 설정

### 초기화

```bash
flowfix config init
```

`~/.flowfix/config.yaml` 파일이 생성됩니다. 환경별 DB 접속 정보를 입력하세요.

### 설정 확인

```bash
flowfix config show
```

### 설정 파일 예시 (`~/.flowfix/config.yaml`)

```yaml
environments:
  production:
    host: 127.0.0.1
    port: 3306
    user: admin
    password: your-password
    database: tc_logistics
    ssh_tunnel:
      enabled: true
      host: 211.188.53.49
      port: 22
      user: platform
      key_path: ~/.ssh/pemkey.pem
  staging:
    host: 127.0.0.1
    port: 3306
    user: admin
    password: your-password
    database: tc_logistics
    ssh_tunnel:
      enabled: true
      host: 211.188.53.49
      port: 22
      user: platform
      key_path: ~/.ssh/pemkey.pem
local:
  docker:
    container_name: mysql
    image: mysql:8.0
    port: 3306
    root_password: your-password
    network: tc-flowfix-platform
dump:
  output_dir: ~/.flowfix/dumps
  compress: true
```

## 명령어

| 명령어 | 설명 |
|--------|------|
| `flowfix db dump <env>` | 원격 DB를 로컬 Docker MySQL로 덤프 & 복원 |
| `flowfix db dump <env> --skip-import` | 덤프 파일만 생성 (로컬 임포트 생략) |
| `flowfix db list` | 덤프 파일 목록 조회 |
| `flowfix config init` | 설정 파일 초기화 |
| `flowfix config show` | 현재 설정 표시 |

`<env>`는 `production` 또는 `staging`

## 동작 흐름 (`db dump`)

1. **SSH 터널 설정** - `ssh_tunnel.enabled: true`이면 자동으로 SSH 터널 생성
2. **mysqldump 실행** - 원격 DB를 `~/.flowfix/dumps/`에 덤프 (gzip 압축)
3. **Docker MySQL 준비** - 로컬 컨테이너 확인 및 자동 생성/시작
4. **덤프 임포트** - 덤프 파일을 로컬 Docker MySQL에 복원

## 사전 요구사항

- Go 1.26+
- Docker
- mysqldump (`brew install mysql-client`)