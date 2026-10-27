package game

import (
	"fmt"
	"math/rand"
)

const (
	StartHP        = 100
	StartSAN       = 100
	StartOil       = 90
	MaxOil         = 120
	MaxSAN         = 100
	MaxSalt        = 3
	TotalSeals     = 3
	SqueezeHPCost  = 20
	SqueezeSANCost = 8
	WardTurns      = 6
	FloorHeal      = 15
)

type Action int

const (
	ActionNone Action = iota
	ActionUp
	ActionDown
	ActionLeft
	ActionRight
	ActionWait
	ActionConfirm
	ActionCancel
	ActionUseSalt
	ActionReadPage
	ActionEatSugar
	ActionHelp
	ActionQuit
	ActionRestart
	ActionChoose1
	ActionChoose2
)

type Phase int

const (
	PhaseTitle Phase = iota
	PhaseIntro
	PhasePlaying
	PhaseEnding
)

type Tone int

const (
	ToneNormal Tone = iota
	ToneGood
	ToneBad
	ToneWhisper
	ToneMystic
)

type Message struct {
	Text string
	Tone Tone
}

type ModalKind int

const (
	ModalHelp ModalKind = iota
	ModalLore
	ModalFloor
	ModalQuit
	ModalGate
)

type Modal struct {
	Kind   ModalKind
	Title  string
	Lines  []string
	Footer string
}

type EndingKind int

const (
	EndingSeal EndingKind = iota
	EndingEmbrace
	EndingDeath
	EndingMadness
)

type Ending struct {
	Kind    EndingKind
	Title   string
	English string
	Body    []string
}

type Model struct {
	rng   *rand.Rand
	seed  int64
	phase Phase
	modal *Modal

	quit   bool
	ending *Ending

	depth  int
	level  *Level
	player Pos

	hp    int
	san   int
	oil   int
	salt  int
	sugar int
	seals int
	pages []int

	turn      int
	steps     int
	wardTurns int
	wardFresh bool

	messages []Message

	sanTier         int
	oilWarned       bool
	lastBumpTurn    int
	lastWardLogTurn int
}

func NewModel(seed int64) *Model {
	return &Model{
		rng:   rand.New(rand.NewSource(seed)),
		seed:  seed,
		phase: PhaseTitle,
	}
}

func (m *Model) Handle(action Action) {
	if action == ActionNone || m.quit {
		return
	}
	if m.modal != nil {
		m.handleModal(action)
		return
	}
	switch m.phase {
	case PhaseTitle:
		m.handleTitle(action)
	case PhaseIntro:
		m.handleIntro(action)
	case PhasePlaying:
		m.handlePlaying(action)
	case PhaseEnding:
		m.handleEnding(action)
	}
}

func (m *Model) handleTitle(action Action) {
	switch action {
	case ActionConfirm, ActionWait:
		m.phase = PhaseIntro
	case ActionHelp:
		m.openHelp()
	case ActionQuit:
		m.quit = true
	}
}

func (m *Model) handleIntro(action Action) {
	switch action {
	case ActionConfirm, ActionWait:
		m.startRun()
	case ActionCancel:
		m.phase = PhaseTitle
	case ActionHelp:
		m.openHelp()
	case ActionQuit:
		m.quit = true
	}
}

func (m *Model) handleEnding(action Action) {
	switch action {
	case ActionRestart:
		m.startRun()
	case ActionCancel:
		m.phase = PhaseTitle
		m.ending = nil
	case ActionQuit:
		m.quit = true
	case ActionHelp:
		m.openHelp()
	}
}

