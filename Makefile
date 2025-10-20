# 辅助工具安装列表
# go install github.com/cloudwego/hertz/cmd/hz@latest
# go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
# go install github.com/hertz-contrib/swagger-generate/thrift-gen-http-swagger@latest

# 默认输出帮助信息
.DEFAULT_GOAL := help

# 纯 Windows 环境：不强制依赖 WSL/Git Bash，默认使用 cmd，必要逻辑用 PowerShell 执行
# 项目 MODULE 名
MODULE = github.com/FantasyRL/go-mcp-demo
REMOTE_REPOSITORY ?= fantasyrl/go-mcp-demo
# 目录相关（避免在 Windows 下调用 pwd 失败，使用内置 CURDIR）
DIR = $(CURDIR)
CMD = $(DIR)/cmd
CONFIG_PATH = $(DIR)/config
IDL_PATH = $(DIR)/idl
OUTPUT_PATH = $(DIR)/output
API_PATH= $(DIR)/cmd/api
# Docker 网络名称
DOCKER_NET := go-mcp-net
# Docker 镜像前缀和标签
IMAGE_PREFIX ?= hachimi
TAG          ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo dev)

# 服务名
SERVICES := host mcp_server
service = $(word 1, $@)

# hertz HTTP脚手架
# init: hz new -idl ./idl/api.thrift -mod github.com/FantasyRL/go-mcp-demo -handler_dir ./api/handler -model_dir ./api/model -router_dir ./api/router
.PHONY: hertz-gen-api
hertz-gen-api:
	hz update -idl ${IDL_PATH}/api.thrift; \
	rm -rf $(DIR)/swagger; \
    thriftgo -g go -p http-swagger $(IDL_PATH)/api.thrift; \
    rm -rf $(DIR)/gen-go

.PHONY: $(SERVICES)
$(SERVICES):
	go run $(CMD)/$(service) -cfg $(CONFIG_PATH)/config.yaml

.PHONY: vendor
vendor:
	@echo ">> go mod tidy && go mod vendor"
	go mod tidy
	go mod vendor

.PHONY: docker-build-%
docker-build-%: vendor
	@echo ">> Building image for service: $* (tag: $(TAG))"
	docker build \
	  --build-arg SERVICE=$* \
	  -f docker/Dockerfile \
	  -t $(IMAGE_PREFIX)/$*:$(TAG) \
	  .

# 创建 Docker 网络供容器间HTTP通信
.PHONY: docker-net
docker-net:
ifeq ($(OS),Windows_NT)
	@powershell -NoProfile -ExecutionPolicy Bypass -Command "docker network inspect $(DOCKER_NET) *> $null; if ($$LASTEXITCODE -ne 0) { docker network create $(DOCKER_NET) | Out-Null }"
else
	@docker network inspect $(DOCKER_NET) >/dev/null 2>&1 || docker network create $(DOCKER_NET)
endif

.PHONY: docker-run-%
docker-run-%: docker-build-% docker-net
ifeq ($(OS),Windows_NT)
		@echo ">> Running docker (STRICT config - Windows)"
		@powershell -NoProfile -ExecutionPolicy Bypass -File "$(DIR)\scripts\docker-run.ps1" -Service "$*" -Image "$(IMAGE_PREFIX)/$*:$(TAG)" -ConfigPath "$(CONFIG_PATH)\config.yaml"
else
		@echo ">> Running docker (STRICT config - Linux)"
		CFG_SRC="$(CONFIG_PATH)/config.yaml"; \
		if [ ! -f "$$CFG_SRC" ]; then \
			echo "ERROR: $$CFG_SRC not found. Please create it." >&2; \
			exit 2; \
		fi; \
		docker rm -f $* >/dev/null 2>&1 || true; \
		docker run --rm -itd \
			--name $* \
			--network host \
			-e SERVICE=$* \
			-e TZ=Asia/Shanghai \
			-v "$$CFG_SRC":/app/config/config.yaml:ro \
			$(IMAGE_PREFIX)/$*:$(TAG)
endif

.PHONY: pull-run-%
pull-run-%:
ifeq ($(OS),Windows_NT)
		@echo ">> Pulling and running docker (STRICT config - Windows): $*"
		@docker pull $(REMOTE_REPOSITORY):$*
		@powershell -NoProfile -ExecutionPolicy Bypass -File "$(DIR)\scripts\docker-run.ps1" -Service "$*" -Image "$(REMOTE_REPOSITORY):$*" -ConfigPath "$(CONFIG_PATH)\config.yaml"
else
		@echo ">> Pulling and running docker (STRICT config - Linux): $*"
		@docker pull $(REMOTE_REPOSITORY):$*
		@CFG_SRC="$(CONFIG_PATH)/config.yaml"; \
		if [ ! -f "$$CFG_SRC" ]; then \
			echo "ERROR: $$CFG_SRC not found. Please create it." >&2; \
			exit 2; \
		fi; \
		docker rm -f $* >/dev/null 2>&1 || true; \
		docker run --rm -itd \
			--name $* \
			--network host \
			-e SERVICE=$* \
			-e TZ=Asia/Shanghai \
			-v "$$CFG_SRC":/app/config/config.yaml:ro \
			$(REMOTE_REPOSITORY):$*
endif

# 帮助信息
.PHONY: help
help:
	@echo "Available targets:"; \
	echo "  host                 - go run cmd/host with config.yaml"; \
	echo "  mcp_server           - go run cmd/mcp_server with config.yaml"; \
	echo "  vendor               - go mod tidy && vendor"; \
	echo "  docker-build-<svc>   - build image for service (host|mcp_server)"; \
	echo "  docker-run-<svc>     - run container (Windows自动映射端口, Linux使用--network host)"; \
	echo "  pull-run-<svc>       - pull and run container (同上)"; \
	echo "  stdio                - build mcp_server and run host with stdio config"; \
	echo "  push-<svc>           - push image to remote repo"


.PHONY: stdio
stdio:
	go build -o bin/mcp_server ./cmd/mcp_server # windows的output需要是.exe，并且在config.stdio.yaml中修改，bin/mcp-server.exe
	go run ./cmd/host -cfg $(CONFIG_PATH)/config.stdio.yaml

.PHONY: push-%
push-%:
	@read -p "Confirm service name to push (type '$*' to confirm): " CONFIRM_SERVICE; \
	if [ "$$CONFIRM_SERVICE" != "$*" ]; then \
		echo "Confirmation failed. Expected '$*', but got '$$CONFIRM_SERVICE'."; \
		exit 1; \
	fi; \
	if echo "$(SERVICES)" | grep -wq "$*"; then \
		if [ "$(ARCH)" = "x86_64" ] || [ "$(ARCH)" = "amd64" ]; then \
			echo "Building and pushing $* for amd64 architecture..."; \
			docker build --build-arg SERVICE=$* -t $(REMOTE_REPOSITORY):$* -f docker/Dockerfile .; \
			docker push $(REMOTE_REPOSITORY):$*; \
		else \
			echo "Building and pushing $* using buildx for amd64 architecture..."; \
			docker buildx build --platform linux/amd64 --build-arg SERVICE=$* -t $(REMOTE_REPOSITORY):$* -f docker/Dockerfile --push .; \
		fi; \
	else \
		echo "Service '$*' is not a valid service. Available: [$(SERVICES)]"; \
		exit 1; \
	fi


