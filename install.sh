#!/bin/bash
set -e

REPO="jgnam-wemeet/tc-flowfix-cli"
INSTALL_DIR="/usr/local/bin"
BINARY="flowfix"

echo "==> tc-flowfix-cli 설치 중..."

# GitHub 최신 릴리즈에서 다운로드
LATEST_TAG=$(curl -sI "https://github.com/${REPO}/releases/latest" | grep -i "^location:" | sed 's/.*tag\///' | tr -d '\r\n')
if [ -z "$LATEST_TAG" ]; then
    echo "오류: 최신 릴리즈를 찾을 수 없습니다."
    exit 1
fi
echo "==> 버전: ${LATEST_TAG}"

DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${LATEST_TAG}/${BINARY}"
TMP_FILE=$(mktemp)

echo "==> 다운로드 중..."
if ! curl -sL -o "$TMP_FILE" "$DOWNLOAD_URL"; then
    echo "오류: 다운로드 실패"
    rm -f "$TMP_FILE"
    exit 1
fi

chmod +x "$TMP_FILE"

echo "==> ${INSTALL_DIR}/${BINARY} 에 설치 중..."
if [ -w "$INSTALL_DIR" ]; then
    mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY}"
else
    sudo mv "$TMP_FILE" "${INSTALL_DIR}/${BINARY}"
fi

# macOS quarantine 속성 제거
if [ "$(uname)" = "Darwin" ]; then
    xattr -d com.apple.quarantine "${INSTALL_DIR}/${BINARY}" 2>/dev/null || true
fi

echo "==> 설치 완료! (${INSTALL_DIR}/${BINARY})"

# 설정 파일이 없으면 자동 초기화
CONFIG_FILE="$HOME/.flowfix/config.yaml"
if [ ! -f "$CONFIG_FILE" ]; then
    echo "==> 설정 파일 초기화 중..."
    "${INSTALL_DIR}/${BINARY}" config init
fi

echo ""
echo "다음 단계:"
echo "  vim ~/.flowfix/config.yaml  # DB 접속 정보 입력"
