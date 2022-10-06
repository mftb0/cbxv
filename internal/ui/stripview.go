package ui

import (
	"fmt"
	_ "image/color"
	"runtime"
	_ "runtime"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/cairo"
	_ "github.com/gotk3/gotk3/cairo"
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
    rendered    bool
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
	u.mainWindow.Add(v.scrollbars)
    v.container.ShowAll()
    v.scrollbars.ShowAll()
	u.mainWindow.ShowAll()
    fmt.Printf("connect\n")
}

func (v *StripView) Disconnect(m *model.Model, u *UI) {
	u.mainWindow.Remove(v.scrollbars)
}

func (v *StripView) Render(m *model.Model) {
	glib.IdleAdd(func() {
		v.RenderHud(m)
	    v.renderSpreads(m)
	})
}

func (v *StripView) RenderHud(m *model.Model) {
}

func (v *StripView) renderSpreads(m *model.Model) {
    fmt.Printf("strip\n")
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

type LongStripSpread struct {
	canvas *gtk.DrawingArea
	cr     *cairo.Context
	pages  []*model.Page
}

func newLongStripSpread(canvas *gtk.DrawingArea, cr *cairo.Context, spread *model.Spread) *LongStripSpread {
	s := &LongStripSpread{}
	s.canvas = canvas
	s.cr = cr
	s.pages = spread.Pages
	return s
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

func positionLongStripPixbuf(canvas *gtk.DrawingArea, p *gdk.Pixbuf) (x int) {
	cW := canvas.GetAllocatedWidth()
	pW := p.GetWidth()
	x = (cW - pW) / 2
	return x
}


