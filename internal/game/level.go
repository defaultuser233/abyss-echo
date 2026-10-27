package game

import "math/rand"

const (
	LevelWidth  = 61
	LevelHeight = 27
)

type Pos struct {
	X int
	Y int
}

type Tile uint8

const (
	TileWall Tile = iota
	TileFloor
)

type ItemKind uint8

const (
	ItemOil ItemKind = iota
	ItemSugar
	ItemSalt
	ItemLore
	ItemSeal
)

type Item struct {
	Kind ItemKind
	Pos  Pos
}

type EnemyKind uint8

const (
	EnemyWhisperer EnemyKind = iota
	EnemyDeepOne
	EnemyCrawler
)

type Enemy struct {
	Kind      EnemyKind
	Pos       Pos
	Stun      int
	Cooldown  int
	MoveClock int
	Alert     bool
	LostSight int
	Gazed     bool
}

type room struct {
	x, y, w, h int
}

func (r room) center() Pos {
	return Pos{X: r.x + r.w/2, Y: r.y + r.h/2}
}

type Level struct {
	Depth       int
	Width       int
	Height      int
	Tiles       []Tile
	Seen        []bool
	Visible     []bool
	Start       Pos
	Stairs      Pos
	StairsKnown bool
	SealTaken   bool
	Items       []*Item
	Enemies     []*Enemy
}

func GenerateLevel(depth int, rng *rand.Rand) *Level {
	for attempt := 0; attempt < 80; attempt++ {
		l := newLevel(depth)
		rooms := carveDungeon(l, rng)
		if len(rooms) < 6 {
			continue
		}
		l.Start = rooms[0].center()

		order := roomsByDistance(rooms, l.Start)
		stairsRoom := order[len(order)-1]
		l.Stairs = stairsRoom.center()

		sealRoom := order[len(order)/2]
		sealPos := sealRoom.center()
		if sealPos == l.Stairs {
			sealPos = Pos{X: sealPos.X + 1, Y: sealPos.Y}
			if !l.walkable(sealPos) {
				continue
			}
		}

		l.Items = append(l.Items, &Item{Kind: ItemSeal, Pos: sealPos})
		used := map[Pos]bool{
			l.Start:  true,
			l.Stairs: true,
			sealPos:  true,
		}

		itemKinds := []ItemKind{ItemOil, ItemOil, ItemSugar, ItemSalt, ItemLore}
		if depth >= 2 {
			itemKinds = append(itemKinds, ItemOil)
		}
		if depth >= 3 {
			itemKinds = append(itemKinds, ItemOil, ItemSugar)
		}
		for _, kind := range itemKinds {
			p, ok := freeRoomCell(l, rooms, used, rng)
			if !ok {
				break
			}
			l.Items = append(l.Items, &Item{Kind: kind, Pos: p})
			used[p] = true
		}

		counts := []struct {
			kind  EnemyKind
			count int
		}{
			{kind: EnemyWhisperer, count: 1 + depth/2},
			{kind: EnemyDeepOne, count: 1 + (depth-1)/2},
			{kind: EnemyCrawler, count: depth - 1},
		}
		for _, entry := range counts {
			for i := 0; i < entry.count; i++ {
				p, ok := enemySpawn(l, rooms, used, rng)
				if !ok {
					continue
				}
				l.Enemies = append(l.Enemies, &Enemy{Kind: entry.kind, Pos: p})
				used[p] = true
			}
		}

		l.ComputeFOV(l.Start, 6)
		if l.isValid() && l.reachableAll() {
			return l
		}
	}
	return fallbackLevel(depth, rng)
}

func newLevel(depth int) *Level {
	l := &Level{
		Depth:   depth,
		Width:   LevelWidth,
		Height:  LevelHeight,
		Tiles:   make([]Tile, LevelWidth*LevelHeight),
		Seen:    make([]bool, LevelWidth*LevelHeight),
		Visible: make([]bool, LevelWidth*LevelHeight),
	}
	return l
}

func (l *Level) index(p Pos) int {
	return p.Y*l.Width + p.X
}

func (l *Level) inBounds(p Pos) bool {
	return p.X >= 0 && p.Y >= 0 && p.X < l.Width && p.Y < l.Height
}

func (l *Level) tile(p Pos) Tile {
	if !l.inBounds(p) {
		return TileWall
	}
	return l.Tiles[l.index(p)]
}

func (l *Level) walkable(p Pos) bool {
	return l.inBounds(p) && l.tile(p) == TileFloor
}

func (l *Level) carve(p Pos) {
	if p.X > 0 && p.Y > 0 && p.X < l.Width-1 && p.Y < l.Height-1 {
		l.Tiles[l.index(p)] = TileFloor
	}
}

