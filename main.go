package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/local/abyss-echo/internal/game"
)

var version = "dev"

func main() {
	var seed int64
	var showVersion bool
	flag.Int64Var(&seed, "seed", 0, "固定随机种子，便于复现同一局调查")
	flag.BoolVar(&showVersion, "version", false, "显示版本号")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "深渊回响 — 克苏鲁式终端生存调查\n\n用法: %s [选项]\n\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Parse()

	if showVersion {
		fmt.Printf("深渊回响 %s\n", version)
		return
	}
	if seed == 0 {
		seed = time.Now().UnixNano()
	}

	screen, err := tcell.NewScreen()
	if err != nil {
		fmt.Fprintf(os.Stderr, "无法初始化终端: %v\n", err)
		os.Exit(1)
	}
	app := game.NewApp(screen, seed)
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "运行失败: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("灯灭了。")
}
