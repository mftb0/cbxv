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
	hdrControl           *StripViewHdrControl
	navControl           *StripViewNavControl
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

	return v
}

func (v *StripView) Connect(m *model.Model, u *UI) {
	sigH := u.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
		keyEvent := gdk.EventKeyNewFromEvent(event)
		keyVal := keyEvent.KeyVal()
		if keyVal == gdk.KEY_w {
			// scroll to top, no msg to send, doesn't affect model
			v.scrollbars.GetVAdjustment().SetValue(0)
		} else if keyVal == gdk.KEY_s {
			// scroll to bottom, no msg to send, doesn't affect model
			b := v.scrollbars.GetVAdjustment().GetUpper()
			v.scrollbars.GetVAdjustment().SetValue(b)
		} else if keyVal == gdk.KEY_n {
			v.sendMessage(util.Message{TypeName: "nextFile"})
		} else if keyVal == gdk.KEY_p {
			v.sendMessage(util.Message{TypeName: "previousFile"})
		}
		v.hud.ShowAll()
		v.hudHidden = false
		v.hudKeepAlive = true
	})
	v.keyPressSignalHandle = &sigH

	u.mainWindow.Add(v.hud)
	v.container.ShowAll()
	v.scrollbars.ShowAll()
	u.mainWindow.ShowAll()
}

func (v *StripView) Disconnect(m *model.Model, u *UI) {
	if v.keyPressSignalHandle != nil {
		u.mainWindow.HandlerDisconnect(*v.keyPressSignalHandle)
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

		// fixme: Shouldn't be needed but actually
		// lower mem use
		p, _ := gdk.PixbufCopy(page.Image)
		p, _ = scalePixbufToWidth(p, x)
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

	if pW > cW {
		scale := cW / pW
		p, err = p.ScaleSimple(int(pW*scale), int(pH*scale), gdk.INTERP_BILINEAR)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}