func (m *Model) handlePlaying(action Action) {
	switch action {
	case ActionUp:
		m.tryMove(0, -1)
	case ActionDown:
		m.tryMove(0, 1)
	case ActionLeft:
		m.tryMove(-1, 0)
	case ActionRight:
		m.tryMove(1, 0)
	case ActionWait:
		waits := []string{
			"你屏住呼吸，听着水滴落在石头上的声音。",
			"你数着自己的心跳，等黑暗先动。",
			"远处有什么拖着脚步经过。你没有回头。",
			"灯焰伏低，又慢慢站起来。",
		}
		m.log(waits[m.rng.Intn(len(waits))], ToneNormal)
		m.advanceTurn()
	case ActionUseSalt:
		m.useSalt()
	case ActionReadPage:
		m.usePage()
	case ActionEatSugar:
		m.useSugar()
	case ActionConfirm:
		m.interact()
	case ActionHelp:
		m.openHelp()
	case ActionQuit:
		m.modal = &Modal{
			Kind:   ModalQuit,
			Title:  "放弃调查",
			Lines:  []string{"雾港的雾会吞掉你的脚印。", "", "确定要离开这场调查吗？进度不会保存。"},
			Footer: "Enter / Y 离开   Esc / N 留下",
		}
	}
}

func (m *Model) handleModal(action Action) {
	switch m.modal.Kind {
	case ModalQuit:
		switch action {
		case ActionConfirm:
			m.quit = true
		case ActionCancel:
			m.modal = nil
		}
	case ModalGate:
		switch action {
		case ActionChoose1:
			m.finish(EndingSeal)
		case ActionChoose2:
			m.finish(EndingEmbrace)
		case ActionCancel:
			m.modal = nil
		}
	case ModalHelp:
		if action == ActionConfirm || action == ActionCancel || action == ActionHelp {
			m.modal = nil
		}
	default:
		if action == ActionConfirm || action == ActionCancel || action == ActionWait {
			m.modal = nil
		}
	}
}

func (m *Model) openHelp() {
	m.modal = &Modal{
		Kind:   ModalHelp,
		Title:  "调查手册",
		Footer: "Enter / Esc 返回",
	}
}

func (m *Model) startRun() {
	m.phase = PhasePlaying
	m.ending = nil
	m.modal = nil
	m.depth = 1
	m.hp = StartHP
	m.san = StartSAN
	m.oil = StartOil
	m.salt = 2
	m.sugar = 0
	m.seals = 0
	m.pages = nil
	m.turn = 0
	m.steps = 0
	m.wardTurns = 0
	m.wardFresh = false
	m.sanTier = 0
	m.oilWarned = false
	m.lastBumpTurn = -10
	m.lastWardLogTurn = -1
	m.messages = nil

	m.level = GenerateLevel(1, m.rng)
	m.player = m.level.Start
	m.level.ComputeFOV(m.player, m.lightRadius())
	m.log("煤油灯亮了。石阶向下延伸，看不见尽头。", ToneMystic)
	m.showFloorModal()
}

func (m *Model) showFloorModal() {
	info := floorInfoFor(m.depth)
	lines := []string{info.English, "", info.Description, "", info.Hint}
	if _, ok := m.level.sealPosition(); ok {
		lines = append(lines, "", fmt.Sprintf("刻印方位：%s。", m.sealLocator()))
	}
	lines = append(lines, "")
	m.modal = &Modal{
		Kind:   ModalFloor,
		Title:  fmt.Sprintf("第%s层 · %s", chineseNumber(m.depth), info.Name),
		Lines:  lines,
		Footer: "Enter 继续探索",
	}
}

func (m *Model) tryMove(dx, dy int) {
	target := Pos{X: m.player.X + dx, Y: m.player.Y + dy}
	if !m.level.walkable(target) {
		m.bump()
		return
	}
	if enemy := m.level.enemyAt(target); enemy != nil {
		m.squeezePast(enemy)
		return
	}
	m.player = target
	m.steps++
	m.advanceTurn()
}

func (m *Model) squeezePast(enemy *Enemy) {
	old := m.player
	if enemy.Stun > 0 || m.wardTurns > 0 {
		m.log("盐圈护着你。你从它身侧滑了过去。", ToneGood)
	} else {
		m.hp -= SqueezeHPCost
		m.san -= SqueezeSANCost
		m.log(fmt.Sprintf("你撞进它的怀里，从另一边跌出去。 [-%d 生命 / -%d 理智]", SqueezeHPCost, SqueezeSANCost), ToneBad)
	}
	m.player = enemy.Pos
	enemy.Pos = old
	enemy.Cooldown = maxInt(enemy.Cooldown, 2)
	m.steps++
	m.advanceTurn()
}

