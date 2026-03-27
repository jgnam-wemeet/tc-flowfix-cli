# CLAUDE.md

## 프로젝트 개요
tc-flowfix-platform의 비즈니스 유틸리티 CLI (Go)

## 기술 스택
- Go 1.26+
- cobra (CLI 프레임워크)
- yaml.v3 (설정 관리)

## 빌드 & 실행
```bash
go build -o flowfix .
./flowfix --help
```

## 프로젝트 구조
```
tc-flowfix-cli/
├── main.go                  # 진입점
├── cmd/                     # CLI 명령어
│   ├── root.go              # 루트 명령어
│   ├── db.go                # db 서브커맨드
│   ├── db_dump.go           # db dump 구현
│   ├── db_list.go           # db list 구현
│   ├── config.go            # config 서브커맨드
│   └── config_init/show     # config init/show 구현
├── internal/
│   ├── config/config.go     # 설정 파일 관리
│   ├── docker/docker.go     # Docker 컨테이너 관리
│   └── dump/dump.go         # mysqldump 실행 & 임포트
└── go.mod
```

## 설정 파일
~/.flowfix/config.yaml

## 명령어
- `flowfix db dump <env>` - 원격 DB를 로컬 Docker로 덤프
- `flowfix db list` - 덤프 파일 목록
- `flowfix config init` - 설정 초기화
- `flowfix config show` - 설정 표시

## 코드 컨벤션
- 한국어 에러 메시지 (사용자 대상)
- internal/ 패키지로 내부 로직 분리
- cmd/ 패키지로 CLI 인터페이스 분리
