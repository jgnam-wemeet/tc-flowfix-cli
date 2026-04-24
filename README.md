# tc-flowfix-cli

tc-flowfix-platform 비즈니스 유틸리티 CLI

원격 DB(production/staging)를 로컬 Docker MySQL로 덤프 & 복원하는 도구입니다.

## 설치

```bash
curl -sSL https://raw.githubusercontent.com/jgnam-wemeet/tc-flowfix-cli/main/install.sh | bash
```

### 소스에서 빌드

```bash
go build -o flowfix .
sudo mv flowfix /usr/local/bin/
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
    host: your-db-host
    port: 3306
    user: your-user
    password: your-password
    database: your-database
    ssh_tunnel:
      enabled: true
      host: your-ssh-host
      port: 22
      user: your-ssh-user
      key_path: ~/.ssh/your-key.pem
  staging:
    host: your-db-host
    port: 3306
    user: your-user
    password: your-password
    database: your-database
    ssh_tunnel:
      enabled: true
      host: your-ssh-host
      port: 22
      user: your-ssh-user
      key_path: ~/.ssh/your-key.pem
local:
  docker:
    container_name: mysql
    image: mysql:8.0
    port: 3306
    root_password: your-password
    network: your-network
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
| `flowfix db backup` | 로컬 Docker MySQL 데이터베이스를 파일로 백업 |
| `flowfix db restore` | 덤프 파일을 선택하여 로컬 Docker MySQL에 복원 |
| `flowfix config init` | 설정 파일 초기화 |
| `flowfix config show` | 현재 설정 표시 |

`<env>`는 `production` 또는 `staging`

## 동작 흐름 (`db backup`)

1. **로컬 Docker MySQL 확인** - 컨테이너 실행 상태 확인
2. **mysqldump 실행** - `docker exec`를 통해 로컬 DB를 `~/.flowfix/dumps/`에 백업 (gzip 압축)

## 동작 흐름 (`db restore`)

1. **덤프 파일 목록 표시** - `~/.flowfix/dumps/` 내 파일 목록을 번호와 함께 표시
2. **번호 선택** - 복원할 파일 번호를 입력
3. **Docker MySQL 준비** - 로컬 컨테이너 확인 및 자동 생성/시작
4. **덤프 임포트** - 선택한 덤프 파일을 로컬 Docker MySQL에 복원

## 동작 흐름 (`db dump`)

1. **SSH 터널 설정** - `ssh_tunnel.enabled: true`이면 자동으로 SSH 터널 생성
2. **mysqldump 실행** - 원격 DB를 `~/.flowfix/dumps/`에 덤프 (gzip 압축)
3. **Docker MySQL 준비** - 로컬 컨테이너 확인 및 자동 생성/시작
4. **덤프 임포트** - 덤프 파일을 로컬 Docker MySQL에 복원

## 사전 요구사항

- Go 1.26+
- Docker
- mysqldump (`brew install mysql-client`)