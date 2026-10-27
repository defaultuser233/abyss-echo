package game

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/uniseg"
)

var (
	colBg       = tcell.NewHexColor(0x06090b)
	colPanel    = tcell.NewHexColor(0x0b1013)
	colPanelAlt = tcell.NewHexColor(0x0e1518)
	colMapBg    = tcell.NewHexColor(0x040709)
	colText     = tcell.NewHexColor(0xd5cdb4)
	colBone     = tcell.NewHexColor(0xeee6cf)
	colDim      = tcell.NewHexColor(0x66716e)
	colFaint    = tcell.NewHexColor(0x303b3c)
	colAmber    = tcell.NewHexColor(0xd6a855)
	colRed      = tcell.NewHexColor(0xb64b4b)
	colBlue     = tcell.NewHexColor(0x7899ba)
	colViolet   = tcell.NewHexColor(0x947fb2)
	colGreen    = tcell.NewHexColor(0x7f9b82)
	colWhisper  = tcell.NewHexColor(0xa8759b)
	colCyan     = tcell.NewHexColor(0x6fa3a3)

	styleBase         = tcell.StyleDefault.Background(colBg).Foreground(colText)
	styleDim          = styleBase.Foreground(colDim)
	styleFaint        = styleBase.Foreground(colFaint)
	styleBone         = styleBase.Foreground(colBone).Bold(true)
	styleAmber        = styleBase.Foreground(colAmber)
	styleRed          = styleBase.Foreground(colRed)
	styleBlue         = styleBase.Foreground(colBlue)
	styleViolet       = styleBase.Foreground(colViolet)
	styleGreen        = styleBase.Foreground(colGreen)
	styleWhisper      = styleBase.Foreground(colWhisper).Italic(true)
	styleCyan         = styleBase.Foreground(colCyan)
	stylePanel        = tcell.StyleDefault.Background(colPanel).Foreground(colText)
	stylePanelDim     = stylePanel.Foreground(colDim)
	stylePanelFaint   = stylePanel.Foreground(colFaint)
	styleModalBg      = tcell.StyleDefault.Background(colPanelAlt).Foreground(colText)
	styleModalDim     = styleModalBg.Foreground(colDim)
	styleModalFaint   = styleModalBg.Foreground(colFaint)
	styleModalBone    = styleModalBg.Foreground(colBone).Bold(true)
	styleModalAmber   = styleModalBg.Foreground(colAmber)
	styleModalRed     = styleModalBg.Foreground(colRed)
	styleModalWhisper = styleModalBg.Foreground(colWhisper).Italic(true)
	styleBorder       = styleBase.Foreground(colDim)
)

type brush struct {
	screen tcell.Screen
	w, h   int
}

func newBrush(screen tcell.Screen) *brush {
	w, h := screen.Size()
	return &brush{screen: screen, w: w, h: h}
}

func (a *App) draw() {
	if a.screen == nil {
		return
	}
	b := newBrush(a.screen)
	if b.w < 46 || b.h < 16 {
		a.drawTooSmall(b)
		b.screen.Show()
		return
	}
	b.fill(styleBase)

	m := a.model
	if m.modal != nil && m.modal.Kind == ModalHelp {
		a.drawHelp(b)
		b.screen.Show()
		return
	}

	switch m.phase {
	case PhaseTitle:
		a.drawTitle(b)
	case PhaseIntro:
		a.drawIntro(b)
	case PhasePlaying:
		a.drawGame(b)
	case PhaseEnding:
		a.drawEnding(b)
	}
	if m.modal != nil && m.modal.Kind != ModalHelp {
		a.drawModal(b)
	}
	b.screen.Show()
}

func (b *brush) fill(style tcell.Style) {
	b.screen.Fill(' ', style)
}

func (b *brush) put(x, y int, style tcell.Style, text string) int {
	if y < 0 || y >= b.h || text == "" {
		return x
	}
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		width := g.Width()
		if width < 0 {
			width = 0
		}
		if x >= b.w {
			return x
		}
		if x >= 0 && x+width <= b.w {
			b.screen.Put(x, y, g.Str(), style)
		}
		x += width
	}
	return x
}

func (b *brush) putClipped(x, y, maxWidth int, style tcell.Style, text string) int {
	if maxWidth <= 0 {
		return x
	}
	end := x + maxWidth
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		width := g.Width()
		if width < 0 {
			width = 0
		}
		if x+width > end {
			break
		}
		if x >= 0 && x+width <= b.w {
			b.screen.Put(x, y, g.Str(), style)
		}
		x += width
	}
	return x
}

func (b *brush) center(y int, style tcell.Style, text string) {
	width := uniseg.StringWidth(text)
	x := (b.w - width) / 2
	if x < 0 {
		x = 0
	}
	b.putClipped(x, y, b.w, style, text)
}

func (b *brush) right(y int, style tcell.Style, text string) {
	width := uniseg.StringWidth(text)
	x := b.w - 1 - width
	if x < 0 {
		x = 0
	}
	b.putClipped(x, y, b.w, style, text)
}

