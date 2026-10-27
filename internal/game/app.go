package game

import (
	"unicode"

	"github.com/gdamore/tcell/v2"
)

type App struct {
	screen tcell.Screen
	model  *Model
}

func NewApp(screen tcell.Screen, seed int64) *App {
	return &App{
		screen: screen,
		model:  NewModel(seed),
	}
}

func (a *App) Run() error {
	if err := a.screen.Init(); err != nil {
		return err
	}
	defer a.screen.Fini()

	a.screen.SetStyle(styleBase)
	a.screen.HideCursor()
	a.screen.Clear()
	a.draw()

	for !a.model.quit {
		event := a.screen.PollEvent()
		if event == nil {
			break
		}
		switch ev := event.(type) {
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyCtrlC {
				a.model.quit = true
				continue
			}
			action := actionForKey(ev)
			if action == ActionNone {
				continue
			}
			a.model.Handle(action)
			a.draw()
		case *tcell.EventResize:
			a.screen.Sync()
			a.draw()
		}
	}
	return nil
}

func actionForKey(event *tcell.EventKey) Action {
	switch event.Key() {
	case tcell.KeyUp:
		return ActionUp
	case tcell.KeyDown:
		return ActionDown
	case tcell.KeyLeft:
		return ActionLeft
	case tcell.KeyRight:
		return ActionRight
	case tcell.KeyEnter:
		return ActionConfirm
	case tcell.KeyEsc:
		return ActionCancel
	case tcell.KeyCtrlC:
		return ActionQuit
	}
	r := unicode.ToLower(event.Rune())
	switch r {
	case 'w':
		return ActionUp
	case 's':
		return ActionDown
	case 'a':
		return ActionLeft
	case 'd':
		return ActionRight
	case ' ':
		return ActionWait
	case 'f':
		return ActionUseSalt
	case 'e':
		return ActionReadPage
	case 'c':
		return ActionEatSugar
	case '?', 'h':
		return ActionHelp
	case 'q':
		return ActionQuit
	case 'r':
		return ActionRestart
	case 'y':
		return ActionConfirm
	case 'n':
		return ActionCancel
	case '1':
		return ActionChoose1
	case '2':
		return ActionChoose2
	}
	return ActionNone
}
