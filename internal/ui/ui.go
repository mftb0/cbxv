package ui

import (
	"fmt"
	_ "image/color"
	"math"
	"path/filepath"
	"runtime"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

	"example.com/cbxv-gotk3/internal/model"
	"example.com/cbxv-gotk3/internal/util"
)

const (
   ALIGN_LEFT = iota
   ALIGN_RIGHT
   ALIGN_CENTER
)

const TICK = 5000

type UI struct {
    sendMessage util.Messenger
    mainWindow *gtk.Window
    hud *gtk.Overlay
    hudHidden bool
    canvas *gtk.DrawingArea
    scrollbars *gtk.ScrolledWindow
    hdrControl *HdrControl
    navControl *NavControl
    longStripRender []*gdk.Pixbuf
    hudKeepAlive bool
}

func NewUI(m *model.Model, messenger util.Messenger) *UI {
    gtk.Init(nil)
    u := &UI{}
    u.sendMessage = messenger
    u.mainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    u.mainWindow.SetPosition(gtk.WIN_POS_CENTER)
    u.mainWindow.SetTitle("cbxv")
    u.mainWindow.Connect("destroy", func() {
        gtk.MainQuit()
    })
    u.mainWindow.SetSizeRequest(1024, 768)

    u.mainWindow.Connect("configure-event", func() {
        w := u.mainWindow.GetAllocatedWidth() - 40
        u.hdrControl.container.SetSizeRequest(w, 8)
        u.navControl.container.SetSizeRequest(w, 8)
    })

    initCss()

    u.hud = u.newHUD(m, "")
    u.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)
    u.scrollbars.Add(u.hud)
	u.mainWindow.Add(u.scrollbars)

    u.initKBHandler(m)

    u.initCanvas(m)

    u.hudKeepAlive = false
    glib.TimeoutAdd(TICK, func () bool {
        if !u.hudHidden && !u.hudKeepAlive {
            u.hdrControl.container.Hide()
            u.navControl.container.Hide()
            u.mainWindow.QueueDraw()
            u.hudHidden = true;
        } else {
            u.hudKeepAlive = false
        }
        return true
    })

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
    glib.IdleAdd(func(){
        u.renderHud(m)
        // causes the draw event to fire
        // which gets the canvas to Render
        // see initRenderer
        u.mainWindow.QueueDraw()
    })
}

func (u *UI) newHUD(m *model.Model, title string) *gtk.Overlay {
	o, _ := gtk.OverlayNew()

    u.hdrControl = NewHdrControl(m, u)
    u.navControl = NewNavControl(m, u)
	o.AddOverlay(u.hdrControl.container)
	o.AddOverlay(u.navControl.container)
    u.hudHidden = false

    return o
}

func (u *UI) renderHud(m *model.Model) {
	u.hdrControl.Render(m)
	u.navControl.Render(m)
}

func (u *UI) initKBHandler(m *model.Model) {
    u.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
        keyEvent := gdk.EventKeyNewFromEvent(event)
        keyVal := keyEvent.KeyVal()
        if keyVal == gdk.KEY_q {
            u.sendMessage(util.Message{TypeName: "quit"})
            u.Quit()
        } else if keyVal == gdk.KEY_d {
            u.sendMessage(util.Message{TypeName: "nextPage"})
        } else if keyVal == gdk.KEY_a {
            u.sendMessage(util.Message{TypeName: "previousPage"})
        } else if keyVal == gdk.KEY_w {
            u.sendMessage(util.Message{TypeName: "firstPage"})
        } else if keyVal == gdk.KEY_s {
            u.sendMessage(util.Message{TypeName: "lastPage"})
        } else if keyVal == gdk.KEY_Tab {
            u.sendMessage(util.Message{TypeName: "selectPage"})
        } else if keyVal == gdk.KEY_1 {
            u.initCanvas(m)
            u.sendMessage(util.Message{TypeName: "setDisplayModeOnePage"})
        } else if keyVal == gdk.KEY_2 {
            u.initCanvas(m)
            u.sendMessage(util.Message{TypeName: "setDisplayModeTwoPage"})
        } else if keyVal == gdk.KEY_3 {
            u.sendMessage(util.Message{TypeName: "setDisplayModeLongStrip"})
        } else if keyVal == gdk.KEY_grave {
            u.sendMessage(util.Message{TypeName: "toggleReadMode"})
        } else if keyVal == gdk.KEY_f {
            if m.Fullscreen {
                u.mainWindow.Unfullscreen()
            } else {
                u.mainWindow.Fullscreen()
            }
            u.sendMessage(util.Message{TypeName: "toggleFullscreen"})
        } else if keyVal == gdk.KEY_o {
            dlg, _ := gtk.FileChooserNativeDialogNew("Open", u.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
            dlg.SetCurrentFolder(m.BrowseDirectory)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "openFile", Data: f}
                u.sendMessage(*m)
            }
            u.initCanvas(m)
        } else if keyVal == gdk.KEY_c {
            u.sendMessage(util.Message{TypeName: "closeFile"})
            u.initCanvas(m)
        } else if keyVal == gdk.KEY_r {
            u.sendMessage(util.Message{TypeName: "toggleSpread"})
        } else if keyVal == gdk.KEY_e {
            dlg, _ := gtk.FileChooserNativeDialogNew("Save", u.mainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
            base := filepath.Base(m.Pages[m.SelectedPage].FilePath)
            dlg.SetCurrentFolder(m.BrowseDirectory)
            dlg.SetCurrentName(base)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "exportFile", Data: f}
                u.sendMessage(*m)
            } else {
            }
        } else if keyVal == gdk.KEY_n {
            u.initCanvas(m)
            u.sendMessage(util.Message{TypeName: "nextFile"})
        } else if keyVal == gdk.KEY_p {
            u.sendMessage(util.Message{TypeName: "previousFile"})
            u.initCanvas(m)
        } else if keyVal == gdk.KEY_space {
            u.sendMessage(util.Message{TypeName: "toggleBookmark"})
        } else if keyVal == gdk.KEY_l {
            u.sendMessage(util.Message{TypeName: "lastBookmark"})
        } else if keyVal == gdk.KEY_question {
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

        u.hud.ShowAll()
        u.hudHidden = false
        u.hudKeepAlive = true
     })
}