func (b *brush) panel(x, y, w, h int, border, fill tcell.Style, title string) {
	if w < 2 || h < 2 || x >= b.w || y >= b.h {
		return
	}
	for yy := y; yy < y+h && yy < b.h; yy++ {
		for xx := x; xx < x+w && xx < b.w; xx++ {
			if xx >= 0 && yy >= 0 {
				b.screen.Put(xx, yy, " ", fill)
			}
		}
	}
	right := x + w - 1
	bottom := y + h - 1
	b.put(x, y, border, "╭")
	b.put(right, y, border, "╮")
	b.put(x, bottom, border, "╰")
	b.put(right, bottom, border, "╯")
	for xx := x + 1; xx < right; xx++ {
		b.put(xx, y, border, "─")
		b.put(xx, bottom, border, "─")
	}
	for yy := y + 1; yy < bottom; yy++ {
		b.put(x, yy, border, "│")
		b.put(right, yy, border, "│")
	}
	if title != "" {
		label := truncateText(" "+title+" ", maxInt(0, w-4))
		if label != "" {
			b.put(x+2, y, border, label)
		}
	}
}

func (b *brush) panelDouble(x, y, w, h int, border, fill tcell.Style, title string) {
	if w < 2 || h < 2 || x >= b.w || y >= b.h {
		return
	}
	for yy := y; yy < y+h && yy < b.h; yy++ {
		for xx := x; xx < x+w && xx < b.w; xx++ {
			if xx >= 0 && yy >= 0 {
				b.screen.Put(xx, yy, " ", fill)
			}
		}
	}
	right := x + w - 1
	bottom := y + h - 1
	b.put(x, y, border, "╔")
	b.put(right, y, border, "╗")
	b.put(x, bottom, border, "╚")
	b.put(right, bottom, border, "╝")
	for xx := x + 1; xx < right; xx++ {
		b.put(xx, y, border, "═")
		b.put(xx, bottom, border, "═")
	}
	for yy := y + 1; yy < bottom; yy++ {
		b.put(x, yy, border, "║")
		b.put(right, yy, border, "║")
	}
	if title != "" {
		label := truncateText("  "+title+"  ", maxInt(0, w-4))
		if label != "" {
			b.put(x+2, y, border, label)
		}
	}
}

func (a *App) drawTooSmall(b *brush) {
	b.fill(styleBase)
	if b.w >= 20 && b.h >= 5 {
		b.center(b.h/2-2, styleBone, "终端太小")
		b.center(b.h/2, styleDim, "请调整到至少 46 × 16")
		b.center(b.h/2+1, styleDim, fmt.Sprintf("当前 %d × %d", b.w, b.h))
	}
}

func (a *App) drawTitle(b *brush) {
	m := a.model
	b.fill(styleBase)
	drawStars(b, m.seed, b.h-5)

	art := []string{
		"              ╭───────────╮",
		"           ╭──╯           ╰──╮",
		"          ╱       ◉         ╲",
		"         │        ╱ ╲         │",
		"          ╲      ╱   ╲       ╱",
		"           ╰──╮ ╱     ╲ ╭──╯",
		"              ╰╯       ╰╯",
	}
	showArt := b.w >= 52 && b.h >= 23
	contentHeight := 9
	if showArt {
		contentHeight += len(art) + 1
	}
	start := (b.h - contentHeight) / 2
	if start < 1 {
		start = 1
	}

	if showArt {
		for i, line := range art {
			b.center(start+i, styleDim, line)
		}
		start += len(art) + 1
	}

	b.center(start, styleBone, "深 渊 回 响")
	b.center(start+1, styleAmber, "E C H O E S   O F   T H E   A B Y S S")

	quote := titleQuotes[positiveMod(m.seed^0x5eed, int64(len(titleQuotes)))]
	for i, line := range wrapText(quote, minInt(58, b.w-8)) {
		b.center(start+3+i, styleWhisper, line)
	}

	menuY := start + 6
	if menuY < b.h-5 {
		if b.w >= 78 {
			b.center(menuY, styleBase, "[ Enter ] 踏入雾港      [ ? ] 调查手册      [ Q ] 离开")
		} else {
			b.center(menuY, styleBase, "[ Enter ] 踏入雾港")
			b.center(menuY+1, styleDim, "[ ? ] 调查手册    [ Q ] 离开")
		}
	}

	if b.h >= 22 {
		drawOcean(b, b.h-4, m.seed)
	}
	b.center(b.h-2, styleFaint, "1927 · 雾港 · 一处不该存在的海岸")
	if b.w > 72 {
		b.put(2, b.h-1, styleFaint, fmt.Sprintf("seed %d", m.seed))
	}
}

