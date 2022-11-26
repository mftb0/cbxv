package ui

import (
	"fmt"
	"path/filepath"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mftb0/cbxv-gotk3/internal/model"
	"github.com/mftb0/cbxv-gotk3/internal/util"
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
}

func NewUI(m *model.Model, messenger util.Messenger) *UI {
	gtk.Init(nil)
    if glib.MainContextDefault().IsOwner() {
        fmt.Printf("true\n")
    }
	u := &UI{}
	u.SendMessage = messenger
	u.MainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
	u.MainWindow.SetPosition(gtk.WIN_POS_CENTER)
	u.MainWindow.SetTitle(m.ProgramName)
	u.MainWindow.Connect("destroy", func() {
		gtk.MainQuit()
	})
	u.MainWindow.SetDefaultSize(1024, 768)

	iPath, _ := util.AppIconPath()
	if iPath != nil {
		u.MainWindow.SetIconFromFile(*iPath)
	}

	u.PageView = NewPageView(m, u, messenger)
	u.StripView = NewStripView(m, u, u.SendMessage)
	u.View = u.PageView

	initCss()

	u.initKBHandler(m)

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

func (u *UI) Render(m *model.Model) {
	glib.IdleAdd(func() {
		u.View.Render(m)

		// causes the draw event to fire
		// which gets the canvas to Render
		// see initRenderer
		u.MainWindow.QueueDraw()
	})
}

func (u *UI) initKBHandler(m *model.Model) {
	u.MainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		keyVal := keyEvent.KeyVal()
		switch keyVal {
		case gdk.KEY_q:
			u.SendMessage(util.Message{TypeName: "quit"})
			u.Quit()
		case gdk.KEY_1:
			u.View.Disconnect(m, u)
			u.View = u.PageView
			u.View.Connect(m, u)
			u.SendMessage(util.Message{TypeName: "setLayoutModeOnePage"})
		case gdk.KEY_2:
			u.View.Disconnect(m, u)
			u.View = u.PageView
			u.View.Connect(m, u)
			u.SendMessage(util.Message{TypeName: "setLayoutModeTwoPage"})
		case gdk.KEY_3:
			u.View.Disconnect(m, u)
			u.View = u.StripView
			u.View.Connect(m, u)
			u.SendMessage(util.Message{TypeName: "setLayoutModeLongStrip"})
		case gdk.KEY_f, gdk.KEY_F11:
			if m.Fullscreen {
				u.MainWindow.Unfullscreen()
			} else {
				u.MainWindow.Fullscreen()
			}
			u.SendMessage(util.Message{TypeName: "toggleFullscreen"})
		case gdk.KEY_o:
			dlg, _ := gtk.FileChooserNativeDialogNew("Open",
				u.MainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
			defer dlg.Destroy()

			dlg.SetCurrentFolder(m.BrowseDir)
			fltr, _ := gtk.FileFilterNew()
			fltr.AddPattern("*.cbz")
			fltr.AddPattern("*.cbr")
			fltr.AddPattern("*.cb7")
			fltr.AddPattern("*.cbt")
			fltr.SetName("cbx files")
			dlg.AddFilter(fltr)
			fltr, _ = gtk.FileFilterNew()
			fltr.AddPattern("*")
			fltr.SetName("All files")
			dlg.AddFilter(fltr)

			output := dlg.NativeDialog.Run()
			if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
				f := dlg.GetFilename()
				m := &util.Message{TypeName: "openFile", Data: f}
				u.SendMessage(*m)
			}
		case gdk.KEY_c:
			u.SendMessage(util.Message{TypeName: "closeFile"})
		case gdk.KEY_e:
			dlg, _ := gtk.FileChooserNativeDialogNew("Save",
				u.MainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
			defer dlg.Destroy()

			base := filepath.Base(m.Pages[m.PageIndex].FilePath)
			dlg.SetCurrentFolder(m.ExportDir)
			dlg.SetCurrentName(base)
			dlg.SetDoOverwriteConfirmation(true)

			output := dlg.NativeDialog.Run()
			if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
				f := dlg.GetFilename()
				m := &util.Message{TypeName: "exportPage", Data: f}
				u.SendMessage(*m)
			}
		case gdk.KEY_question, gdk.KEY_F1:
			dlg := gtk.MessageDialogNewWithMarkup(u.MainWindow,
				gtk.DialogFlags(gtk.DIALOG_MODAL),
				gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, "Help")
			defer dlg.Destroy()

			dlg.SetTitle("Help")
			dlg.SetMarkup(util.HELP_TXT)
			css, _ := dlg.GetStyleContext()
			css.AddClass("msg-dlg")

			dlg.Run()
		}
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
