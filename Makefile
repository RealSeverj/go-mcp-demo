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

.PHONY: stdio
stdio:
	go build -o bin/mcp_server ./cmd/mcp_server # windows的output需要是，并且在config.stdio.yaml中修改，bin/mcp-server.exe
	go run ./cmd/host -cfg $(CONFIG_PATH)/config.stdio.yaml