func (a *App) drawIntro(b *brush) {
	b.fill(styleBase)
	drawStars(b, a.model.seed+17, b.h/2)

	boxW := minInt(76, b.w-6)
	bodyW := boxW - 6
	type introLine struct {
		text  string
		amber bool
	}
	var content []introLine
	for _, paragraph := range introParagraphs {
		for _, line := range wrapText(paragraph, bodyW) {
			content = append(content, introLine{text: line, amber: strings.HasPrefix(paragraph, "不要让灯")})
		}
		content = append(content, introLine{})
	}
	maxH := b.h - 4
	if len(content)+5 > maxH {
		content = content[:0]
		for _, line := range introCompact {
			for _, wrapped := range wrapText(line, bodyW) {
				content = append(content, introLine{
					text:  wrapped,
					amber: strings.HasPrefix(line, "不要让灯") || strings.HasPrefix(line, "不要回应"),
				})
			}
		}
	}
	boxH := minInt(maxH, len(content)+5)
	x := (b.w - boxW) / 2
	y := (b.h - boxH) / 2
	b.panel(x, y, boxW, boxH, styleBorder, stylePanel, "调查简报 // 雾港，1927")

	bodyX := x + 3
	lineY := y + 2
	for _, line := range content {
		if lineY >= y+boxH-3 {
			break
		}
		style := stylePanel
		if line.amber {
			style = stylePanel.Foreground(colAmber)
		}
		b.putClipped(bodyX, lineY, bodyW, style, line.text)
		lineY++
	}

	footer := "Enter 踏入教堂      Esc 返回标题      ? 调查手册"
	if uniseg.StringWidth(footer) > bodyW {
		footer = "Enter 继续   Esc 返回   ? 手册"
	}
	b.center(y+boxH-2, stylePanelDim, footer)
}

func (a *App) drawGame(b *brush) {
	m := a.model
	if m.level == nil {
		return
	}
	b.fill(styleBase)
	footerY := b.h - 4
	mapY := 2
	mapH := footerY - mapY
	sidebar := b.w >= 88
	mapX := 1
	mapW := b.w - 2
	sideX, sideW := 0, 0
	if sidebar {
		sideW = 25
		sideX = b.w - sideW - 1
		mapW = sideX - mapX - 3
		if mapW < 44 {
			sidebar = false
			mapW = b.w - 2
		}
	}

	a.drawHeader(b, sidebar)
	floor := floorInfoFor(m.depth)
	title := fmt.Sprintf("%s层 · %s", chineseNumber(m.depth), floor.Name)
	a.drawMap(b, mapX, mapY, mapW, mapH, title)
	if sidebar {
		a.drawSidebar(b, sideX, mapY, sideW, mapH)
	}
	a.drawFooter(b, footerY, sidebar)
}

func (a *App) drawHeader(b *brush, sidebar bool) {
	m := a.model
	left := "◈ 深渊回响"
	b.put(1, 0, styleBone, left)
	x := 1 + uniseg.StringWidth(left)
	if b.w > 96 {
		b.put(x, 0, styleFaint, "  /  雾港调查记录")
	}

	floor := floorInfoFor(m.depth)
	right := fmt.Sprintf("第%s层 · %s", chineseNumber(m.depth), floor.Name)
	b.right(0, styleDim, right)

	dir := directionName(m.player, m.level.Stairs)
	sealLoc := m.sealLocator()
	gate := m.depth == TotalSeals
	status := ""
	switch {
	case m.player == m.level.Stairs && gate && m.seals == TotalSeals:
		status = "你站在潮汐之门前。按 Enter 做出选择。"
	case m.player == m.level.Stairs && gate:
		status = fmt.Sprintf("潮汐之门还缺刻印。%s：%s。", sealName(m.depth), sealLoc)
	case m.player == m.level.Stairs && m.level.SealTaken:
		status = "你站在阶梯上。按 Enter 下潜。"
	case m.player == m.level.Stairs:
		status = fmt.Sprintf("盐晶封住了阶梯。%s：%s。", sealName(m.depth), sealLoc)
	case gate && m.seals == TotalSeals:
		status = fmt.Sprintf("三枚刻印归位。潮汐之门在%s。", dir)
	case gate:
		status = fmt.Sprintf("潮汐之门在%s。梦之刻印：%s。", dir, sealLoc)
	case m.level.SealTaken:
		status = fmt.Sprintf("阶梯已开启，在%s。按 Enter 下潜。", dir)
	case m.level.StairsKnown:
		status = fmt.Sprintf("阶梯在%s，%s：%s。", dir, sealName(m.depth), sealLoc)
	default:
		status = fmt.Sprintf("寻找 ◈ %s（%s）。", sealName(m.depth), sealLoc)
	}
	b.putClipped(1, 1, b.w-2, styleGreen, status)
	if sidebar && b.w > 106 {
		b.right(1, styleFaint, fmt.Sprintf("步数 %d  ·  灯油 %d", m.steps, m.oil))
	}
}