func (u *UI) initRenderer(m *model.Model) {
    u.longStripRender = nil
    u.canvas.Connect("draw", func(canvas *gtk.DrawingArea, cr *cairo.Context) {
        cr.SetSourceRGB(0,0,0)
        cr.Rectangle(0,0,float64(u.canvas.GetAllocatedWidth()), float64(u.canvas.GetAllocatedHeight()))
        cr.Fill()
        if m.Leaves == nil {
            return
        }

        leaf := m.Leaves[m.CurrentLeaf]
        if(m.LeafMode == model.TWO_PAGE) {
            lo := newTwoPageSpread(m, canvas, cr, leaf)
            renderTwoPageSpread(lo)
        } else if m.LeafMode == model.ONE_PAGE {
            lo := newOnePageSpread(canvas, cr, leaf.Pages[0])
            renderOnePageSpread(lo)
        } else {
            lo := newLongStripSpread(canvas, cr, leaf)
            renderLongStripSpread(m, u, lo)
        }
        w := u.mainWindow.GetAllocatedWidth() - 40
        u.hdrControl.container.SetSizeRequest(w, 8)
        u.navControl.container.SetSizeRequest(w, 8)
        runtime.GC()
    })

    u.canvas.AddEvents(4)
    u.canvas.AddEvents(int(gdk.BUTTON_PRESS_MASK))
    u.canvas.Connect("event", func(canvas *gtk.DrawingArea, event *gdk.Event) bool {
        //reset the hud hiding
        u.hdrControl.container.Show()
        u.navControl.container.Show()
        u.hudHidden = false
        u.hudKeepAlive = true
        return false
    })

    u.canvas.Connect("button-press-event", func(canvas *gtk.DrawingArea, event *gdk.Event) {
        w := u.mainWindow.GetAllocatedWidth()
        half := float64(w/2)
        e := &gdk.EventButton{Event:event}
        // fixme: Don't have to deal w/rtl here because it's dealt with
        // in the app
        if e.X() < half {
            u.sendMessage(util.Message{TypeName: "previousPage"})
        } else {
            u.sendMessage(util.Message{TypeName: "nextPage"})
        }
        //reset the hud hiding
        u.hud.ShowAll()
        u.hudHidden = false
        u.hudKeepAlive = true
    })
}

func (u *UI) initCanvas(m *model.Model) {
    if u.canvas != nil {
        u.hud.Remove(u.canvas)
        u.canvas.Destroy()
        u.canvas = nil
    }

    u.canvas, _ = gtk.DrawingAreaNew()
	u.hud.Add(u.canvas)
    u.initRenderer(m)
    u.mainWindow.ShowAll()
}

type PagePosition int

type OnePageSpread struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    page *model.Page
}

func newOnePageSpread(canvas *gtk.DrawingArea, cr *cairo.Context,
    page *model.Page) *OnePageSpread {
    return &OnePageSpread{canvas, cr, page}
}

type TwoPageSpread struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    leftPage *model.Page
    rightPage *model.Page
}

