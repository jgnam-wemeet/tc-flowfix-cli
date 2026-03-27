# Feature Backlog

추가 예정 기능 목록

## 우선순위 높음

### `flowfix db restore <file>`
- 기존 덤프 파일을 로컬 Docker MySQL에 다시 임포트
- `flowfix db list`에서 파일 선택하여 복원

### `flowfix db status`
- 원격 DB (production/staging) 연결 상태 확인
- 로컬 Docker MySQL 상태 확인
- SSH 터널 연결 테스트

## 우선순위 중간

### `flowfix db clean`
- 오래된 덤프 파일 자동 정리 (용량 관리)
- 예: `flowfix db clean --keep 5` (최근 5개만 유지)

### `flowfix db dump --tables users,orders`
- 특정 테이블만 선택적으로 덤프
- 대용량 DB에서 필요한 테이블만 빠르게 덤프

### `flowfix config edit`
- config.yaml을 기본 에디터로 바로 열기

## 우선순위 낮음

### `flowfix version`
- 현재 설치된 버전 표시

### `flowfix update`
- GitHub Releases에서 최신 바이너리 다운로드 & 자동 업데이트