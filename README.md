# 深渊回响 / Echoes of the Abyss

一个克苏鲁式终端生存调查游戏。三层程序化迷宫、有限灯光、理智崩坏、无法被杀死的追猎者，以及四个结局。

```
              ╭───────────╮
           ╭──╯           ╰──╮
          ╱       ◉         ╲
         │        ╱ ╲         │
          ╲      ╱   ╲       ╱
           ╰──╮ ╱     ╲ ╭──╯
              ╰╯       ╰╯

                 深 渊 回 响
        E C H O E S   O F   T H E   A B Y S S
```

## 快速开始

需要 Go 1.24 或更高版本。

```sh
# Termux
pkg install golang

# Linux / macOS / Windows (PowerShell)
go run .
```

固定种子可以复现同一张地图和同一局调查：

```sh
go run . -seed 42
```

## 操作

| 按键 | 作用 |
| --- | --- |
| `WASD` / 方向键 | 移动 |
| `空格` | 等待一回合 |
| `F` | 撒圣盐，生成 6 回合免伤盐圈 |
| `E` | 阅读一页残页，揭示阶梯方位 |
| `C` | 食用一块方糖，恢复 20 点理智 |
| `Enter` | 站在阶梯上时下潜 / 互动 |
| `?` / `H` | 调查手册 |
| `Q` | 放弃调查（游戏中会二次确认） |
| `R` | 结局画面重新调查 |
| `Esc` | 返回 / 关闭弹层 |
| `Ctrl+C` | 立即退出 |

## 生存规则

- 生命上限为 100。每层都藏着一枚刻印，会以 `◈` 穿过迷雾显示，目标栏同时给出方位和大概距离；集齐三枚后，第三层的 `>` 会通向潮汐之门。
- 每次移动消耗灯油。灯油越少，视野越小；耗尽后黑暗会持续侵蚀理智。
- 敌人无法被杀死。被追上会损失生命与理智；按 `F` 撒盐会生成持续 6 回合的盐圈，期间不会受伤，也能直接穿过怪物。
- 追猎者攻击后会停顿两回合。它们失去你的视线三回合后，也会放弃追踪。
- 地图上的方糖拾取后进入库存，不会自动食用。按 `C` 主动吃一块，恢复 20 点理智；满理智时不会消耗。
- 走投无路时，没有盐也可以强行撞过敌人，但会损失 20 生命和 8 理智。
- 残页拾取后不会自动使用。按 `E` 阅读，才会揭示阶梯方位，并损失 6 点理智。
- 低理智会让墙上出现不该存在的符号，也会让低语越来越频繁。
- 理智或生命归零都会结束调查。潮汐之门前还有最后一次选择。

## 构建

```sh
make build          # 当前平台 -> bin/abyss-echo
make test           # 逻辑与渲染测试
make dist           # 全平台交叉编译 -> dist/
```

也可以直接运行脚本：

```sh
./scripts/build-all.sh
```

产物：

```
abyss-echo_1.0.0_linux_amd64
abyss-echo_1.0.0_linux_arm64
abyss-echo_1.0.0_linux_arm
abyss-echo_1.0.0_darwin_amd64
abyss-echo_1.0.0_darwin_arm64
abyss-echo_1.0.0_windows_amd64.exe
abyss-echo_1.0.0_windows_arm64.exe
SHA256SUMS
```

构建使用 `CGO_ENABLED=0`，不需要目标平台的 C 工具链或系统库。Windows、macOS 和 Linux 的二进制均可在任意一台装有 Go 的机器上交叉编译。

- `linux_arm64` 产物可直接在 Termux 中运行。
- 从网络下载的 macOS 产物如果被 Gatekeeper 拦截，可对本机文件执行 `xattr -d com.apple.quarantine <文件>`。
- Windows 未签名产物可能触发 SmartScreen 提示；使用 Windows Terminal 运行。

## 终端要求

- 最小尺寸 `46 × 16`，推荐 `100 × 30` 或更大。
- 需要 UTF-8 终端和一款包含中日韩字形的等宽字体。
- Windows 建议使用 Windows Terminal；经典 `cmd.exe` 的字体与颜色支持可能不完整。
- 窄屏会自动隐藏侧栏，改为紧凑状态栏；调整窗口后界面会立即重排。

## 项目结构

```
main.go                 入口与命令行参数
internal/game/
  app.go                tcell 事件循环
  model.go              回合、生存、敌人、结局状态机
  level.go              地图生成、视野、寻路
  content.go            剧情、物品、结局文本
  render.go             自适应 TUI 渲染
  game_test.go          地图/机制/渲染快照测试
scripts/build-all.sh    七平台交叉编译与校验和
Makefile                常用开发命令
```

## 依赖

- [tcell](https://github.com/gdamore/tcell) — 跨平台终端屏幕与输入
- [uniseg](https://github.com/rivo/uniseg) — Unicode 字素宽度处理

如果访问官方 Go module proxy 超时，可以使用：

```sh
GOPROXY=https://goproxy.cn,direct go mod download
```
