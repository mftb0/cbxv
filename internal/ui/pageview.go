package ui

import (
	"fmt"
	"math"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"github.com/mftb0/cbxv-gotk3/internal/model"
	"github.com/mftb0/cbxv-gotk3/internal/util"
)

const (
    ALIGN_LEFT = iota
    ALIGN_RIGHT
    ALIGN_CENTER
)

const TICK = 3000

type PageView struct {
    ui                   *UI
    hud                  *gtk.Overlay
    hudHidden            bool
    hudKeepAlive         bool
    canvas               *gtk.DrawingArea
    keyPressSignalHandle *glib.SignalHandle
    hdrControl           *PageViewHdrControl
    navControl           *PageViewNavControl
}

func NewPageView(m *model.Model, u *UI, messenger util.Messenger) View {
    v := &PageView{}
    v.ui = u

    v.hud = v.newHUD(m, u)

    v.Connect(m, u)
    v.canvas, _ = gtk.DrawingAreaNew()
    v.hud.Add(v.canvas)
    v.initRenderer(m)
    v.hud.ShowAll()

    // DND
    target,_ := gtk.TargetEntryNew("text/uri-list", gtk.TargetFlags(0), 0)
    v.hud.DragDestSet(gtk.DEST_DEFAULT_ALL, []gtk.TargetEntry{*target}, gdk.ACTION_COPY)
    v.hud.Connect("drag-data-received", func(widget *gtk.Overlay, context *gdk.DragContext, x int, y int, selData *gtk.SelectionData) {
        if selData != nil {
            handleDropData(selData.GetData(), u.Commands.Names["openFile"])
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

    return v
}

func (v *PageView) Render(m *model.Model) {
    glib.IdleAdd(func() {
        v.renderHud(m)
    })
}

func (v *PageView) newHUD(m *model.Model, u *UI) *gtk.Overlay {
    o, _ := gtk.OverlayNew()

    v.hdrControl = NewHdrControl(m, u)
    v.navControl = NewNavControl(m, u)
    o.AddOverlay(v.hdrControl.container)
    o.AddOverlay(v.navControl.container)
    v.hudHidden = false

    return o
}

func (v *PageView) renderHud(m *model.Model) {
    v.hdrControl.Render(m)
    v.navControl.Render(m)
}

func (v *PageView) Connect(m *model.Model, u *UI) {
    u.MainWindow.Add(v.hud)
    sigH := u.MainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) bool {
        keyEvent := gdk.EventKeyNewFromEvent(event)
        keyVal := keyEvent.KeyVal()
        cmd := u.Commands.KeyCodes[keyVal]
        if cmd != nil {
            cmd.Execute()
        }

        v.hud.ShowAll()
        v.hudHidden = false
        v.hudKeepAlive = true
        return true
    })
    v.keyPressSignalHandle = &sigH
    u.MainWindow.ShowAll()
}

func (v *PageView) Disconnect(m *model.Model, u *UI) {
    if v.keyPressSignalHandle != nil {
        u.MainWindow.HandlerDisconnect(*v.keyPressSignalHandle)
        v.keyPressSignalHandle = nil
    }
    u.MainWindow.Remove(v.hud)
}

func (v *PageView) initRenderer(m *model.Model) {
    v.canvas.Connect("draw", func(canvas *gtk.DrawingArea, cr *cairo.Context) {
        cr.SetSourceRGB(0, 0, 0)
        cr.Rectangle(0, 0, float64(v.canvas.GetAllocatedWidth()), float64(v.canvas.GetAllocatedHeight()))
        cr.Fill()
        if m.Spreads == nil {
            return
        }

        spread := m.Spreads[m.SpreadIndex]
        if m.LayoutMode == model.TWO_PAGE {
            s := newTwoPageSpread(m, canvas, cr, spread)
            renderTwoPageSpread(s)
        } else if m.LayoutMode == model.ONE_PAGE {
            s := newOnePageSpread(canvas, cr, spread.Pages[0])
            renderOnePageSpread(s)
        }
        w := v.hud.GetAllocatedWidth() - 40
        v.hdrControl.container.SetSizeRequest(w, 8)
        v.navControl.container.SetSizeRequest(w, 8)
    })

    v.canvas.AddEvents(4)
    v.canvas.AddEvents(int(gdk.BUTTON_PRESS_MASK))
    v.canvas.Connect("event", func(canvas *gtk.DrawingArea, event *gdk.Event) bool {
        if v.hudHidden {
            //reset the hud hiding
            v.hdrControl.container.Show()
            v.navControl.container.Show()
            v.hudHidden = false
            v.hudKeepAlive = true
        }
        return false
    })

    v.canvas.Connect("button-press-event", func(canvas *gtk.DrawingArea, event *gdk.Event) bool {
        r := false
        w := v.hud.GetAllocatedWidth()
        half := float64(w / 2)
        e := &gdk.EventButton{Event: event}
        if e.X() < half {
            v.ui.Commands.Names["leftPage"].Execute()
            r = true
        } else {
            v.ui.Commands.Names["rightPage"].Execute()
            r = true
        }
        //reset the hud hiding
        v.hud.ShowAll()
        v.hudHidden = false
        v.hudKeepAlive = true
        return r
    })
}

type PagePosition int

type OnePageSpread struct {
    canvas *gtk.DrawingArea
    cr     *cairo.Context
    page   *model.Page
}

func newOnePageSpread(canvas *gtk.DrawingArea, cr *cairo.Context,
    page *model.Page) *OnePageSpread {
    return &OnePageSpread{canvas, cr, page}
}

type TwoPageSpread struct {
    canvas    *gtk.DrawingArea
    cr        *cairo.Context
    leftPage  *model.Page
    rightPage *model.Page
}

// Create a two pg spread accounting for direction
func newTwoPageSpread(m *model.Model, canvas *gtk.DrawingArea, cr *cairo.Context, spread *model.Spread) *TwoPageSpread {
    s := &TwoPageSpread{}
    s.canvas = canvas
    s.cr = cr
    if m.Direction == model.LTR {
        s.leftPage = spread.Pages[0]
        if len(spread.Pages) > 1 {
            s.rightPage = spread.Pages[1]
        }
    } else {
        if len(spread.Pages) > 1 {
            s.leftPage = spread.Pages[1]
            s.rightPage = spread.Pages[0]
        } else {
            s.leftPage = spread.Pages[0]
        }
    }

    return s
}

func scalePixbufToFit(canvas *gtk.DrawingArea, p *gdk.Pixbuf, w int, h int) (*gdk.Pixbuf, error) {
    cW := float64(w)
    cH := float64(h)
    pW := float64(p.GetWidth())
    pH := float64(p.GetHeight())
    var err error
    if pW > cW || pH > cH {
        scale := math.Min(cW/pW, cH/pH)
        p, err = p.ScaleSimple(int(pW*scale), int(pH*scale), gdk.INTERP_BILINEAR)
        if err != nil {
            return nil, err
        }
    } else {
        scale := math.Min(cW/pW, cH/pH)
        p, err = p.ScaleSimple(int(pW*scale), int(pH*scale), gdk.INTERP_BILINEAR)
        if err != nil {
            return nil, err
        }
    }
    return p, nil
}

func positionPixbuf(canvas *gtk.DrawingArea, p *gdk.Pixbuf, pos PagePosition) (x, y int) {
    var cW int
    if pos != ALIGN_CENTER {
        cW = canvas.GetAllocatedWidth() / 2
    } else {
        cW = canvas.GetAllocatedWidth()
    }
    cH := canvas.GetAllocatedHeight()
    pW := p.GetWidth()
    pH := p.GetHeight()

    if pos == ALIGN_CENTER {
        x = (cW - pW) / 2
    } else if pos == ALIGN_LEFT {
        x = cW
    } else {
        x = (cW - pW)
    }
    y = (cH - pH) / 2
    return x, y
}

func renderPixbuf(cr *cairo.Context, p *gdk.Pixbuf, x, y int) {
    gtk.GdkCairoSetSourcePixBuf(cr, p, float64(x), float64(y))
    cr.Paint()
}

func renderOnePageSpread(s *OnePageSpread) error {
    if s.page.Loaded == false {
        return fmt.Errorf("Image required by spread not loaded")
    }

    cW := s.canvas.GetAllocatedWidth()
    cH := s.canvas.GetAllocatedHeight()
    p, err := scalePixbufToFit(s.canvas, s.page.Image, cW, cH)
    if err != nil {
        return err
    }

    x, y := positionPixbuf(s.canvas, p, ALIGN_CENTER)

    renderPixbuf(s.cr, p, x, y)
    return nil
}

// direction (rtl or ltr) has already been accounted for
// so left and right here are literal
func renderTwoPageSpread(s *TwoPageSpread) error {
    if s.leftPage.Loaded == false {
        return fmt.Errorf("Image required by spread not loaded")
    }

    var x, y, cW, cH int
    if s.rightPage != nil {
        //put the left pg on the left, right-aligned
        cW = s.canvas.GetAllocatedWidth() / 2
        cH = s.canvas.GetAllocatedHeight()
        lp, err := scalePixbufToFit(s.canvas, s.leftPage.Image, cW, cH)
        if err != nil {
            return err
        }

        x, y = positionPixbuf(s.canvas, lp, ALIGN_RIGHT)
        renderPixbuf(s.cr, lp, x, y)

        //put the right pg on the right, left-aligned
        if s.rightPage.Loaded == false {
            return fmt.Errorf("Image required by spread not loaded")
        }

        rp, err := scalePixbufToFit(s.canvas, s.rightPage.Image, cW, cH)
        if err != nil {
            return err
        }

        x, y = positionPixbuf(s.canvas, rp, ALIGN_LEFT)
        renderPixbuf(s.cr, rp, x, y)
    } else {
        //there is no right page, then center the left page
        cW = s.canvas.GetAllocatedWidth()
        cH = s.canvas.GetAllocatedHeight()
        lp, err := scalePixbufToFit(s.canvas, s.leftPage.Image, cW, cH)
        if err != nil {
            return err
        }

        x, y = positionPixbuf(s.canvas, lp, ALIGN_CENTER)
        renderPixbuf(s.cr, lp, x, y)
    }
    return nil
}
