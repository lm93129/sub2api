.PHONY: build generate test test-unit test-integration test-e2e

VERSION ?= $(shell ./scripts/resolve-version.sh)
LDFLAGS ?= -s -w -X main.Version=$(VERSION)

# ---------------------------------------------------------------------------
# 编译/测试的并发度与内存上限
#
# 本项目体量较大：internal/service 包有 591 个文件，ent/mutation.go 达 5.7 万行，
# 全项目约 116 万行 Go 代码。Go 按包整体编译，默认并行编译时单个 compile 进程
# 内存可飙升到 2GB+，在 4C/4G 的机器上会被 cgroup OOM Kill：
#   Memory cgroup out of memory: Killed process (compile) anon-rss:1975924kB
#
# 因此默认限制并发度并用 GOMEMLIMIT 迫使编译器更频繁 GC。
# 内存充裕的机器可覆盖，例如：
#   make build GO_BUILD_PARALLELISM=4 GOMEMLIMIT=4GiB
# ---------------------------------------------------------------------------
GO_BUILD_PARALLELISM ?= 1
GOMEMLIMIT ?= 1200MiB
GOGC ?= 40
GO_MEM_ENV = CGO_ENABLED=0 GOMEMLIMIT=$(GOMEMLIMIT) GOGC=$(GOGC)

build:
	$(GO_MEM_ENV) go build -p $(GO_BUILD_PARALLELISM) -ldflags="$(LDFLAGS)" -trimpath -o bin/server ./cmd/server

generate:
	$(GO_MEM_ENV) go generate -p $(GO_BUILD_PARALLELISM) ./ent
	$(GO_MEM_ENV) go generate -p $(GO_BUILD_PARALLELISM) ./cmd/server

test:
	$(GO_MEM_ENV) go test -p $(GO_BUILD_PARALLELISM) ./...
	golangci-lint run ./...

test-unit:
	$(GO_MEM_ENV) go test -p $(GO_BUILD_PARALLELISM) -tags=unit ./...

test-integration:
	$(GO_MEM_ENV) go test -p $(GO_BUILD_PARALLELISM) -tags=integration ./...

test-e2e:
	./scripts/e2e-test.sh

test-e2e-local:
	$(GO_MEM_ENV) go test -p $(GO_BUILD_PARALLELISM) -tags=e2e -v -timeout=300s ./internal/integration/...