func (m *Model) bump() {
	if m.turn-m.lastBumpTurn < 3 {
		return
	}
	m.lastBumpTurn = m.turn
	lines := []string{
		"湿冷的石壁拦住了去路。",
		"你摸到一片黏滑的苔藓。墙后似乎有什么在呼吸。",
		"再往前只有墙。或者，墙在假装自己是墙。",
	}
	m.log(lines[m.rng.Intn(len(lines))], ToneNormal)
}

func (m *Model) advanceTurn() {
	if m.phase != PhasePlaying {
		return
	}
	m.turn++

	beforeOil := m.oil
	if m.oil > 0 {
		m.oil--
	}
	if beforeOil > 0 && m.oil == 0 {
		m.log("灯芯爆出最后一点火星。黑暗合拢了。", ToneBad)
	}
	if m.oil > 30 {
		m.oilWarned = false
	}
	if m.oil > 0 && m.oil <= 20 && !m.oilWarned {
		m.log("煤油不多了。灯焰开始发抖。", ToneBad)
		m.oilWarned = true
	}

	m.level.ComputeFOV(m.player, m.lightRadius())
	m.checkPickups()
	m.checkGaze()
	m.moveEnemies()
	if m.wardTurns > 0 {
		if m.wardFresh {
			m.wardFresh = false
		} else {
			m.wardTurns--
			if m.wardTurns == 0 {
				m.log("盐线暗了下去。黑暗重新贴近你的脚跟。", ToneNormal)
			}
		}
	}

	if m.oil == 0 {
		if m.turn%2 == 0 {
			m.san--
		}
		if m.turn%9 == 0 {
			m.log("黑暗贴在你的耳后，用你的声音数着步子。", ToneWhisper)
		}
	}
	m.ambience()
	m.updateSanityTier()
	m.checkEnd()
}

func (m *Model) lightRadius() int {
	switch {
	case m.oil > 30:
		return 6
	case m.oil > 0:
		return 4
	default:
		return 2
	}
}

func (m *Model) checkPickups() {
	if m.level == nil {
		return
	}
	remaining := m.level.Items[:0]
	for _, item := range m.level.Items {
		if item.Pos != m.player {
			remaining = append(remaining, item)
			continue
		}
		if !m.pickUp(item) {
			remaining = append(remaining, item)
		}
		if m.phase != PhasePlaying {
			break
		}
	}
	m.level.Items = remaining
}

func (m *Model) pickUp(item *Item) bool {
	switch item.Kind {
	case ItemOil:
		if m.oil >= MaxOil {
			m.log("灯油还是满的。你没有打开这一罐。", ToneNormal)
			return false
		}
		m.oil = minInt(MaxOil, m.oil+30)
		m.log("你拾起一罐煤油。灯焰重新挺直了脊背。 (+30 灯油)", ToneGood)
	case ItemSugar:
		m.sugar++
		m.log("你收起一块方糖。按 C 食用：理智未满时恢复 20 点。 (+1 方糖)", ToneGood)
	case ItemSalt:
		if m.salt >= MaxSalt {
			m.log("一小袋粗盐。你的口袋里已经装不下了。", ToneNormal)
			return false
		}
		m.salt++
		m.log("你收起一袋粗盐。盐能画出它们不愿跨过的线。 (+1 圣盐)", ToneGood)
	case ItemLore:
		m.pages = append(m.pages, m.depth)
		m.log("你拾起一页残纸。按 E 阅读：它会指出阶梯，但会磨损理智。 (+1 残页)", ToneMystic)
	case ItemSeal:
		m.seals++
		m.level.SealTaken = true
		m.level.RevealStairs()
		m.san = minInt(MaxSAN, m.san+4)
		m.log(fmt.Sprintf("%s在你掌心发烫。阶梯的封印松动了。", sealName(m.depth)), ToneGood)
		m.log(fmt.Sprintf("刻印 ◈ %d/%d 归位。阶梯在%s。", m.seals, TotalSeals, directionName(m.player, m.level.Stairs)), ToneMystic)
	}
	return true
}