func (a *App) drawMap(b *brush, x, y, w, h int, title string) {
	m := a.model
	l := m.level
	b.panel(x, y, w, h, depthBorder(m.depth), tcell.StyleDefault.Background(colMapBg).Foreground(colText), title)
	vw, vh := w-2, h-2
	if vw < 1 || vh < 1 {
		return
	}
	camX := cameraOrigin(m.player.X, vw, l.Width)
	camY := cameraOrigin(m.player.Y, vh, l.Height)
	ox, oy := x+1, y+1

	for sy := 0; sy < vh; sy++ {
		wy := camY + sy
		if wy < 0 || wy >= l.Height {
			continue
		}
		for sx := 0; sx < vw; sx++ {
			wx := camX + sx
			if wx < 0 || wx >= l.Width {
				continue
			}
			p := Pos{X: wx, Y: wy}
			i := l.index(p)
			gx := ox + sx
			gy := oy + sy
			visible := l.Visible[i]
			seen := l.Seen[i]

			style := depthFloorStyle(m.depth, visible)
			glyph := ""
			if visible {
				if l.tile(p) == TileWall {
					glyph = "▓"
					style = depthWallStyle(m.depth, true)
				} else {
					glyph = "·"
				}
			} else if seen {
				if l.tile(p) == TileWall {
					glyph = "▒"
					style = depthWallStyle(m.depth, false)
				}
			}

			if glyph != "" {
				b.put(gx, gy, style, glyph)
			}
			if p == l.Stairs && (visible || l.StairsKnown) {
				stairStyle := styleBase.Foreground(colAmber)
				if !visible {
					stairStyle = stairStyle.Dim(true)
				}
				b.put(gx, gy, stairStyle, ">")
			}

			if item := l.itemAt(p); item != nil {
				if item.Kind == ItemSeal {
					itemStyle := itemStyle(item.Kind)
					if !visible {
						itemStyle = itemStyle.Dim(true)
					}
					b.put(gx, gy, itemStyle, itemGlyph(item.Kind))
				} else if visible || seen {
					itemStyle := itemStyle(item.Kind)
					if !visible {
						itemStyle = itemStyle.Dim(true)
					}
					b.put(gx, gy, itemStyle, itemGlyph(item.Kind))
				}
			}

			if enemy := l.enemyAt(p); enemy != nil && visible {
				enemyStyle := enemyStyle(enemy.Kind)
				if enemy.Stun > 0 {
					enemyStyle = enemyStyle.Dim(true)
				}
				b.put(gx, gy, enemyStyle, enemyGlyph(enemy.Kind))
			}

			if m.san <= 18 && !visible && seen && glyph == "" && glitchAt(wx, wy, m.turn) {
				b.put(gx, gy, styleBase.Foreground(colWhisper).Dim(true), "?")
			}
		}
	}

	px := ox + m.player.X - camX
	py := oy + m.player.Y - camY
	if px >= ox && px < ox+vw && py >= oy && py < oy+vh {
		playerStyle := styleBone
		if m.san <= 25 {
			playerStyle = playerStyle.Foreground(colWhisper)
		}
		b.put(px, py, playerStyle, "@")
	}
}

func (a *App) drawSidebar(b *brush, x, y, w, h int) {
	m := a.model
	b.panel(x, y, w, h, styleBorder, stylePanel, "调查员")
	cx := x + 2
	cw := w - 4
	cy := y + 1

	barRow(b, cx, cy, cw, "生命", m.hp, StartHP, colRed)
	barRow(b, cx, cy+1, cw, "理智", maxInt(0, m.san), MaxSAN, colBlue)
	barRow(b, cx, cy+2, cw, "灯油", m.oil, MaxOil, colAmber)

	cy += 4
	sealLine := "刻印 "
	for i := 0; i < TotalSeals; i++ {
		if i < m.seals {
			sealLine += "◈"
		} else {
			sealLine += "◇"
		}
	}
	b.put(cx, cy, styleAmber, sealLine)
	b.put(cx+uniseg.StringWidth(sealLine)+3, cy, styleBase, fmt.Sprintf("圣盐 ✦ %d", m.salt))

	cy++
	b.put(cx, cy, stylePanelDim, fmt.Sprintf("步数 %-4d  残页 ? %d", m.steps, len(m.pages)))
	cy++
	b.put(cx, cy, stylePanelDim, fmt.Sprintf("方糖 ▫ %d     回合 %d", m.sugar, m.turn))
	if m.wardTurns > 0 {
		cy++
		b.put(cx, cy, styleAmber, fmt.Sprintf("盐圈 ●●●  %d 回合", m.wardTurns))
	}

	cy += 2
	b.put(cx, cy, stylePanelDim, "图 例")
	b.put(cx+6, cy, stylePanelFaint, strings.Repeat("─", maxInt(0, cw-6)))
	cy++
	legend := []struct {
		glyph string
		name  string
		style tcell.Style
	}{
		{"@", "你", styleBone},
		{"◈", "刻印", styleAmber},
		{">", "阶梯", styleAmber},
		{"◍", "煤油", styleAmber},
		{"▫", "方糖", styleViolet},
		{"✦", "圣盐", styleBone},
		{"?", "残页", styleCyan},
		{"☨", "低语者", styleWhisper},
		{"ψ", "深潜者", styleGreen},
		{"≈", "无面者", styleRed},
	}
	for i, entry := range legend {
		col := i % 2
		row := i / 2
		lx := cx + col*(cw/2)
		b.put(lx, cy+row, entry.style, entry.glyph)
		b.put(lx+2, cy+row, stylePanelDim, entry.name)
	}
	cy += 5
	cy++
	b.put(cx, cy, stylePanelDim, "操 作")
	b.put(cx+6, cy, stylePanelFaint, strings.Repeat("─", maxInt(0, cw-6)))
	cy++
	controls := []string{
		"WASD 移动   空格 等待",
		"F 撒盐    E 残页",
		"C 方糖    ? 手册",
		"Enter 阶梯  Q 离开",
	}
	for _, line := range controls {
		if cy >= y+h-1 {
			break
		}
		b.putClipped(cx, cy, cw, stylePanelDim, line)
		cy++
	}
}

