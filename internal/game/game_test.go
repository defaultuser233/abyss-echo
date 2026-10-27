package game

import (
	"math/rand"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
)

func TestGeneratedLevelsAreReachable(t *testing.T) {
	for depth := 1; depth <= 3; depth++ {
		for seed := int64(1); seed <= 30; seed++ {
			level := GenerateLevel(depth, rand.New(rand.NewSource(seed)))
			dist := level.distanceField(level.Start)
			if dist[level.index(level.Stairs)] < 0 {
				t.Fatalf("depth %d seed %d: stairs unreachable", depth, seed)
			}
			positions := map[Pos]bool{}
			seals := 0
			for _, item := range level.Items {
				if !level.walkable(item.Pos) {
					t.Fatalf("depth %d seed %d: item on wall at %+v", depth, seed, item.Pos)
				}
				if dist[level.index(item.Pos)] < 0 {
					t.Fatalf("depth %d seed %d: item unreachable at %+v", depth, seed, item.Pos)
				}
				if positions[item.Pos] {
					t.Fatalf("depth %d seed %d: duplicate item at %+v", depth, seed, item.Pos)
				}
				positions[item.Pos] = true
				if item.Kind == ItemSeal {
					seals++
				}
			}
			if seals != 1 {
				t.Fatalf("depth %d seed %d: got %d seals, want 1", depth, seed, seals)
			}
			for _, enemy := range level.Enemies {
				if !level.walkable(enemy.Pos) {
					t.Fatalf("depth %d seed %d: enemy on wall at %+v", depth, seed, enemy.Pos)
				}
				if dist[level.index(enemy.Pos)] < 0 {
					t.Fatalf("depth %d seed %d: enemy unreachable at %+v", depth, seed, enemy.Pos)
				}
			}
		}
	}
}

func TestGenerationIsDeterministic(t *testing.T) {
	a := GenerateLevel(2, rand.New(rand.NewSource(99)))
	b := GenerateLevel(2, rand.New(rand.NewSource(99)))
	if a.Start != b.Start || a.Stairs != b.Stairs {
		t.Fatal("start or stairs differ for the same seed")
	}
	for i := range a.Tiles {
		if a.Tiles[i] != b.Tiles[i] {
			t.Fatalf("tile %d differs for the same seed", i)
		}
	}
	if len(a.Items) != len(b.Items) || len(a.Enemies) != len(b.Enemies) {
		t.Fatal("entity counts differ for the same seed")
	}
	for i := range a.Items {
		if a.Items[i].Kind != b.Items[i].Kind || a.Items[i].Pos != b.Items[i].Pos {
			t.Fatalf("item %d differs for the same seed", i)
		}
	}
	for i := range a.Enemies {
		if a.Enemies[i].Kind != b.Enemies[i].Kind || a.Enemies[i].Pos != b.Enemies[i].Pos {
			t.Fatalf("enemy %d differs for the same seed", i)
		}
	}
}

func TestFieldOfViewBlockedByWall(t *testing.T) {
	level := newLevel(1)
	for _, x := range []int{1, 2, 4, 5, 6} {
		level.carve(Pos{X: x, Y: 2})
	}
	origin := Pos{X: 1, Y: 2}
	wall := Pos{X: 3, Y: 2}
	target := Pos{X: 5, Y: 2}
	level.ComputeFOV(origin, 8)

	if !level.Visible[level.index(origin)] {
		t.Fatal("origin should be visible")
	}
	if !level.Visible[level.index(wall)] {
		t.Fatal("the blocking wall itself should be visible")
	}
	if level.Visible[level.index(target)] {
		t.Fatal("target behind a wall should not be visible")
	}
}

func TestDirectionNames(t *testing.T) {
	from := Pos{X: 5, Y: 5}
	cases := []struct {
		to   Pos
		want string
	}{
		{Pos{X: 5, Y: 5}, "脚下"},
		{Pos{X: 5, Y: 1}, "北方"},
		{Pos{X: 9, Y: 5}, "东方"},
		{Pos{X: 9, Y: 1}, "东北方"},
		{Pos{X: 1, Y: 9}, "西南方"},
	}
	for _, tc := range cases {
		if got := directionName(from, tc.to); got != tc.want {
			t.Errorf("directionName(%+v) = %q, want %q", tc.to, got, tc.want)
		}
	}
}

