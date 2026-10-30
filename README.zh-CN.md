# codex-sess

用 Go 编写的 Codex 会话管理 TUI。在保存会话的机器上运行，可列出所有已索引会话，搜索标题 / UUID / 项目路径，归档、恢复以及永久删除会话。

默认读取 `$CODEX_HOME`（未设置时为 `~/.codex`）中版本号最高的 `state_*.sqlite`，包括已归档记录。SQLite 以只读模式打开；状态变更通过 `codex archive`、`codex unarchive`、`codex delete` 完成。程序不直接修改 Codex 数据库或移动会话文件。

## 构建与运行

需要 Go 1.26 或更新版本和支持 archive / unarchive / delete 的 Codex CLI。已验证 codex-cli 0.153.4。

```bash
go build -buildvcs=false -o bin/codex-sess .
./bin/codex-sess
```

内置纯 Go SQLite 驱动，不依赖 sqlite3 命令或 CGO。构建时需要下载 Go 模块，构建后的程序浏览列表无需联网；变更操作还需要 Codex CLI。

```bash
# 在保存会话的远端 Linux 用户下运行
./bin/codex-sess --home "$HOME/.codex"

# 指定 Codex CLI
./bin/codex-sess --codex /usr/local/bin/codex

# 直接输出全部会话（含已归档）
./bin/codex-sess --list

# 只列出已归档，并搜索关键词
./bin/codex-sess --list --status archived --search project

# 输出 JSON，可供脚本处理
./bin/codex-sess --json

# 启动时只显示未归档会话
./bin/codex-sess --status active
```

`active` 在这里仅指未归档，不代表会话正在运行。列表覆盖指定 Codex 数据目录中数据库已索引的会话，不会自动汇总其他用户、SSH 主机或云端会话。没有被索引的历史 JSONL 文件不会出现在列表中。

## 键盘操作

| 按键 | 操作 |
| --- | --- |
| ↑ / ↓ 或 j / k | 选择会话 |
| PageUp / PageDown | 翻页 |
| / | 搜索标题、UUID 或目录 |
| Enter / Esc（搜索框中） | 回到会话列表 |
| Tab（列表中） | 全部 → 未归档 → 已归档 |
| a | 归档选中会话（确认后执行） |
| u | 恢复选中会话（确认后执行） |
| d | 永久删除选中会话，弹窗选择“Delete permanently”确认 |
| r | 刷新 |
| q / Ctrl+C | 退出 |

支持鼠标选择。界面仅显示会话列表，标题最多处理前 100 个字符、路径前 240 个字符；上下移动不会解析完整标题或生成详情。操作期间界面仍可绘制，仍可浏览、搜索、切换筛选和退出；仅阻止重复提交变更操作。命令默认 30 秒超时；发生超时或错误时会尝试重新读取状态，请确认最终结果后再重试。

删除不可撤销。工具确认 UUID 后才向 `codex delete UUID --force` 发出调用，不会通过 shell 拼接命令。归档或删除正在运行的会话可能影响其运行，请在操作前确认选中的标题与 UUID。

## 数据格式兼容性

`state_*.sqlite` 是 Codex 内部数据库格式，不是稳定公共 API。本工具使用当前已验证的 `threads` 字段（id、title、cwd、archived、updated_at）。未来版本发生不兼容结构变化时会报告读取错误，不尝试修复数据库。

自定义 `--home` 同时用于读取数据库和子进程的 `CODEX_HOME`，避免查看一个目录却操作另一个目录。请使用与桌面端相同的用户和 Codex 数据目录。在 SSH 远端直接运行本工具即可管理该远端的数据。

## 验证

```bash
go test ./...
go vet ./...
```

测试使用临时 SQLite 数据库、模拟 Codex 可执行文件以及 tcell 模拟终端，覆盖只读保护、筛选、版本选择、参数与数据目录传递、命令失败，以及删除确认界面、操作期间的导航和按需生成表格。测试不操作真实会话。

## 开发与项目结构

默认界面和命令提示使用英文。英文文档见 [README.md](README.md)。

```sh
make fmt
make check
make bench
make build VERSION=0.1.8
```

入口、命令行输出、TUI、显示辅助、会话模型、数据库访问和 Codex 命令调用分别位于 `main.go`、`cli.go`、`ui.go`、`display.go`、`session.go`、`store.go` 和 `runner.go`。GitHub Actions 使用 Go 1.26.x 执行格式检查、测试、静态检查和无 CGO 构建。

模块路径为 `github.com/sugoncloud/codex-session`。如需安装名为 `codex-sess` 的可执行文件，请克隆仓库后执行 `make install`，或下载 Release 压缩包。模块名称与可执行文件名称有意不同。

## Tag 自动发布

推送版本 tag 会触发 `.github/workflows/release.yml`：

```sh
git tag -a v0.1.8 -m "Release v0.1.8"
git push origin v0.1.8
```

GoReleaser 会为 Linux、macOS 和 Windows 的 amd64、arm64 架构生成压缩包，并发布校验和。工作流使用 GitHub 自动提供的 `GITHUB_TOKEN`。