func (a *App) drawFooter(b *brush, y int, sidebar bool) {
	m := a.model
	if sidebar {
		b.put(1, y, styleFaint, strings.Repeat("─", b.w-2))
		b.put(2, y, styleDim, " 低语记录 ")
		b.right(y, styleFaint, fmt.Sprintf("刻印 %d/%d ", m.seals, TotalSeals))
	} else {
		drawCompactStats(b, y, m)
		switch {
		case m.wardTurns > 0 && b.w >= 52:
			b.right(y, styleAmber, fmt.Sprintf("盐圈 %d", m.wardTurns))
		case len(m.pages) > 0 && m.sugar > 0 && b.w >= 72:
			b.right(y, styleFaint, "E 残页  C 方糖  ? 手册")
		case len(m.pages) > 0 && b.w >= 72:
			b.right(y, styleFaint, "E 阅读残页  ? 手册")
		case len(m.pages) > 0 && b.w >= 60:
			b.right(y, styleFaint, "E 阅读  ? 手册")
		case m.sugar > 0 && b.w >= 60:
			b.right(y, styleFaint, "C 方糖  ? 手册")
		case b.w >= 72:
			b.right(y, styleFaint, "WASD 移动  F 撒盐  ? 手册")
		case b.w >= 52:
			b.right(y, styleFaint, "? 手册")
		}
	}
	rows := b.h - y - 1
	if rows < 1 {
		return
	}
	drawLogs(b, m.messages, 2, y+1, b.w-4, rows)
}

func (a *App) drawModal(b *brush) {
	m := a.model
	if m.modal == nil {
		return
	}
	b.fill(styleBase)
	drawStars(b, m.seed+int64(m.turn)+19, b.h-4)
	bodyWidth := 0
	for _, line := range m.modal.Lines {
		bodyWidth = maxInt(bodyWidth, uniseg.StringWidth(line))
	}
	bodyWidth = minInt(68, maxInt(38, bodyWidth))
	boxW := minInt(b.w-4, bodyWidth+6)
	x := (b.w - boxW) / 2
	wrapWidth := boxW - 6

	var lines []string
	for _, line := range m.modal.Lines {
		if line == "" {
			lines = append(lines, "")
			continue
		}
		lines = append(lines, wrapText(line, wrapWidth)...)
	}
	maxLines := b.h - 8
	if maxLines < 1 {
		maxLines = 1
	}
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	boxH := len(lines) + 5
	if boxH > b.h-2 {
		boxH = b.h - 2
	}
	y := (b.h - boxH) / 2
	b.panelDouble(x, y, boxW, boxH, styleModalAmber.Dim(true), styleModalBg, m.modal.Title)

	lineY := y + 2
	for _, line := range lines {
		style := styleModalBg
		switch {
		case strings.HasPrefix(line, "1  "), strings.HasPrefix(line, "1 "):
			style = styleModalAmber
		case strings.HasPrefix(line, "2  "), strings.HasPrefix(line, "2 "):
			style = styleModalRed
		case strings.HasPrefix(line, "「"):
			style = styleModalWhisper
		}
		b.putClipped(x+3, lineY, wrapWidth, style, line)
		lineY++
	}
	if m.modal.Footer != "" {
		b.center(y+boxH-2, styleModalDim, m.modal.Footer)
	}
}

