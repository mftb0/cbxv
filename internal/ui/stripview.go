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
	container   *gtk.Grid
	scrollbars  *gtk.ScrolledWindow
	canvas      []*gtk.DrawingArea
}

func NewStripView(m *model.Model, u *UI, messenger util.Messenger) View {
	v := &StripView{}
	v.sendMessage = messenger

	v.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)

	v.Init(m, u)


	return v
}

func (v *StripView) Init(m *model.Model, u *UI) {

    var err error
    v.container, err  = gtk.GridNew()
	if err != nil {
		fmt.Printf("Error creating container %s\n", err)
	}
	v.container.SetHExpand(true)

	for i := range m.Pages {
		page := m.Pages[i]
        if !page.Loaded {
            page.Load()
        }
        c, _ := gtk.DrawingAreaNew()
        v.canvas = append(v.canvas, c)
	    v.container.Attach(c, i, 0, 1, 1)
	}

	v.initRenderer(m)
	u.mainWindow.ShowAll()
}

func (v *StripView) Connect(m *model.Model, u *UI) {
//	u.mainWindow.Add(v.container)
}

func (v *StripView) Disconnect(m *model.Model, u *UI) {
	u.mainWindow.Remove(v.container)
}

func (v *StripView) Render(m *model.Model) {
	glib.IdleAdd(func() {
		v.RenderHud(m)
	})
}

func (v *StripView) RenderHud(m *model.Model) {
}

func (v *StripView) initRenderer(m *model.Model) {
    if m.LayoutMode == model.LONG_STRIP {
        if m.Spreads == nil {
            return
        }

        for i := range m.Spreads {
            v.canvas[i].Connect("draw", func(canvas *gtk.DrawingArea, cr *cairo.Context) {
                cr.SetSourceRGB(0, 0, 0)
                cr.Rectangle(0, 0, float64(v.canvas[i].GetAllocatedWidth()), float64(v.canvas[i].GetAllocatedHeight()))
                cr.Fill()
                spread := m.Spreads[i]
                s := newLongStripSpread(canvas, cr, spread)
                renderLongStripSpread(m, s)
                runtime.GC()
            })
        }
    }
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

func scalePixbufToWidth(canvas *gtk.DrawingArea, p *gdk.Pixbuf, w int) (*gdk.Pixbuf, error) {
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

func renderLongStripSpread(m *model.Model, s *LongStripSpread) error {
	var x, y int
    cW := s.canvas.GetAllocatedWidth()

    page := s.pages[0]
    p, err := scalePixbufToWidth(s.canvas, page.Image, cW)
    if err != nil {
        return err
    }
    x = positionLongStripPixbuf(s.canvas, p)
    renderPixbuf(s.cr, p, x, y)
    s.canvas.SetSizeRequest(x, y)

    return nil
}