func (l *Level) itemAt(p Pos) *Item {
	for _, item := range l.Items {
		if item.Pos == p {
			return item
		}
	}
	return nil
}

func (l *Level) sealPosition() (Pos, bool) {
	for _, item := range l.Items {
		if item.Kind == ItemSeal {
			return item.Pos, true
		}
	}
	return Pos{}, false
}

func (l *Level) enemyAt(p Pos) *Enemy {
	for _, enemy := range l.Enemies {
		if enemy.Pos == p {
			return enemy
		}
	}
	return nil
}

func (l *Level) isValid() bool {
	if !l.walkable(l.Start) || !l.walkable(l.Stairs) {
		return false
	}
	for _, item := range l.Items {
		if !l.walkable(item.Pos) {
			return false
		}
	}
	for _, enemy := range l.Enemies {
		if !l.walkable(enemy.Pos) {
			return false
		}
	}
	return true
}

func (l *Level) reachableAll() bool {
	dist := l.distanceField(l.Start)
	for _, item := range l.Items {
		if dist[l.index(item.Pos)] < 0 {
			return false
		}
	}
	return dist[l.index(l.Stairs)] >= 0
}

func (l *Level) ComputeFOV(origin Pos, radius int) {
	for i := range l.Visible {
		l.Visible[i] = false
	}
	if radius < 1 {
		radius = 1
	}
	r2 := radius * radius
	minX := maxInt(0, origin.X-radius)
	maxX := minInt(l.Width-1, origin.X+radius)
	minY := maxInt(0, origin.Y-radius)
	maxY := minInt(l.Height-1, origin.Y+radius)
	for y := minY; y <= maxY; y++ {
		for x := minX; x <= maxX; x++ {
			dx := x - origin.X
			dy := y - origin.Y
			if dx*dx+dy*dy > r2 {
				continue
			}
			p := Pos{X: x, Y: y}
			if !l.lineOfSight(origin, p) {
				continue
			}
			i := l.index(p)
			l.Visible[i] = true
			l.Seen[i] = true
		}
	}
}

func (l *Level) lineOfSight(from, to Pos) bool {
	if from == to {
		return true
	}
	dx := absInt(to.X - from.X)
	dy := absInt(to.Y - from.Y)
	sx := 1
	if from.X > to.X {
		sx = -1
	}
	sy := 1
	if from.Y > to.Y {
		sy = -1
	}
	err := dx - dy
	x, y := from.X, from.Y
	for x != to.X || y != to.Y {
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
		if x == to.X && y == to.Y {
			break
		}
		if l.tile(Pos{X: x, Y: y}) == TileWall {
			return false
		}
	}
	return true
}

func (l *Level) distanceField(from Pos) []int {
	dist := make([]int, l.Width*l.Height)
	for i := range dist {
		dist[i] = -1
	}
	if !l.walkable(from) {
		return dist
	}
	dist[l.index(from)] = 0
	queue := make([]Pos, 1, l.Width*l.Height)
	queue[0] = from
	for head := 0; head < len(queue); head++ {
		cur := queue[head]
		base := dist[l.index(cur)]
		for _, d := range cardinal {
			next := Pos{X: cur.X + d.X, Y: cur.Y + d.Y}
			if !l.walkable(next) {
				continue
			}
			i := l.index(next)
			if dist[i] >= 0 {
				continue
			}
			dist[i] = base + 1
			queue = append(queue, next)
		}
	}
	return dist
}

func (l *Level) RevealStairs() {
	l.StairsKnown = true
	i := l.index(l.Stairs)
	l.Seen[i] = true
	for _, d := range cardinal {
		p := Pos{X: l.Stairs.X + d.X, Y: l.Stairs.Y + d.Y}
		if l.inBounds(p) {
			l.Seen[l.index(p)] = true
		}
	}
}

func carveDungeon(l *Level, rng *rand.Rand) []room {
	rooms := make([]room, 0, 13)
	const maxRooms = 13
	for tries := 0; tries < 240 && len(rooms) < maxRooms; tries++ {
		w := 5 + rng.Intn(6)
		h := 3 + rng.Intn(4)
		if l.Width-w-4 <= 0 || l.Height-h-4 <= 0 {
			continue
		}
		r := room{
			x: 2 + rng.Intn(l.Width-w-4),
			y: 2 + rng.Intn(l.Height-h-4),
			w: w,
			h: h,
		}
		overlap := false
		for _, other := range rooms {
			if rectsTouch(r, other) {
				overlap = true
				break
			}
		}
		if overlap {
			continue
		}
		for y := r.y; y < r.y+r.h; y++ {
			for x := r.x; x < r.x+r.w; x++ {
				l.carve(Pos{X: x, Y: y})
			}
		}
		rooms = append(rooms, r)
	}
	for i := 1; i < len(rooms); i++ {
		carveCorridor(l, rooms[i-1].center(), rooms[i].center(), rng)
	}
	for i := 0; i < len(rooms)-1; i++ {
		if len(rooms) > 4 && rng.Intn(3) == 0 {
			j := i + 2 + rng.Intn(2)
			if j < len(rooms) {
				carveCorridor(l, rooms[i].center(), rooms[j].center(), rng)
			}
		}
	}
	return rooms
}