func (a *App) drawHelp(b *brush) {
	b.fill(styleBase)
	drawStars(b, a.model.seed+31, b.h-3)
	boxW := minInt(92, b.w-4)
	boxH := minInt(24, b.h-2)
	x := (b.w - boxW) / 2
	y := (b.h - boxH) / 2
	b.panelDouble(x, y, boxW, boxH, styleModalAmber.Dim(true), styleModalBg, "调查手册 // FIELD NOTES")

	innerX := x + 3
	innerW := boxW - 6
	cy := y + 2
	b.center(cy, styleModalBone, "你无法杀死它们。你只能活到天亮。")
	cy += 2

	if innerW >= 66 {
		leftW := 30
		gap := 4
		rightX := innerX + leftW + gap
		rightW := innerW - leftW - gap

		b.put(innerX, cy, styleModalAmber, "行 动")
		b.put(rightX, cy, styleModalAmber, "物品与符号")
		b.put(innerX+10, cy, styleModalFaint, strings.Repeat("─", maxInt(0, leftW-10)))
		b.put(rightX+12, cy, styleModalFaint, strings.Repeat("─", maxInt(0, rightW-12)))
		cy++

		controls := [][2]string{
			{"移动", "WASD / 方向键"},
			{"等待", "空格"},
			{"撒盐", "F（6 回合免伤）"},
			{"阅读", "E（消耗一页残页）"},
			{"食用", "C（消耗一块方糖）"},
			{"互动", "Enter（站在阶梯上）"},
			{"手册", "?"},
			{"离开", "Q（游戏中二次确认）"},
		}
		items := [][2]string{
			{"◈ 刻印", "穿雾显示；每层一枚，三枚开门"},
			{"◍ 煤油", "恢复照明；耗尽会损理智"},
			{"▫ 方糖", "拾取后按 C 食用：+20 理智"},
			{"✦ 圣盐", "F 撒盐：6 回合免伤"},
			{"? 残页", "E 阅读：揭示阶梯，-6 理智"},
		}
		rows := maxInt(len(controls), len(items))
		for i := 0; i < rows; i++ {
			if i < len(controls) {
				drawKV(b, innerX, cy+i, leftW, controls[i][0], controls[i][1], styleModalDim, styleModalBg)
			}
			if i < len(items) {
				drawKV(b, rightX, cy+i, rightW, items[i][0], items[i][1], styleModalAmber, styleModalDim)
			}
		}
		cy += rows + 1

		b.put(innerX, cy, styleModalAmber, "存 在")
		b.put(innerX+10, cy, styleModalFaint, strings.Repeat("─", maxInt(0, innerW-10)))
		cy++
		enemies := [][2]string{
			{"☨ 低语者", "缓慢，主要灼烧理智。"},
			{"ψ 深潜者", "紧追不舍，主要撕咬生命。"},
			{"≈ 无面者", "靠近时理智会迅速崩塌。"},
		}
		for i, enemy := range enemies {
			drawKV(b, innerX, cy+i, innerW, enemy[0], enemy[1], styleModalWhisper, styleModalDim)
		}
		cy += len(enemies) + 1
	} else {
		lines := []string{
			"移动 WASD/方向键   等待 空格",
			"撒盐 F 盐圈6回合   阅读 E 残页",
			"食用 C 方糖        手册 ?",
			"互动 Enter          离开 Q",
			"◈ 刻印 穿雾显示，集齐三枚开门",
			"◍ 煤油 照明 / ▫ 方糖 C食用+20",
			"✦ 圣盐 6回合免伤  ? 残页 E阅读",
			"☨ 低语者  ψ 深潜者  ≈ 无面者",
		}
		for _, line := range lines {
			b.putClipped(innerX, cy, innerW, styleModalDim, line)
			cy++
		}
		cy++
	}

	if cy < y+boxH-3 {
		b.put(innerX, cy, styleModalRed, "不要战斗。不要凝视。不要回应。")
		cy++
		b.putClipped(innerX, cy, innerW, styleModalDim, "灯油会耗尽，理智会崩塌；刻印不会等你。")
		cy++
		if cy < y+boxH-2 {
			b.putClipped(innerX, cy, innerW, styleModalDim, "残页拾取后不会自动使用：按 E 阅读，才会揭示阶梯并扣除理智。")
			cy++
		}
		if cy < y+boxH-2 {
			b.putClipped(innerX, cy, innerW, styleModalDim, "盐圈持续 6 回合：期间不会受伤，也能直接穿过怪物。")
		}
	}
	b.center(y+boxH-2, styleModalDim, a.model.modal.Footer)
}

