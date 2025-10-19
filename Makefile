# 辅助工具安装列表
# go install github.com/cloudwego/hertz/cmd/hz@latest
# go install github.com/cloudwego/kitex/tool/cmd/kitex@latest
# go install github.com/hertz-contrib/swagger-generate/thrift-gen-http-swagger@latest

# 默认输出帮助信息
.DEFAULT_GOAL := help
# 项目 MODULE 名
MODULE = github.com/FantasyRL/go-mcp-demo
# 目录相关
DIR = $(shell pwd)
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
	@docker network inspect $(DOCKER_NET) >/dev/null 2>&1 || docker network create $(DOCKER_NET)

.PHONY: docker-run-%
docker-run-%: docker-build-% docker-net
	@echo ">> Running docker (STRICT config): $* on network $(DOCKER_NET)"
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

.PHONY: stdio
stdio:
	go build -o bin/mcp_server ./cmd/mcp_server # windows的output需要是.exe，并且在config.stdio.yaml中修改，bin/mcp-server.exe
	go run ./cmd/host -cfg $(CONFIG_PATH)/config.stdio.yaml