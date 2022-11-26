package ui

import (
	"fmt"
	"math"

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
	target, _ := gtk.TargetEntryNew("text/uri-list", gtk.TargetFlags(0), 0)
	v.hud.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*target}, gdk.ACTION_COPY)
	v.hud.Connect("drag-data-received", func(widget *gtk.Overlay, context *gdk.DragContext, x int, y int, selData *gtk.SelectionData) {
		if selData != nil {
			util.HandleDropData(selData.GetData(), u.SendMessage)
		}
	})

	v.hudKeepAlive = false
	glib.TimeoutAdd(TICK, func() bool {
		if !v.hudHidden && !v.hudKeepAlive {
			v.hdrControl.container.Hide()
			v.navControl.container.Hide()
			u.MainWindow.QueueDraw()
			v.hudHidden = true
		} else {
			v.hudKeepAlive = false
		}
		return true
	})

	v.width = u.MainWindow.GetAllocatedWidth()
	return v
}

func (v *StripView) Connect(m *model.Model, u *UI) {
	kpsH := u.MainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
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

	confsH := u.MainWindow.Connect("configure-event", func(widget *gtk.Window, event *gdk.Event) {
		e := &gdk.EventConfigure{Event: event}

		if v.width == e.Width() {
			return
		}

		v.Render(m)
		v.width = e.Width()
	})
	v.configSignalHandle = &confsH

	u.MainWindow.Add(v.hud)
	v.container.ShowAll()
	v.scrollbars.ShowAll()
	u.MainWindow.ShowAll()
}

func (v *StripView) Disconnect(m *model.Model, u *UI) {
	if v.keyPressSignalHandle != nil {
		u.MainWindow.HandlerDisconnect(*v.keyPressSignalHandle)
		v.keyPressSignalHandle = nil
	}
	if v.configSignalHandle != nil {
		u.MainWindow.HandlerDisconnect(*v.configSignalHandle)
		v.configSignalHandle = nil
	}
	u.MainWindow.Remove(v.hud)
}

func (v *StripView) Render(m *model.Model) {
	glib.IdleAdd(func() {
		v.renderHud(m)
		v.renderSpreads(m)
	})
}

func (v *StripView) ScrollToTop() {
    v.scrollbars.GetVAdjustment().SetValue(0)
}

func (v *StripView) ScrollToBottom() {
    b := v.scrollbars.GetVAdjustment().GetUpper()
    v.scrollbars.GetVAdjustment().SetValue(b)
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

	v.container.GetChildren().FreeFull(func(item any) {
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
        p, _ := v.scalePixbufToWidth(page.Image, v.width)
		c, _ := gtk.ImageNewFromPixbuf(p)
		v.container.PackStart(c, true, true, 0)
		v.scrollbars.ShowAll()
	}
}

// In strip mode it's possible for the images to be very tall
// Allow some scaling up, but with a couple constraints:
// gtk won't display anything greater than 32kx32k pix
// Height's max then is 32k
// Width we don't scroll so max is the lesser of 32k and window width
// Overall we don't want to scroll more than 2x
func (v *StripView) clampScale(scale float64, pW float64, pH float64) float64 {
    // clamp no greater than 32k pix
	maxH := float64(32000)

    // clamp no greater than 32k pix or window width
    maxW := math.Min(32000, float64(v.width))

    // clamp no greater than 2x
    maxFactor := math.Min(scale, float64(2))

    // try to find an acceptable factor
    for ; maxFactor >= .80; maxFactor -= float64(.20) {
        if maxFactor*pW < maxW && maxFactor*pH < maxH {
            return maxFactor
        }
    }

	// Never could find anything acceptable, just skip scaling
	return 1
}

func (v *StripView)scalePixbufToWidth(p *gdk.Pixbuf, w int) (*gdk.Pixbuf, error) {
	cW := float64(w)
	pW := float64(p.GetWidth())
	pH := float64(p.GetHeight())
	var err error

	if pW != cW {
		scale := cW / pW
		if scale > 1 {
			scale = v.clampScale(scale, pW, pH)
		}

		p, err = p.ScaleSimple(int(pW*scale), int(pH*scale), gdk.INTERP_BILINEAR)
		if err != nil {
			return nil, err
		}
	}

	return p, nil
}