func (a *App) drawEnding(b *brush) {
	b.fill(styleBase)
	m := a.model
	if m.ending == nil {
		return
	}
	drawStars(b, m.seed+71, b.h/2)
	compact := b.h < 24
	endingY, titleY, englishY, ruleY, bodyY := 1, 4, 5, 7, 9
	if compact {
		endingY, titleY, englishY, ruleY, bodyY = 0, 2, 3, 5, 7
	}
	b.center(endingY, styleDim, "终 局  //  E N D I N G")

	titleStyle := styleBone
	switch m.ending.Kind {
	case EndingSeal:
		titleStyle = styleAmber.Bold(true)
	case EndingEmbrace:
		titleStyle = styleViolet.Bold(true)
	case EndingDeath:
		titleStyle = styleRed.Bold(true)
	case EndingMadness:
		titleStyle = styleWhisper.Bold(true)
	}
	b.center(titleY, titleStyle, m.ending.Title)
	b.center(englishY, styleDim, m.ending.English)

	ruleW := minInt(34, b.w-10)
	b.put((b.w-ruleW)/2, ruleY, styleFaint, strings.Repeat("─", ruleW))

	bodyW := minInt(62, b.w-8)
	bodyX := (b.w - bodyW) / 2
	statsY := b.h - 4
	lineY := bodyY
	for _, paragraph := range m.ending.Body {
		for _, line := range wrapText(paragraph, bodyW) {
			if lineY >= statsY-1 {
				break
			}
			b.putClipped(bodyX, lineY, bodyW, tcell.StyleDefault.Background(colBg).Foreground(colText), line)
			lineY++
		}
		lineY++
	}

	stats := fmt.Sprintf("抵达 第%s层  ·  刻印 %d/%d  ·  步数 %d  ·  回合 %d",
		chineseNumber(m.depth), m.seals, TotalSeals, m.steps, m.turn)
	b.center(statsY, styleDim, stats)
	prompt := "R 重新调查    Esc 返回标题    Q 离开"
	if b.w < 58 {
		prompt = "R 重启   Esc 标题   Q 退出"
	}
	b.center(statsY+2, styleBone, prompt)
}

func drawStars(b *brush, seed int64, maxY int) {
	if maxY > b.h {
		maxY = b.h
	}
	seedBits := uint64(seed)
	for y := 0; y < maxY; y++ {
		for x := 0; x < b.w; x++ {
			h := uint64(x)*0x9e3779b1 ^ uint64(y)*0x85ebca6b ^ seedBits
			if h%937 < 3 {
				style := styleFaint
				if h%7 == 0 {
					style = styleDim
				}
				b.put(x, y, style, "·")
			}
		}
	}
}

func drawOcean(b *brush, y int, seed int64) {
	if y < 0 {
		return
	}
	for row := 0; row < 2 && y+row < b.h; row++ {
		for x := 0; x < b.w; x++ {
			v := (x*7 + row*19 + positiveMod(seed, 23)) % 29
			if v < 5 {
				style := styleBase.Foreground(colFaint)
				if row == 0 {
					style = styleBase.Foreground(colDim)
				}
				b.put(x, y+row, style, "~")
			}
		}
	}
}

func drawCompactStats(b *brush, y int, m *Model) {
	x := 2
	x = b.putClipped(x, y, b.w-4, styleRed, fmt.Sprintf("生命 %d", maxInt(0, m.hp)))
	x = b.putClipped(x, y, b.w-4-(x-2), styleDim, "  ")
	x = b.putClipped(x, y, b.w-4-(x-2), styleBlue, fmt.Sprintf("理智 %d", maxInt(0, m.san)))
	x = b.putClipped(x, y, b.w-4-(x-2), styleDim, "  ")
	x = b.putClipped(x, y, b.w-4-(x-2), styleAmber, fmt.Sprintf("灯油 %d", m.oil))
	x = b.putClipped(x, y, b.w-4-(x-2), styleDim, "  ")
	x = b.putClipped(x, y, b.w-4-(x-2), styleBone, fmt.Sprintf("盐 %d", m.salt))
	x = b.putClipped(x, y, b.w-4-(x-2), styleAmber, fmt.Sprintf("  刻印 %d/%d", m.seals, TotalSeals))
}

type logLine struct {
	text string
	tone Tone
}

func drawLogs(b *brush, messages []Message, x, y, width, rows int) {
	if width < 6 || rows < 1 {
		return
	}
	lines := make([]logLine, 0, rows)
	for i := len(messages) - 1; i >= 0 && len(lines) < rows; i-- {
		wrapped := wrapText(messages[i].Text, width-2)
		for j := len(wrapped) - 1; j >= 0 && len(lines) < rows; j-- {
			if wrapped[j] == "" {
				continue
			}
			lines = append(lines, logLine{text: wrapped[j], tone: messages[i].Tone})
		}
	}
	for i, j := 0, len(lines)-1; i < j; i, j = i+1, j-1 {
		lines[i], lines[j] = lines[j], lines[i]
	}
	for i, line := range lines {
		b.put(x, y+i, styleFaint, "› ")
		b.putClipped(x+2, y+i, width-2, toneStyle(line.tone), line.text)
	}
}

func toneStyle(tone Tone) tcell.Style {
	switch tone {
	case ToneGood:
		return styleGreen
	case ToneBad:
		return styleRed
	case ToneWhisper:
		return styleWhisper
	case ToneMystic:
		return styleCyan
	default:
		return styleDim
	}
}

