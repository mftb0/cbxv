package ui

import (
	"fmt"
	_ "image/color"
	"path/filepath"

	_ "golang.org/x/image/colornames"

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
    sendMessage util.Messenger
    mainWindow  *gtk.Window
    pageView    View
    stripView   View
    View        View
}

func NewUI(m *model.Model, messenger util.Messenger) *UI {
    gtk.Init(nil)
    u := &UI{}
    u.sendMessage = messenger
    u.mainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    u.mainWindow.SetPosition(gtk.WIN_POS_CENTER)
    u.mainWindow.SetTitle(m.ProgramName)
    u.mainWindow.Connect("destroy", func() {
        gtk.MainQuit()
    })
    u.mainWindow.SetSizeRequest(1024, 768)

    iPath, _ := util.AppIconPath()
    if iPath != nil {
	fmt.Printf("iconPath:%s\n", iPath)
        u.mainWindow.SetIconFromFile(*iPath)
    }

    if m.LayoutMode != model.LONG_STRIP {
        u.pageView = NewPageView(m, u, messenger)
        u.View = u.pageView
    }

    initCss()

    u.initKBHandler(m)

    u.mainWindow.ShowAll()

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
        u.mainWindow.QueueDraw()
    })
}

func (u *UI) initKBHandler(m *model.Model) {
    u.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
        keyEvent := gdk.EventKeyNewFromEvent(event)
        keyVal := keyEvent.KeyVal()
        switch keyVal {
        case gdk.KEY_q:
            u.sendMessage(util.Message{TypeName: "quit"})
            u.Quit()
        case gdk.KEY_1: 
            u.View.Disconnect(m, u)
            u.View = u.pageView
            u.View.Connect(m, u)
            u.sendMessage(util.Message{TypeName: "setDisplayModeOnePage"})
        case gdk.KEY_2:
            u.View.Disconnect(m, u)
            u.View = u.pageView
            u.View.Connect(m, u)
            u.sendMessage(util.Message{TypeName: "setDisplayModeTwoPage"})
        case gdk.KEY_3:
            u.View.Disconnect(m, u)
            if u.stripView == nil {
                u.stripView = NewStripView(m, u, u.sendMessage)
            }
            u.View = u.stripView
            u.View.Connect(m, u)
            u.sendMessage(util.Message{TypeName: "setDisplayModeLongStrip"})
        case gdk.KEY_f:
            if m.Fullscreen {
                u.mainWindow.Unfullscreen()
            } else {
                u.mainWindow.Fullscreen()
            }
            u.sendMessage(util.Message{TypeName: "toggleFullscreen"})
        case gdk.KEY_o:
            dlg, _ := gtk.FileChooserNativeDialogNew("Open", u.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
            dlg.SetCurrentFolder(m.BrowseDir)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "openFile", Data: f}
                u.sendMessage(*m)
            }
        case gdk.KEY_c:
            u.sendMessage(util.Message{TypeName: "closeFile"})
        case gdk.KEY_e:
            dlg, _ := gtk.FileChooserNativeDialogNew("Save", u.mainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
            base := filepath.Base(m.Pages[m.PageIndex].FilePath)
            dlg.SetCurrentFolder(m.ExportDir)
            dlg.SetCurrentName(base)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "exportFile", Data: f}
                u.sendMessage(*m)
            }
        case gdk.KEY_question:
            dlg := gtk.MessageDialogNewWithMarkup(u.mainWindow,
                gtk.DialogFlags(gtk.DIALOG_MODAL),
                gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, "Help")
            dlg.SetTitle("Help")
            dlg.SetMarkup(util.HELP_TXT)
            css, _ := dlg.GetStyleContext()
            css.AddClass("msg-dlg")
            dlg.Run()
            dlg.Destroy()
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