func TestSealUnlocksStairsAndDescends(t *testing.T) {
	m := NewModel(7)
	m.startRun()
	m.modal = nil
	m.player = m.level.Stairs
	m.level.SealTaken = false

	m.interact()
	if m.depth != 1 {
		t.Fatal("descended without the floor seal")
	}

	m.level.SealTaken = true
	m.interact()
	if m.depth != 2 {
		t.Fatalf("depth = %d, want 2", m.depth)
	}
	if m.phase != PhasePlaying {
		t.Fatalf("phase = %v, want playing", m.phase)
	}
	if m.modal == nil || m.modal.Kind != ModalFloor {
		t.Fatal("expected a floor introduction after descending")
	}
}

func TestLorePageRequiresReading(t *testing.T) {
	m := NewModel(9)
	m.startRun()
	m.modal = nil
	oldSAN := m.san
	m.pickUp(&Item{Kind: ItemLore, Pos: m.player})

	if len(m.pages) != 1 {
		t.Fatalf("pages = %d, want 1", len(m.pages))
	}
	if m.level.StairsKnown {
		t.Fatal("picking up a page should not reveal the stairs automatically")
	}
	if m.modal != nil {
		t.Fatal("picking up a page should not open the lore modal")
	}
	if m.san != oldSAN {
		t.Fatal("picking up a page should not cost sanity")
	}

	m.Handle(ActionReadPage)
	if len(m.pages) != 0 {
		t.Fatalf("pages = %d, want 0 after reading", len(m.pages))
	}
	if !m.level.StairsKnown {
		t.Fatal("reading a page should reveal the stairs")
	}
	if m.modal == nil || m.modal.Kind != ModalLore {
		t.Fatal("reading a page should open the lore modal")
	}
	if m.san != oldSAN-6 {
		t.Fatalf("san = %d, want %d", m.san, oldSAN-6)
	}
}

func TestThreeFloorSealProgression(t *testing.T) {
	m := NewModel(31)
	m.startRun()
	m.modal = nil

	for floor := 1; floor <= TotalSeals; floor++ {
		sealPos, ok := m.level.sealPosition()
		if !ok {
			t.Fatalf("floor %d has no seal", floor)
		}
		seal := &Item{Kind: ItemSeal, Pos: sealPos}
		m.pickUp(seal)
		if m.seals != floor {
			t.Fatalf("floor %d: seals = %d", floor, m.seals)
		}
		m.player = m.level.Stairs
		m.interact()

		if floor < TotalSeals {
			if m.depth != floor+1 {
				t.Fatalf("after floor %d: depth = %d", floor, m.depth)
			}
			m.modal = nil
			continue
		}
		if m.modal == nil || m.modal.Kind != ModalGate {
			t.Fatal("three seals should open the gate choice")
		}
	}
}

func TestSugarInventoryAndUse(t *testing.T) {
	m := NewModel(41)
	m.startRun()
	m.modal = nil
	m.level.Items = []*Item{{Kind: ItemSugar, Pos: m.player}}
	m.san = MaxSAN

	m.checkPickups()
	if m.sugar != 1 || len(m.level.Items) != 0 {
		t.Fatal("sugar should be stored in the inventory when picked up")
	}
	if m.san != MaxSAN {
		t.Fatalf("picking up sugar changed sanity to %d", m.san)
	}

	m.Handle(ActionEatSugar)
	if m.sugar != 1 {
		t.Fatal("sugar should not be consumed at full sanity")
	}

	m.san = 90
	m.Handle(ActionEatSugar)
	if m.sugar != 0 {
		t.Fatalf("sugar = %d, want 0 after eating", m.sugar)
	}
	if m.san != MaxSAN {
		t.Fatalf("san = %d, want %d after eating sugar", m.san, MaxSAN)
	}
}

