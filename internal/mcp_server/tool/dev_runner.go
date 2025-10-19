package tool

import (
	"github.com/FantasyRL/go-mcp-demo/internal/mcp_server/internal/dev_runner"
	"github.com/FantasyRL/go-mcp-demo/pkg/base/tool_set"
	"github.com/mark3labs/mcp-go/mcp"
)

// WithDevRunnerTools 本地开发辅助工具
// 这组工具让 AI 能像本地助手一样：查看项目目录树(fs_tree)、读取文件(fs_cat)、运行项目/脚本(code_run)。
// - fs_tree：列出指定目录的树形结构（可控制深度/忽略模式），帮助 AI 感知项目布局。
// - fs_cat ：读取指定文件的内容（可限制最大字节），帮助 AI 查看未直接提供的代码。
// - code_run：在给定根目录下自动/按命令运行项目或单文件，返回 stdout/stderr/exit code，并给出基于错误输出的建议。
func WithDevRunnerTools() tool_set.Option {
	return func(toolSet *tool_set.ToolSet) {

		// fs_tree 目录树查看，让AI感知在哪个目录下运行代码
		toolTree := mcp.NewTool("fs_tree",
			mcp.WithDescription("List a directory as a plain text tree to understand project layout."),
			mcp.WithString("path", mcp.Required(), mcp.Description("Directory path to list")),
			// depth 最大遍历深度
			mcp.WithNumber("depth", mcp.Description("Max depth to traverse (default 4)")),
			// ignore 如 node_modules, *.log
			mcp.WithString("ignore", mcp.Description("Comma-separated glob patterns to ignore (optional)")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolTree)
		toolSet.HandlerFunc[toolTree.Name] = dev_runner.HandleFsTree

		// fs_cat 读取文件里的内容
		toolCat := mcp.NewTool("fs_cat",
			mcp.WithDescription("Read a file content to inspect code that was not provided in the prompt."),
			// 文件路径
			mcp.WithString("path", mcp.Required(), mcp.Description("File path to read")),
			// 最大读取字节数
			mcp.WithNumber("max_bytes", mcp.Description("Max bytes to read (default 65536)")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolCat)
		toolSet.HandlerFunc[toolCat.Name] = dev_runner.HandleFsCat

		// code_run 运行命令行
		toolRun := mcp.NewTool("code_run",
			// 工具用途：在本地命令行运行项目/脚本，返回 stdout/stderr/exit code，并基于错误输出给建议
			mcp.WithDescription("Run a code file/project locally in the given root directory with the EXACT command provided by the AI,return stdout"),
			// required ：工作目录（项目根目录）
			mcp.WithString("root", mcp.Required(), mcp.Description("Working directory of the project")),
			// 可选参数：运行前将 content 写入到 root 下的 file（相对路径）
			//mcp.WithString("file", mcp.Description("Optional file path relative to root to write/update before running")),
			//mcp.WithString("content", mcp.Description("Optional content to write to `file` before running")),
			// required ：显式运行命令
			mcp.WithString("command", mcp.Description("Explicit shell command to run under the root directory(eg `python main.py`,`go run cmd/host`,`npm run dev`)")),
			// optional ：超时（秒），默认 120s
			mcp.WithNumber("timeout_sec", mcp.Description("Timeout in seconds (default 120)")),
			// optional ：传给程序的标准输入
			mcp.WithString("stdin", mcp.Description("Optional STDIN to pass to the program")),
		)
		toolSet.Tools = append(toolSet.Tools, &toolRun)
		toolSet.HandlerFunc[toolRun.Name] = dev_runner.HandleCodeRun
	}
}