func (m *Model) usePage() {
	if len(m.pages) == 0 {
		m.log("你没有残页。地图上的 ? 需要先拾取。", ToneNormal)
		return
	}
	pageDepth := m.pages[len(m.pages)-1]
	m.pages = m.pages[:len(m.pages)-1]
	m.san -= 6
	m.level.RevealStairs()
	title, lines := lorePage(pageDepth)
	label := "阶梯"
	if m.depth == TotalSeals {
		label = "潮汐之门"
	}
	m.modal = &Modal{
		Kind:   ModalLore,
		Title:  title,
		Lines:  append(lines, "", fmt.Sprintf("你记住了%s的位置：%s。", label, directionName(m.player, m.level.Stairs))),
		Footer: "Enter 收起残页",
	}
	m.log("残页上的墨迹在你的视网膜里燃烧。 (-6 理智)", ToneWhisper)
	m.updateSanityTier()
	m.checkEnd()
}

func (m *Model) useSugar() {
	if m.sugar <= 0 {
		m.log("你没有方糖。地图上的 ▫ 需要先拾取。", ToneNormal)
		return
	}
	if m.san >= MaxSAN {
		m.log("你现在不需要方糖。它还能再放一会儿。", ToneNormal)
		return
	}
	gain := minInt(20, MaxSAN-m.san)
	m.sugar--
	m.san += gain
	m.log(fmt.Sprintf("你含化一块方糖。甜味把世界暂时按回原位。 (+%d 理智)", gain), ToneGood)
}

func (m *Model) checkGaze() {
	for _, enemy := range m.level.Enemies {
		if !m.level.Visible[m.level.index(enemy.Pos)] || enemy.Gazed {
			continue
		}
		enemy.Gazed = true
		sanLoss := 0
		switch enemy.Kind {
		case EnemyWhisperer:
			sanLoss = 3
		case EnemyDeepOne:
			sanLoss = 5
		case EnemyCrawler:
			sanLoss = 7
		}
		m.san -= sanLoss
		m.log(gazeText(enemy.Kind), ToneWhisper)
	}
}

func (m *Model) moveEnemies() {
	field := m.level.distanceField(m.player)
	occupied := make(map[Pos]bool, len(m.level.Enemies))
	for _, enemy := range m.level.Enemies {
		occupied[enemy.Pos] = true
	}
	lostTrackLogged := false

	for _, enemy := range m.level.Enemies {
		if enemy.Stun > 0 {
			enemy.Stun--
			if enemy.Cooldown > 0 {
				enemy.Cooldown--
			}
			continue
		}
		if enemy.Cooldown > 0 {
			enemy.Cooldown--
		}

		distance := manhattan(enemy.Pos, m.player)
		if distance <= 1 && enemy.Cooldown == 0 {
			m.enemyAttack(enemy)
			enemy.Cooldown = 2
			continue
		}
		if distance <= 1 {
			continue
		}

		sees := distance <= 2 || (distance <= 8 && m.level.Visible[m.level.index(enemy.Pos)])
		if enemy.Alert {
			if sees {
				enemy.LostSight = 0
			} else {
				enemy.LostSight++
			}
			if enemy.LostSight >= 3 || distance > 14 {
				enemy.Alert = false
				enemy.LostSight = 0
				if !lostTrackLogged && distance <= 12 {
					m.log("远处的脚步停了。它失去了你的踪迹。", ToneGood)
					lostTrackLogged = true
				}
			}
		} else if sees {
			enemy.Alert = true
			enemy.LostSight = 0
		}

		delete(occupied, enemy.Pos)
		if enemy.Alert {
			if enemy.MoveClock < enemySpeed(enemy.Kind)-1 {
				enemy.MoveClock++
				occupied[enemy.Pos] = true
				continue
			}
			enemy.MoveClock = 0
			next, ok := m.stepEnemy(enemy, field, occupied)
			if ok {
				enemy.Pos = next
			}
		} else if m.rng.Intn(4) == 0 {
			next, ok := m.wanderEnemy(enemy, occupied)
			if ok {
				enemy.Pos = next
			}
		}
		occupied[enemy.Pos] = true
	}
}

