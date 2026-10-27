package game

import "math/rand"

type FloorInfo struct {
	Name        string
	English     string
	Description string
	Hint        string
}

var floorInfos = []FloorInfo{
	{
		Name:        "潮汐教堂",
		English:     "THE DROWNED CHAPEL",
		Description: "盐霜覆盖着长椅。祭坛下的石阶不该存在——它比教堂本身更古老。",
		Hint:        "寻找骨之刻印。它在最深的祷词里。",
	},
	{
		Name:        "沉没墓穴",
		English:     "THE SUNKEN CRYPT",
		Description: "海水从墙缝渗进来，淹没脚踝。墓穴里的死者没有被安葬，他们只是被留在原地。",
		Hint:        "寻找盐之刻印。不要回应刻在棺盖上的名字。",
	},
	{
		Name:        "无光之海",
		English:     "THE LIGHTLESS SEA",
		Description: "这里没有天顶。脚下的石板漂浮在黑色的海上，远处有某种巨大的东西在呼吸。",
		Hint:        "寻找梦之刻印，然后走向潮汐之门。",
	},
}

var introParagraphs = []string{
	"1927 年，雾港。",
	"灯塔在三日前熄灭。渔民说，海底有东西在敲钟。",
	"你带着一盏煤油灯、一袋盐和三枚空白刻印。",
	"潮汐教堂下方，有一道不该存在的阶梯。",
	"不要让灯熄灭。\n不要直视它们。\n不要回应低语。",
}

var introCompact = []string{
	"1927 年，雾港。",
	"灯塔熄灭。海底有东西在敲钟。",
	"你带着煤油灯、盐与三枚刻印，",
	"走进潮汐教堂下的阶梯。",
	"不要让灯熄灭。不要直视它们。",
	"不要回应低语。",
}

var titleQuotes = []string{
	"「群星尚未归位，但海已经在做梦。」",
	"「有些门从建成那天起，就一直在等人推开。」",
	"「你不是来调查的。你是来被记住的。」",
}

var loreTitles = map[int]string{
	1: "残页 · 圣咏",
	2: "残页 · 航海日志",
	3: "残页 · 无名",
}

var loreLines = map[int][]string{
	1: {
		"主祷文被划掉了，旁边用指甲补上一行：",
		"「祂不居住在海底。海底只是祂翻身时留下的凹痕。」",
		"纸背面是一幅教堂剖面图。你看见阶梯，正在掌心里发烫。",
	},
	2: {
		"第七日，所有船员做了同一个梦。梦里我们走在干燥的海床上，头顶悬着一座倒置的城市。",
		"船长下令把钟沉进水里。钟没有沉。它开始呼吸。",
		"最后一页只有水渍。",
	},
	3: {
		"纸上反复写着同一句话，字迹一层叠着一层：",
		"「门不是用来打开的。门是用来阻止另一边醒来的。」",
		"最后几个字变成了压痕，像是有人隔着纸，用指甲写完了余生。",
	},
}

var whisperLines = []string{
	"你听见有人用你的声音，从很深的地方喊你的名字。",
	"墙上的水痕排列成一行字。你眨了一次眼，它就不见了。",
	"有东西在你的余光里站起来，又在你转头时坐下。",
	"你的影子在墙上多了一只手。它正朝你挥。",
	"水滴声停了。整座深渊在等你先动。",
	"有人在你背后翻动那本你从没带下来的日志。",
}

var sealNames = []string{"", "骨之刻印", "盐之刻印", "梦之刻印"}

func floorInfoFor(depth int) FloorInfo {
	if depth < 1 {
		depth = 1
	}
	if depth > len(floorInfos) {
		depth = len(floorInfos)
	}
	return floorInfos[depth-1]
}

func chineseNumber(n int) string {
	switch n {
	case 1:
		return "一"
	case 2:
		return "二"
	case 3:
		return "三"
	default:
		return "?"
	}
}

func sealName(depth int) string {
	if depth >= 1 && depth < len(sealNames) {
		return sealNames[depth]
	}
	return "无名刻印"
}

func lorePage(depth int) (string, []string) {
	title, ok := loreTitles[depth]
	if !ok {
		title = "残页 · 无名"
	}
	lines, ok := loreLines[depth]
	if !ok {
		lines = []string{"字迹被海水泡烂了。你只认出最后一个词：醒来。"}
	}
	return title, append([]string(nil), lines...)
}

func gazeText(kind EnemyKind) string {
	switch kind {
	case EnemyWhisperer:
		return "一个披着盐壳的人形正在祷告。它没有脸，但你听懂了它的祷词。"
	case EnemyDeepOne:
		return "它从阴影里侧过头。那双眼睛比人类更旧。"
	default:
		return "有什么东西把四肢折向不该折的方向，向你爬来。"
	}
}

func attackText(kind EnemyKind) string {
	switch kind {
	case EnemyWhisperer:
		return "低语者贴在你耳边，念出你童年的秘密。"
	case EnemyDeepOne:
		return "深潜者的爪子撕开你的肩膀。"
	default:
		return "无面者从墙上剥落，扑向你。"
	}
}

func gateText() []string {
	return []string{
		"三枚刻印在你掌心排成一条直线。",
		"门缝后透出潮湿的风，像有什么东西在极深的地方翻身。",
		"门的另一边在等你做出选择。",
	}
}

func randomWhisper(rng *rand.Rand) string {
	return whisperLines[rng.Intn(len(whisperLines))]
}

func buildEnding(kind EndingKind) *Ending {
	switch kind {
	case EndingSeal:
		return &Ending{
			Kind:    kind,
			Title:   "黎 明 之 前",
			English: "BEFORE THE DAWN",
			Body: []string{
				"你把三枚刻印按进潮汐之门的凹槽。\n门后传来一声漫长的叹息，像整片海在退潮。",
				"当你爬回教堂时，天边已经发白。雾港的钟没有响。",
				"但你的影子里，多出了一道不属于你的轮廓。它很安静，只是在等。",
			},
		}
	case EndingEmbrace:
		return &Ending{
			Kind:    kind,
			Title:   "深 渊 之 嗣",
			English: "CHILD OF THE ABYSS",
			Body: []string{
				"你俯身，把耳朵贴在门上。门后有什么东西，也在听你。",
				"它问了你一个名字。你回答了。",
				"第二天，渔民在岸边发现一盏熄灭的灯。灯旁有一串脚印，走向海里。",
				"雾港的雾更浓了。",
			},
		}
	case EndingDeath:
		return &Ending{
			Kind:    kind,
			Title:   "退 潮 之 后",
			English: "AFTER THE TIDE",
			Body: []string{
				"你的手指在湿冷的地面上抓出一道痕迹。",
				"黑暗没有变深，只是变得更有耐心。",
				"退潮时，海把尸体推回岸边。它没有名字，眼睛睁着，望着内陆。",
			},
		}
	default:
		return &Ending{
			Kind:    kind,
			Title:   "低 语 成 句",
			English: "THE WHISPER BECOMES A WORD",
			Body: []string{
				"低语终于连成了句子。你终于听懂了。",
				"你开始笑。笑声很轻，正好和墙后的另一个你重合。",
				"调查员再也没有离开雾港。但每天夜里，教堂里都会多一盏亮着的灯。",
			},
		}
	}
}