// Create a two pg spread accounting for readmode
func newTwoPageSpread(m *model.Model, canvas *gtk.DrawingArea, cr *cairo.Context, leaf *model.Leaf) *TwoPageSpread {
    s := &TwoPageSpread{}
    s.canvas = canvas
    s.cr = cr
    if m.ReadMode == model.LTR {
        s.leftPage = leaf.Pages[0]
        if(len(leaf.Pages) > 1) {
            s.rightPage = leaf.Pages[1]
        }
    } else {
        if(len(leaf.Pages) > 1) {
            s.leftPage = leaf.Pages[1]
            s.rightPage = leaf.Pages[0]
        } else {
            s.leftPage = leaf.Pages[0]
        }
    }

    return s
}

type LongStripSpread struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    pages []*model.Page
}

func newLongStripSpread(canvas *gtk.DrawingArea, cr *cairo.Context, leaf *model.Leaf) *LongStripSpread {
    s := &LongStripSpread{}
    s.canvas = canvas
    s.cr = cr
    s.pages = leaf.Pages
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
        p, err = p.ScaleSimple(int(pW * scale), int(pH * scale), gdk.INTERP_BILINEAR)
        if err != nil {
            return nil, err
        }
    } else {
        scale := math.Min(cW/pW, cH/pH)
        p, err = p.ScaleSimple(int(pW * scale), int(pH * scale), gdk.INTERP_BILINEAR)
        if err != nil {
            return nil, err
        }
    }
    return p, nil
}

func scalePixbufToWidth(canvas *gtk.DrawingArea, p *gdk.Pixbuf, w int) (*gdk.Pixbuf, error) {
    cW := float64(w)
    pW := float64(p.GetWidth())
    pH := float64(p.GetHeight())
    var err error
    if pW > cW {
        scale := cW/pW
        p, err = p.ScaleSimple(int(pW * scale), int(pH * scale), gdk.INTERP_BILINEAR)
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

func positionLongStripPixbuf(canvas *gtk.DrawingArea, p *gdk.Pixbuf) (x int) {
    cW := canvas.GetAllocatedWidth()
    pW := p.GetWidth()
    x = (cW - pW) / 2
    return x
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
    // fixme: Dramatically reduces memory use ~70%, makes little sense but
    // possible explanation, some gtk methods are refusing to unref, possible
    // thread-safety issue
    p,_ := gdk.PixbufCopy(s.page.Image)
    p, err := scalePixbufToFit(s.canvas, s.page.Image, cW, cH)
    if err != nil {
        return err
    }

    x, y := positionPixbuf(s.canvas, p, ALIGN_CENTER)

    renderPixbuf(s.cr, p, x, y)
    return nil
}

// readmode (rtl or ltr) has already been accounted for
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
        // fixme: Dramatically reduces memory use ~70%
        lp,_ := gdk.PixbufCopy(s.leftPage.Image)
        lp, err := scalePixbufToFit(s.canvas, lp, cW, cH)
        if err != nil {
            return err
        }

        x, y = positionPixbuf(s.canvas, lp, ALIGN_RIGHT)
        renderPixbuf(s.cr, lp, x, y)

	    //put the right pg on the right, left-aligned
        if s.rightPage.Loaded == false {
            return fmt.Errorf("Image required by spread not loaded")
        }

        // fixme: Dramatically reduces memory use ~70%
        rp,_ := gdk.PixbufCopy(s.rightPage.Image)
        rp, err = scalePixbufToFit(s.canvas, rp, cW, cH)
        if err != nil {
            return err
        }

        x, y = positionPixbuf(s.canvas, rp, ALIGN_LEFT)
        renderPixbuf(s.cr, rp, x, y)
    } else {
	    //there is no right page, then center the left page
		cW = s.canvas.GetAllocatedWidth()
		cH = s.canvas.GetAllocatedHeight()
        // fixme: Dramatically reduces memory use ~70%
        lp,_ := gdk.PixbufCopy(s.leftPage.Image)
        lp, err := scalePixbufToFit(s.canvas, s.leftPage.Image, cW, cH)
        if err != nil {
            return err
        }

		x, y = positionPixbuf(s.canvas, lp, ALIGN_CENTER)
        renderPixbuf(s.cr, lp, x, y)
    }
    return nil
}

func renderLongStripSpread(m *model.Model, u *UI, s *LongStripSpread) error {
    var x, y int
    if u.longStripRender == nil {
        cW := s.canvas.GetAllocatedWidth()

        for i := range s.pages {
            page := s.pages[i]
            p, err := scalePixbufToWidth(s.canvas, page.Image, cW)
            if err != nil {
                return err
            }
            x = positionLongStripPixbuf(s.canvas, p)
            renderPixbuf(s.cr, p, x, y)
            y += p.GetHeight()
            u.longStripRender = append(u.longStripRender, p)
        }
    } else {
        for i := range u.longStripRender {
            p := u.longStripRender[i]
            x = positionLongStripPixbuf(s.canvas, p)
            renderPixbuf(s.cr, p, x, y)
            y += p.GetHeight()
        }
    }
    s.canvas.SetSizeRequest(x, y)

    return nil
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