func (m *Model) stepEnemy(enemy *Enemy, field []int, occupied map[Pos]bool) (Pos, bool) {
	current := field[m.level.index(enemy.Pos)]
	if current <= 0 {
		return enemy.Pos, false
	}
	best := make([]Pos, 0, 4)
	bestDist := current
	for _, d := range cardinal {
		next := Pos{X: enemy.Pos.X + d.X, Y: enemy.Pos.Y + d.Y}
		if !m.level.walkable(next) || next == m.player || occupied[next] {
			continue
		}
		dist := field[m.level.index(next)]
		if dist < 0 {
			continue
		}
		if dist < bestDist {
			bestDist = dist
			best = best[:0]
			best = append(best, next)
		} else if dist == bestDist {
			best = append(best, next)
		}
	}
	if len(best) == 0 {
		return enemy.Pos, false
	}
	return best[m.rng.Intn(len(best))], true
}

func (m *Model) wanderEnemy(enemy *Enemy, occupied map[Pos]bool) (Pos, bool) {
	options := make([]Pos, 0, 4)
	for _, d := range cardinal {
		next := Pos{X: enemy.Pos.X + d.X, Y: enemy.Pos.Y + d.Y}
		if m.level.walkable(next) && next != m.player && !occupied[next] {
			options = append(options, next)
		}
	}
	if len(options) == 0 {
		return enemy.Pos, false
	}
	return options[m.rng.Intn(len(options))], true
}

func (m *Model) enemyAttack(enemy *Enemy) {
	if m.wardTurns > 0 {
		enemy.Cooldown = maxInt(enemy.Cooldown, 1)
		if m.lastWardLogTurn != m.turn {
			m.log("盐线骤然发亮。它停在圈外，喉咙里发出潮湿的声音。", ToneGood)
			m.lastWardLogTurn = m.turn
		}
		return
	}
	var hpLoss, sanLoss int
	switch enemy.Kind {
	case EnemyWhisperer:
		hpLoss, sanLoss = 8, 7
	case EnemyDeepOne:
		hpLoss, sanLoss = 20, 4
	case EnemyCrawler:
		hpLoss, sanLoss = 14, 6
	}
	m.hp -= hpLoss
	m.san -= sanLoss
	enemy.Stun = maxInt(enemy.Stun, 2)
	enemy.Alert = false
	enemy.LostSight = 0
	m.log(fmt.Sprintf("%s [-%d 生命 / -%d 理智]", attackText(enemy.Kind), hpLoss, sanLoss), ToneBad)
	m.log("它伏下去撕咬伤口。你有两个回合的空隙。", ToneGood)
}

func (m *Model) useSalt() {
	if m.salt <= 0 {
		m.log("盐袋空了。你只能握紧手里那点温热的灰。", ToneBad)
		return
	}
	m.salt--
	m.wardTurns = WardTurns
	m.wardFresh = true
	m.lastWardLogTurn = -1
	affected := 0
	for _, enemy := range m.level.Enemies {
		if manhattan(enemy.Pos, m.player) <= 6 {
			enemy.Stun = WardTurns
			enemy.Cooldown = maxInt(enemy.Cooldown, WardTurns)
			enemy.Alert = false
			affected++
		}
	}
	m.san = minInt(MaxSAN, m.san+4)
	if affected > 0 {
		m.log(fmt.Sprintf("盐粒落地，画出一圈白光。%d 回合内，没有东西能伤到你。 (+4 理智)", WardTurns), ToneGood)
	} else {
		m.log(fmt.Sprintf("盐粒落地，画出一圈白光。%d 回合内，没有东西能伤到你。 (+4 理智)", WardTurns), ToneMystic)
	}
	m.advanceTurn()
}

func (m *Model) ambience() {
	if m.san <= 25 && m.turn%9 == 0 {
		m.log(randomWhisper(m.rng), ToneWhisper)
	}
	if m.depth == 3 && m.turn%23 == 0 {
		m.log("整座深渊缓慢地吸气。你的骨头跟着发颤。", ToneWhisper)
	}
	if m.hp <= 25 && m.turn%17 == 0 {
		m.log("血把袖口黏在皮肤上。你还能走。", ToneBad)
	}
}

