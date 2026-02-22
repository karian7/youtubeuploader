# ============================================================
# Makefile - youtubeuploader
# ============================================================

BINARY_NAME = youtubeuploader
INSTALL_DIR = ~/bin
CMD_PATH = ./cmd/youtubeuploader
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

.PHONY: help build test install lint clean
.DEFAULT_GOAL := help

help: ## 사용 가능한 명령어 표시
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 바이너리 빌드
	go build -ldflags="-X main.appVersion=$(VERSION)" -o $(BINARY_NAME) $(CMD_PATH)

test: ## 테스트 실행 (race detection 포함)
	go test -v -race ./...

install: build ## 빌드 후 ~/bin에 설치
	install -m 755 $(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

lint: ## golangci-lint 실행
	golangci-lint run

clean: ## 빌드 산출물 정리
	rm -f $(BINARY_NAME)