func TestSaltStunsNearbyEnemy(t *testing.T) {
	m := NewModel(11)
	m.startRun()
	m.modal = nil
	m.salt = 1

	var neighbor Pos
	found := false
	for _, d := range cardinal {
		p := Pos{X: m.player.X + d.X, Y: m.player.Y + d.Y}
		if m.level.walkable(p) {
			neighbor = p
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no walkable neighbor at the starting position")
	}
	enemy := &Enemy{Kind: EnemyDeepOne, Pos: neighbor}
	m.level.Enemies = []*Enemy{enemy}
	m.useSalt()

	if m.salt != 0 {
		t.Fatalf("salt = %d, want 0", m.salt)
	}
	if enemy.Stun <= 0 {
		t.Fatal("nearby enemy was not stunned")
	}
	if m.wardTurns <= 0 {
		t.Fatal("salt ward was not activated")
	}
}

func TestSaltWardPreventsSqueezeDamage(t *testing.T) {
	m := NewModel(21)
	m.startRun()
	m.modal = nil
	m.level.Items = nil
	m.salt = 1

	var neighbor Pos
	var dx, dy int
	found := false
	for _, d := range cardinal {
		p := Pos{X: m.player.X + d.X, Y: m.player.Y + d.Y}
		if m.level.walkable(p) {
			neighbor = p
			dx, dy = d.X, d.Y
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no walkable neighbor at the starting position")
	}
	enemy := &Enemy{Kind: EnemyDeepOne, Pos: neighbor}
	m.level.Enemies = []*Enemy{enemy}
	m.useSalt()
	oldHP := m.hp
	m.tryMove(dx, dy)

	if m.player != neighbor {
		t.Fatalf("player = %+v, want %+v", m.player, neighbor)
	}
	if m.hp != oldHP {
		t.Fatalf("salt ward did not prevent squeeze damage: hp = %d, want %d", m.hp, oldHP)
	}

	enemy.Stun = 0
	enemy.Cooldown = 0
	m.advanceTurn()
	if m.hp != oldHP {
		t.Fatalf("salt ward did not prevent an enemy attack: hp = %d, want %d", m.hp, oldHP)
	}
}

func TestEnemyAttackCreatesEscapeWindow(t *testing.T) {
	m := NewModel(51)
	m.startRun()
	m.modal = nil
	m.level.Items = nil

	var neighbor Pos
	for _, d := range cardinal {
		p := Pos{X: m.player.X + d.X, Y: m.player.Y + d.Y}
		if m.level.walkable(p) {
			neighbor = p
			break
		}
	}
	enemy := &Enemy{Kind: EnemyDeepOne, Pos: neighbor}
	m.level.Enemies = []*Enemy{enemy}
	oldHP := m.hp
	m.advanceTurn()

	if m.hp >= oldHP {
		t.Fatal("enemy attack did not damage the player")
	}
	if enemy.Stun < 2 {
		t.Fatalf("enemy stun = %d, want at least 2", enemy.Stun)
	}
	if enemy.Alert {
		t.Fatal("enemy should lose alert after attacking")
	}
}

func TestEnemyLosesTrackWithoutSight(t *testing.T) {
	m := NewModel(52)
	m.startRun()
	m.modal = nil
	m.level.Items = nil

	var far Pos
	found := false
	for y := 0; y < m.level.Height && !found; y++ {
		for x := 0; x < m.level.Width; x++ {
			p := Pos{X: x, Y: y}
			if !m.level.walkable(p) || m.level.Visible[m.level.index(p)] {
				continue
			}
			if manhattan(p, m.player) >= 5 {
				far = p
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatal("no distant hidden cell found")
	}
	enemy := &Enemy{Kind: EnemyDeepOne, Pos: far, Alert: true, LostSight: 2}
	m.level.Enemies = []*Enemy{enemy}
	m.advanceTurn()

	if enemy.Alert {
		t.Fatal("enemy should lose the player after three turns without sight")
	}
}

func TestSqueezePastEnemyCostsHealth(t *testing.T) {
	m := NewModel(12)
	m.startRun()
	m.modal = nil
	m.level.Items = nil

	var neighbor Pos
	var dx, dy int
	found := false
	for _, d := range cardinal {
		p := Pos{X: m.player.X + d.X, Y: m.player.Y + d.Y}
		if m.level.walkable(p) {
			neighbor = p
			dx, dy = d.X, d.Y
			found = true
			break
		}
	}
	if !found {
		t.Fatal("no walkable neighbor at the starting position")
	}
	enemy := &Enemy{Kind: EnemyDeepOne, Pos: neighbor}
	m.level.Enemies = []*Enemy{enemy}
	oldHP := m.hp
	oldPlayer := m.player
	m.tryMove(dx, dy)

	if m.player != neighbor {
		t.Fatalf("player = %+v, want %+v", m.player, neighbor)
	}
	if enemy.Pos != oldPlayer {
		t.Fatalf("enemy = %+v, want %+v", enemy.Pos, oldPlayer)
	}
	if m.hp != oldHP-SqueezeHPCost {
		t.Fatalf("hp = %d, want %d", m.hp, oldHP-SqueezeHPCost)
	}
}

func TestGateChoicesProduceEndings(t *testing.T) {
	m := NewModel(13)
	m.startRun()
	m.modal = nil
	m.depth = TotalSeals
	m.seals = TotalSeals
	m.player = m.level.Stairs

	m.interact()
	if m.modal == nil || m.modal.Kind != ModalGate {
		t.Fatal("expected the gate choice")
	}
	m.Handle(ActionChoose1)
	if m.phase != PhaseEnding || m.ending == nil || m.ending.Kind != EndingSeal {
		t.Fatal("gate choice 1 did not produce the seal ending")
	}

	m.startRun()
	m.modal = nil
	m.depth = TotalSeals
	m.seals = TotalSeals
	m.player = m.level.Stairs
	m.interact()
	m.Handle(ActionChoose2)
	if m.phase != PhaseEnding || m.ending == nil || m.ending.Kind != EndingEmbrace {
		t.Fatal("gate choice 2 did not produce the embrace ending")
	}
}

func TestSimulationRenderFlow(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(100, 30)

	app := NewApp(screen, 42)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "深 渊 回 响") {
		t.Fatalf("title screen missing title:\n%s", got)
	}

	app.model.Handle(ActionConfirm)
	app.model.Handle(ActionConfirm)
	app.model.Handle(ActionConfirm)
	app.draw()
	text := screenText(screen)
	if !strings.Contains(text, "调查员") || !strings.Contains(text, "低语记录") {
		t.Fatalf("game screen missing HUD:\n%s", text)
	}
	if !strings.Contains(text, "潮汐教堂") {
		t.Fatalf("game screen missing floor name:\n%s", text)
	}

	app.model.Handle(ActionHelp)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "FIELD NOTES") {
		t.Fatalf("help screen missing heading:\n%s", got)
	}

	app.model.modal = nil
	app.model.finish(EndingDeath)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "退 潮 之 后") {
		t.Fatalf("ending screen missing title:\n%s", got)
	}
}

func TestNarrowLayoutRenders(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(60, 20)

	app := NewApp(screen, 5)
	app.draw()
	for i := 0; i < 3; i++ {
		app.model.Handle(ActionConfirm)
	}
	app.draw()
	text := screenText(screen)
	if !strings.Contains(text, "生命") || !strings.Contains(text, "刻印") {
		t.Fatalf("compact HUD missing on narrow screen:\n%s", text)
	}
}

func TestVisibleStairsRenderAsMarker(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(46, 16)

	app := NewApp(screen, 8)
	app.model.startRun()
	app.model.modal = nil
	stairs := app.model.level.Stairs
	for _, d := range cardinal {
		p := Pos{X: stairs.X + d.X, Y: stairs.Y + d.Y}
		if app.model.level.walkable(p) {
			app.model.player = p
			break
		}
	}
	app.model.level.StairsKnown = false
	app.model.level.ComputeFOV(app.model.player, app.model.lightRadius())
	app.draw()
	if got := screenText(screen); !strings.Contains(got, ">") {
		t.Fatalf("visible stairs marker missing:\n%s", got)
	}
}

func TestMinimumSizeAndTooSmall(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		t.Fatalf("init simulation screen: %v", err)
	}
	defer screen.Fini()
	screen.SetSize(46, 16)

	app := NewApp(screen, 3)
	app.draw()
	app.model.Handle(ActionConfirm)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "不要让灯熄灭") {
		t.Fatalf("compact intro missing warning:\n%s", got)
	}
	app.model.Handle(ActionConfirm)
	app.model.Handle(ActionConfirm)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "生命") {
		t.Fatalf("minimum-size HUD missing:\n%s", got)
	}

	app.model.finish(EndingDeath)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "退 潮 之 后") {
		t.Fatalf("minimum-size ending missing:\n%s", got)
	}

	screen.SetSize(40, 10)
	app.draw()
	if got := screenText(screen); !strings.Contains(got, "终端太小") {
		t.Fatalf("small terminal warning missing:\n%s", got)
	}
}

func screenText(screen tcell.Screen) string {
	w, h := screen.Size()
	var builder strings.Builder
	for y := 0; y < h; y++ {
		for x := 0; x < w; {
			text, _, width := screen.Get(x, y)
			if width < 1 {
				width = 1
			}
			builder.WriteString(text)
			x += width
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}
