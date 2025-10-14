# go-mcp-demo
a demo to learn how to use mcp in go

# 项目结构
- 直接将hertz HTTP服务作为mcp-host
- mcp-host与mcp-server通过http/SSE通信
- mcp-host与ollama通过http通信

# quick start
- copy `config.example.yaml` to `config.yaml` (`config.stdio.yaml`同理)
- windows需要安装`makefile`相关工具
## stdio
```bash
make stdio # windows需要修改config.stdio.yaml中的mcp.stdio.server_cmd 为./bin/mcp-server.exe
```