func barRow(b *brush, x, y, width int, label string, value, maxValue int, color tcell.Color) {
	if width < 16 {
		return
	}
	b.put(x, y, stylePanelDim, label)
	barX := x + 5
	barW := width - 11
	if barW > 14 {
		barW = 14
	}
	if barW < 3 {
		barW = 3
	}
	filled := 0
	if maxValue > 0 {
		filled = value * barW / maxValue
	}
	if filled < 0 {
		filled = 0
	}
	if filled > barW {
		filled = barW
	}
	for i := 0; i < barW; i++ {
		if i < filled {
			b.put(barX+i, y, stylePanel.Foreground(color), "█")
		} else {
			b.put(barX+i, y, stylePanelFaint, "░")
		}
	}
	b.put(barX+barW+1, y, stylePanel, fmt.Sprintf("%3d", maxInt(0, value)))
}

func drawKV(b *brush, x, y, width int, key, value string, keyStyle, valueStyle tcell.Style) {
	keyW := 10
	if width < keyW+6 {
		keyW = minInt(width/2, 8)
	}
	b.putClipped(x, y, keyW, keyStyle, key)
	b.putClipped(x+keyW, y, maxInt(0, width-keyW-1), valueStyle, value)
}

func truncateText(text string, width int) string {
	if width <= 0 {
		return ""
	}
	if uniseg.StringWidth(text) <= width {
		return text
	}
	var builder strings.Builder
	used := 0
	g := uniseg.NewGraphemes(text)
	for g.Next() {
		w := g.Width()
		if used+w > width {
			break
		}
		builder.WriteString(g.Str())
		used += w
	}
	return builder.String()
}

func wrapText(text string, width int) []string {
	if width < 1 {
		return []string{""}
	}
	paragraphs := strings.Split(text, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		var builder strings.Builder
		lineWidth := 0
		g := uniseg.NewGraphemes(paragraph)
		hasContent := false
		for g.Next() {
			cluster := g.Str()
			w := g.Width()
			if w < 0 {
				w = 0
			}
			if lineWidth+w > width && hasContent {
				lines = append(lines, builder.String())
				builder.Reset()
				lineWidth = 0
				hasContent = false
				if cluster == " " || cluster == "\t" {
					continue
				}
			}
			builder.WriteString(cluster)
			lineWidth += w
			hasContent = true
		}
		if hasContent || builder.Len() > 0 {
			lines = append(lines, builder.String())
		}
	}
	if len(lines) == 0 {
		lines = append(lines, "")
	}
	return lines
}

func itemGlyph(kind ItemKind) string {
	switch kind {
	case ItemOil:
		return "◍"
	case ItemSugar:
		return "▫"
	case ItemSalt:
		return "✦"
	case ItemLore:
		return "?"
	case ItemSeal:
		return "◈"
	default:
		return "*"
	}
}

func itemStyle(kind ItemKind) tcell.Style {
	switch kind {
	case ItemOil:
		return styleAmber
	case ItemSugar:
		return styleViolet
	case ItemSalt:
		return styleBone
	case ItemLore:
		return styleCyan
	case ItemSeal:
		return styleAmber.Bold(true)
	default:
		return styleBase
	}
}

func enemyGlyph(kind EnemyKind) string {
	switch kind {
	case EnemyWhisperer:
		return "☨"
	case EnemyDeepOne:
		return "ψ"
	default:
		return "≈"
	}
}

func enemyStyle(kind EnemyKind) tcell.Style {
	switch kind {
	case EnemyWhisperer:
		return styleWhisper
	case EnemyDeepOne:
		return styleGreen
	default:
		return styleRed
	}
}

func cameraOrigin(player, view, world int) int {
	if world <= view {
		return -((view - world) / 2)
	}
	origin := player - view/2
	if origin < 0 {
		return 0
	}
	if origin+view > world {
		return world - view
	}
	return origin
}

func depthWallStyle(depth int, visible bool) tcell.Style {
	var color tcell.Color
	switch depth {
	case 1:
		color = tcell.NewHexColor(0x52646d)
	case 2:
		color = tcell.NewHexColor(0x536353)
	default:
		color = tcell.NewHexColor(0x5b4d6a)
	}
	style := tcell.StyleDefault.Background(colMapBg).Foreground(color)
	if !visible {
		style = style.Dim(true)
	}
	return style
}

func depthFloorStyle(depth int, visible bool) tcell.Style {
	var color tcell.Color
	switch depth {
	case 1:
		color = tcell.NewHexColor(0x727d75)
	case 2:
		color = tcell.NewHexColor(0x687760)
	default:
		color = tcell.NewHexColor(0x71647e)
	}
	style := tcell.StyleDefault.Background(colMapBg).Foreground(color)
	if !visible {
		style = style.Dim(true)
	}
	return style
}

func depthBorder(depth int) tcell.Style {
	color := colFaint
	switch depth {
	case 2:
		color = colGreen
	case 3:
		color = colViolet
	}
	return tcell.StyleDefault.Background(colBg).Foreground(color).Dim(true)
}

func glitchAt(x, y, turn int) bool {
	h := x*73 + y*131 + turn*17
	if h < 0 {
		h = -h
	}
	return h%79 == 0
}

func positiveMod(v, n int64) int {
	if n <= 0 {
		return 0
	}
	r := v % n
	if r < 0 {
		r += n
	}
	return int(r)
}
