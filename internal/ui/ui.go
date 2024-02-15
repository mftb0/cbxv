package ui

import (
	"fmt"
	"strings"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mftb0/cbxv/internal/model"
	"github.com/mftb0/cbxv/internal/util"
)

type View interface {
	Connect(m *model.Model, u *UI)
	Disconnect(m *model.Model, u *UI)
	Render(m *model.Model)
}

type UI struct {
	SendMessage util.Messenger
	MainWindow  *gtk.Window
	PageView    View
	StripView   View
	View        View
    Commands    *CommandList
}

func NewUI(m *model.Model, messenger util.Messenger) *UI {
	gtk.Init(nil)
	u := &UI{}
	u.SendMessage = messenger
	u.MainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	u.MainWindow.SetPosition(gtk.WIN_POS_CENTER)
	u.MainWindow.SetTitle(m.ProgramName)
	u.MainWindow.Connect("delete-event", func() bool {
        u.Commands.Names["quit"].Execute()
        return true
	})
	u.MainWindow.SetDefaultSize(1024, 768)

	iPath := util.AppIconPath()
	if iPath != nil {
		u.MainWindow.SetIconFromFile(*iPath)
	}

	u.PageView = NewPageView(m, u, messenger)
	u.StripView = NewStripView(m, u)
	u.View = u.PageView

	initCss()

    u.Commands = NewCommands(m, u)

	u.MainWindow.ShowAll()

	return u
}

func (u *UI) Run() {
	gtk.Main()
}

func (u *UI) Quit() {
	gtk.MainQuit()
}

func (u *UI) Dispose() {
	//noop may need cleanup eventually
}

func (u *UI) RunFunc(f interface{}) {
	glib.IdleAdd(f)
}

func (u *UI) DisplayErrorDlg(message string) {
    dlg := gtk.MessageDialogNewWithMarkup(u.MainWindow,
        gtk.DialogFlags(gtk.DIALOG_MODAL),
        gtk.MESSAGE_ERROR, gtk.BUTTONS_CLOSE, "Error")
    defer dlg.Destroy()

    dlg.SetTitle("Error")
    dlg.SetMarkup(message)
    css, _ := dlg.GetStyleContext()
    css.AddClass("msg-dlg")

    dlg.Run()
}

func (u *UI) ShowCursor() {
    d, _ := gdk.DisplayGetDefault()
    c, _ := gdk.CursorNewFromName(d, "default")
    w, err := u.MainWindow.GetWindow()
    if err != nil {
        return
    }
    w.SetCursor(c)
}

func (u *UI) HideCursor() {
    d, _ := gdk.DisplayGetDefault()
    c, _ := gdk.CursorNewFromName(d, "none")
    w, err := u.MainWindow.GetWindow()
    if err != nil {
        return
    }
    w.SetCursor(c)
}

func (u *UI) Render(m *model.Model) {
	glib.IdleAdd(func() {
		u.View.Render(m)

		// causes the draw event to fire
		// which gets the canvas to Render
		// see initRenderer
		u.MainWindow.QueueDraw()
	})
}

func initCss() {
	css, err := gtk.CssProviderNew()
	if err != nil {
		fmt.Printf("css error %s\n", err)
	}

	data, err := util.LoadTextFile("assets/index.css")
	if err != nil {
		fmt.Printf("error loading file%s\n", err)
	}

	err = css.LoadFromData(*data)
	if err != nil {
		fmt.Printf("css error %s\n", err)
	}

	s, err := gdk.ScreenGetDefault()
	if err != nil {
		fmt.Printf("css error %s\n", err)
	}

	gtk.AddProviderForScreen(s, css, gtk.STYLE_PROVIDER_PRIORITY_APPLICATION)
}

func handleDropData(buf []byte, command *Command) {
    uris := strings.Split(string(buf), "\n")
    if len(uris) > 0 {
        p := util.ParseFileUrl(strings.Trim(uris[0], "\r\n\t"))
        if p != nil {
            command.Execute(*p)
        }
    }
}