func (m *Model) updateSanityTier() {
	tier := 0
	switch {
	case m.san <= 10:
		tier = 3
	case m.san <= 25:
		tier = 2
	case m.san <= 50:
		tier = 1
	}
	if tier > m.sanTier {
		switch tier {
		case 1:
			m.log("墙缝里传来极轻的祷告声。你发现自己正在跟着念。", ToneWhisper)
		case 2:
			m.log("你看见自己的影子慢了半拍。", ToneWhisper)
		case 3:
			m.log("低语不再来自外面。它们就在你的牙缝里。", ToneWhisper)
		}
	}
	m.sanTier = tier
}

func (m *Model) interact() {
	if m.level == nil || m.player != m.level.Stairs {
		m.log("这里没有可以互动的阶梯。", ToneNormal)
		return
	}
	if m.depth == TotalSeals {
		if m.seals < TotalSeals {
			m.log(fmt.Sprintf("潮汐之门纹丝不动。%s：%s。", sealName(m.depth), m.sealLocator()), ToneBad)
			return
		}
		m.modal = &Modal{
			Kind:   ModalGate,
			Title:  "潮汐之门",
			Lines:  append(gateText(), "", "1  以三枚刻印封门", "2  倾听门后的声音"),
			Footer: "1 / 2 做出选择   Esc 后退",
		}
		return
	}
	if !m.level.SealTaken {
		m.log(fmt.Sprintf("阶梯被一层盐晶封住。%s：%s。", sealName(m.depth), m.sealLocator()), ToneBad)
		return
	}
	m.descend()
}

func (m *Model) descend() {
	m.depth++
	m.hp = minInt(StartHP, m.hp+FloorHeal)
	m.san = minInt(MaxSAN, m.san+3)
	m.wardTurns = 0
	m.wardFresh = false
	m.level = GenerateLevel(m.depth, m.rng)
	m.player = m.level.Start
	m.oilWarned = m.oil <= 20
	m.level.ComputeFOV(m.player, m.lightRadius())
	m.modal = nil
	m.showFloorModal()
	m.log("你沿阶梯下沉。身后的门没有发出任何声音。", ToneMystic)
}

func (m *Model) finish(kind EndingKind) {
	m.modal = nil
	m.ending = buildEnding(kind)
	m.phase = PhaseEnding
}

func (m *Model) checkEnd() {
	if m.phase != PhasePlaying {
		return
	}
	if m.hp <= 0 {
		m.finish(EndingDeath)
		return
	}
	if m.san <= 0 {
		m.finish(EndingMadness)
	}
}

func (m *Model) log(text string, tone Tone) {
	m.messages = append(m.messages, Message{Text: text, Tone: tone})
	if len(m.messages) > 40 {
		m.messages = m.messages[len(m.messages)-40:]
	}
}

func (m *Model) sealLocator() string {
	if m.level == nil {
		return "未知"
	}
	pos, ok := m.level.sealPosition()
	if !ok {
		return "未知"
	}
	distance := manhattan(m.player, pos)
	switch {
	case distance == 0:
		return "就在脚下"
	case distance <= 3:
		return "就在附近"
	default:
		return fmt.Sprintf("%s，约 %d 步", directionName(m.player, pos), distance)
	}
}

func directionName(from, to Pos) string {
	dx := to.X - from.X
	dy := to.Y - from.Y
	if dx == 0 && dy == 0 {
		return "脚下"
	}
	vertical := ""
	if dy < 0 {
		vertical = "北"
	} else if dy > 0 {
		vertical = "南"
	}
	horizontal := ""
	if dx < 0 {
		horizontal = "西"
	} else if dx > 0 {
		horizontal = "东"
	}
	if vertical == "" {
		return horizontal + "方"
	}
	if horizontal == "" {
		return vertical + "方"
	}
	return horizontal + vertical + "方"
}

func enemySpeed(kind EnemyKind) int {
	switch kind {
	case EnemyWhisperer:
		return 2
	case EnemyCrawler:
		return 1
	default:
		return 1
	}
}
