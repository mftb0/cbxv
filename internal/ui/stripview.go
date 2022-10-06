package ui

import (
	"fmt"
	_ "image/color"
	"runtime"
	_ "runtime"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"example.com/cbxv-gotk3/internal/model"
	"example.com/cbxv-gotk3/internal/util"
)

type StripView struct {
	sendMessage util.Messenger
	container   *gtk.Box
	scrollbars  *gtk.ScrolledWindow
	canvas      []*gtk.DrawingArea
	keyPressSignalHandle *glib.SignalHandle
}

func NewStripView(m *model.Model, u *UI, messenger util.Messenger) View {
	v := &StripView{}
	v.sendMessage = messenger

	v.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)

    var err error
    v.container, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		fmt.Printf("Error creating container %s\n", err)
	}
	v.container.SetHExpand(true)
	v.container.SetVExpand(true)

    v.scrollbars.Add(v.container)

	return v
}

func (v *StripView) Connect(m *model.Model, u *UI) {
	sigH := u.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		keyVal := keyEvent.KeyVal()
		if keyVal == gdk.KEY_w {
			v.sendMessage(util.Message{TypeName: "top"})
		} else if keyVal == gdk.KEY_s {
			v.sendMessage(util.Message{TypeName: "bottom"})
		} else if keyVal == gdk.KEY_n {
			v.sendMessage(util.Message{TypeName: "nextFile"})
		} else if keyVal == gdk.KEY_p {
			v.sendMessage(util.Message{TypeName: "previousFile"})
		}

	})
	v.keyPressSignalHandle = &sigH

	u.mainWindow.Add(v.scrollbars)
    v.container.ShowAll()
    v.scrollbars.ShowAll()
	u.mainWindow.ShowAll()
    fmt.Printf("connect\n")
}

func (v *StripView) Disconnect(m *model.Model, u *UI) {
	if v.keyPressSignalHandle != nil {
		u.mainWindow.HandlerDisconnect(*v.keyPressSignalHandle)
	}
	u.mainWindow.Remove(v.scrollbars)
}

func (v *StripView) Render(m *model.Model) {
	glib.IdleAdd(func() {
		v.renderHud(m)
	    v.renderSpreads(m)
	})
}

func (v *StripView) renderHud(m *model.Model) {
}

func (v *StripView) renderSpreads(m *model.Model) {
    if m.Spreads == nil {
        return
    }

    v.scrollbars.Remove(v.container)
    var err error
    v.container, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		fmt.Printf("Error creating container %s\n", err)
	}
	v.container.SetHExpand(true)
	v.container.SetVExpand(true)
    v.scrollbars.Add(v.container)
    v.scrollbars.ShowAll()
    
    x := v.scrollbars.GetAllocatedWidth() 
    for _, page := range m.Spreads[0].Pages {
        if !page.Loaded {
            page.Load()
        }
        p, _ := scalePixbufToWidth(page.Image, x)
        c, _ := gtk.ImageNewFromPixbuf(p) 
        v.container.PackStart(c, false, false, 0)
        v.scrollbars.ShowAll()
    }

    runtime.GC()
}

func scalePixbufToWidth(p *gdk.Pixbuf, w int) (*gdk.Pixbuf, error) {
	cW := float64(w)
	pW := float64(p.GetWidth())
	pH := float64(p.GetHeight())
	var err error

    if pW > cW {
        scale := cW / pW
        p, err = p.ScaleSimple(int(pW*scale), int(pH*scale), gdk.INTERP_BILINEAR)
        if err != nil {
            return nil, err
        }
    }

	return p, nil
}

