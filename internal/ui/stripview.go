package ui

import (
	"fmt"
	_ "image/color"
	"math"
	"runtime"
	_ "runtime"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/mftb0/cbxv-gotk3/internal/model"
	"github.com/mftb0/cbxv-gotk3/internal/util"
)

type StripView struct {
	sendMessage          util.Messenger
	container            *gtk.Box
	scrollbars           *gtk.ScrolledWindow
	hud                  *gtk.Overlay
	hudHidden            bool
	hudKeepAlive         bool
	keyPressSignalHandle *glib.SignalHandle
	configSignalHandle   *glib.SignalHandle
	hdrControl           *StripViewHdrControl
	navControl           *StripViewNavControl
	width                int
}

func NewStripView(m *model.Model, u *UI, messenger util.Messenger) View {
	v := &StripView{}
	v.sendMessage = messenger

	v.hud = v.newHUD(m, u)
	v.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)

	var err error
	v.container, err = gtk.BoxNew(gtk.ORIENTATION_VERTICAL, 0)
	if err != nil {
		fmt.Printf("Error creating container %s\n", err)
	}
	v.container.SetHExpand(true)
	v.container.SetVExpand(true)

	v.scrollbars.Add(v.container)
	v.hud.Add(v.scrollbars)

	v.scrollbars.Connect("scroll-event", func() {
		v.hud.ShowAll()
		v.hudHidden = false
		v.hudKeepAlive = true
	})

    // DND
    target,_ := gtk.TargetEntryNew("text/uri-list", gtk.TargetFlags(0), 0)
    v.hud.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*target}, gdk.ACTION_COPY)
    v.hud.Connect("drag-data-received", func(widget *gtk.Overlay, context *gdk.DragContext, x int, y int, selData *gtk.SelectionData) {
        if selData != nil {
            util.HandleDropData(selData.GetData(), u.sendMessage)
        }
    })

	v.hudKeepAlive = false
	glib.TimeoutAdd(TICK, func() bool {
		if !v.hudHidden && !v.hudKeepAlive {
			v.hdrControl.container.Hide()
			v.navControl.container.Hide()
			u.mainWindow.QueueDraw()
			v.hudHidden = true
		} else {
			v.hudKeepAlive = false
		}
		return true
	})

    v.width = u.mainWindow.GetAllocatedWidth()
	return v
}

func (v *StripView) Connect(m *model.Model, u *UI) {
	kpsH := u.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		keyVal := keyEvent.KeyVal()
		switch keyVal {
		case gdk.KEY_w, gdk.KEY_Up, gdk.KEY_k:
			// scroll to top, no msg to send, doesn't affect model
			v.scrollbars.GetVAdjustment().SetValue(0)
		case gdk.KEY_s, gdk.KEY_Down, gdk.KEY_j:
			// scroll to bottom, no msg to send, doesn't affect model
			b := v.scrollbars.GetVAdjustment().GetUpper()
			v.scrollbars.GetVAdjustment().SetValue(b)
		case gdk.KEY_n:
			v.sendMessage(util.Message{TypeName: "nextFile"})
		case gdk.KEY_p:
			v.sendMessage(util.Message{TypeName: "previousFile"})
		}
		v.hud.ShowAll()
		v.hudHidden = false
		v.hudKeepAlive = true
	})
	v.keyPressSignalHandle = &kpsH

    confsH := u.mainWindow.Connect("configure-event", func(widget *gtk.Window, event *gdk.Event) {
		e := &gdk.EventConfigure{Event: event}

        if v.width == e.Width() {
            return
        }

		v.Render(m)
		v.width = e.Width()
	})
	v.configSignalHandle = &confsH

	u.mainWindow.Add(v.hud)
	v.container.ShowAll()
	v.scrollbars.ShowAll()
	u.mainWindow.ShowAll()
}

func (v *StripView) Disconnect(m *model.Model, u *UI) {
	if v.keyPressSignalHandle != nil {
		u.mainWindow.HandlerDisconnect(*v.keyPressSignalHandle)
        v.keyPressSignalHandle = nil
	}
	if v.configSignalHandle != nil {
		u.mainWindow.HandlerDisconnect(*v.configSignalHandle)
        v.configSignalHandle = nil
	}
	u.mainWindow.Remove(v.hud)
}

func (v *StripView) Render(m *model.Model) {
	glib.IdleAdd(func() {
		v.renderHud(m)
		v.renderSpreads(m)
	})
}

func (v *StripView) newHUD(m *model.Model, u *UI) *gtk.Overlay {
	o, _ := gtk.OverlayNew()

	v.hdrControl = NewStripViewHdrControl(m, u)
	v.navControl = NewStripViewNavControl(m, u)
	o.AddOverlay(v.hdrControl.container)
	o.AddOverlay(v.navControl.container)
	v.hudHidden = false

	return o
}

func (v *StripView) renderHud(m *model.Model) {
	v.hdrControl.Render(m)
	v.navControl.Render(m)
}

func (v *StripView) renderSpreads(m *model.Model) {
	if m.Spreads == nil {
		return
	}

	if len(m.Spreads[0].Pages) < len(m.Pages) {
		v.sendMessage(util.Message{TypeName: "loadAllPages"})
		return
	}

    v.container.GetChildren().FreeFull(func (item any) {
        v.container.Remove(item.(gtk.IWidget))
    })
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

	for i := range m.Spreads[0].Pages {
		page := m.Spreads[0].Pages[i]
		// fixme: Shouldn't be needed but actually
		// lower mem use
		p, _ := gdk.PixbufCopy(page.Image)
		p, _ = scalePixbufToWidth(p, v.width)
		c, _ := gtk.ImageNewFromPixbuf(p)
		v.container.PackStart(c, true, true, 0)
		v.scrollbars.ShowAll()
	}

	runtime.GC()
}

func scalePixbufToWidth(p *gdk.Pixbuf, w int) (*gdk.Pixbuf, error) {
	cW := float64(w)
	pW := float64(p.GetWidth())
	pH := float64(p.GetHeight())
	var err error

	if pW != cW {
		scale := cW / pW
        scale = math.Min(scale, 1.10)
		p, err = p.ScaleSimple(int(pW*scale), int(pH*scale), gdk.INTERP_BILINEAR)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}