func rectsTouch(a, b room) bool {
	return a.x-1 < b.x+b.w && a.x+a.w+1 > b.x &&
		a.y-1 < b.y+b.h && a.y+a.h+1 > b.y
}

func carveCorridor(l *Level, a, b Pos, rng *rand.Rand) {
	if rng.Intn(2) == 0 {
		carveHorizontal(l, a.X, b.X, a.Y)
		carveVertical(l, a.Y, b.Y, b.X)
	} else {
		carveVertical(l, a.Y, b.Y, a.X)
		carveHorizontal(l, a.X, b.X, b.Y)
	}
}

func carveHorizontal(l *Level, x0, x1, y int) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	for x := x0; x <= x1; x++ {
		l.carve(Pos{X: x, Y: y})
	}
}

func carveVertical(l *Level, y0, y1, x int) {
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		l.carve(Pos{X: x, Y: y})
	}
}

func roomsByDistance(rooms []room, origin Pos) []room {
	order := append([]room(nil), rooms...)
	for i := 1; i < len(order); i++ {
		for j := i; j > 0; j-- {
			a := manhattan(order[j-1].center(), origin)
			b := manhattan(order[j].center(), origin)
			if a <= b {
				break
			}
			order[j-1], order[j] = order[j], order[j-1]
		}
	}
	return order
}

func freeRoomCell(l *Level, rooms []room, used map[Pos]bool, rng *rand.Rand) (Pos, bool) {
	for tries := 0; tries < 400; tries++ {
		r := rooms[rng.Intn(len(rooms))]
		p := Pos{
			X: r.x + 1 + rng.Intn(maxInt(1, r.w-2)),
			Y: r.y + 1 + rng.Intn(maxInt(1, r.h-2)),
		}
		if used[p] || !l.walkable(p) {
			continue
		}
		return p, true
	}
	return Pos{}, false
}

func enemySpawn(l *Level, rooms []room, used map[Pos]bool, rng *rand.Rand) (Pos, bool) {
	for tries := 0; tries < 500; tries++ {
		r := rooms[rng.Intn(len(rooms))]
		p := Pos{
			X: r.x + rng.Intn(r.w),
			Y: r.y + rng.Intn(r.h),
		}
		if used[p] || !l.walkable(p) {
			continue
		}
		if manhattan(p, l.Start) < 7 {
			continue
		}
		return p, true
	}
	return Pos{}, false
}

func fallbackLevel(depth int, rng *rand.Rand) *Level {
	l := newLevel(depth)
	for y := 2; y < l.Height-2; y++ {
		for x := 2; x < l.Width-2; x++ {
			l.carve(Pos{X: x, Y: y})
		}
	}
	l.Start = Pos{X: 4, Y: l.Height / 2}
	l.Stairs = Pos{X: l.Width - 5, Y: l.Height / 2}
	l.Items = []*Item{
		{Kind: ItemSeal, Pos: Pos{X: l.Width / 2, Y: 5}},
		{Kind: ItemOil, Pos: Pos{X: l.Width / 2, Y: l.Height - 6}},
		{Kind: ItemSugar, Pos: Pos{X: 8, Y: 5}},
		{Kind: ItemSalt, Pos: Pos{X: 8, Y: l.Height - 6}},
		{Kind: ItemLore, Pos: Pos{X: l.Width - 9, Y: 5}},
	}
	for i := 0; i < depth+1; i++ {
		l.Enemies = append(l.Enemies, &Enemy{
			Kind: EnemyDeepOne,
			Pos:  Pos{X: 16 + i*5, Y: 8 + rng.Intn(l.Height-16)},
		})
	}
	l.ComputeFOV(l.Start, 6)
	return l
}

var cardinal = []Pos{{X: 0, Y: -1}, {X: 0, Y: 1}, {X: -1, Y: 0}, {X: 1, Y: 0}}

func manhattan(a, b Pos) int {
	return absInt(a.X-b.X) + absInt(a.Y-b.Y)
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
