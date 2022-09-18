package ui

import (
	"fmt"
    "image"
	_ "image/color"
	"math"
	"path/filepath"
	"time"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"

    "example.com/cbxv-gotk3/internal/util"
    "example.com/cbxv-gotk3/internal/model"
)

const (
   LEFT_ALIGN = iota
   RIGHT_ALIGN
   CENTER
)

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

    u.hud = u.newHUD("")
    u.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)
    u.scrollbars.Add(u.hud)
	u.mainWindow.Add(u.scrollbars)

    u.initKBHandler(m)

    u.initCanvas(m)

    hudChan = make(chan bool)
    hudTicker = time.NewTicker(time.Second * 10)
    defer hudTicker.Stop()

    go hudHandler(u)

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
    hudChan <-false
}

func (u *UI) RunFunc(f interface{}) {
    glib.IdleAdd(f)
}

func (u *UI) Render(m *model.Model) {
    glib.IdleAdd(func(){
        u.renderHud(m)
        u.mainWindow.QueueDraw()
    })
}

func (u *UI) newHUD(title string) *gtk.Overlay {
	o, _ := gtk.OverlayNew()

    u.hdrControl = NewHdrControl()
    u.navControl = NewNavControl()
	o.AddOverlay(u.hdrControl.container)
	o.AddOverlay(u.navControl.container)
    u.hudHidden = false

    return o
}

func (u *UI) renderHud(m *model.Model) {
	u.hdrControl.Render(m)
	u.navControl.Render(m)
}

func (u *UI) initKBHandler(model *model.Model) {
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
            u.initCanvas(model)
            u.sendMessage(util.Message{TypeName: "setDisplayModeOnePage"})
        } else if keyVal == gdk.KEY_2 {
            u.initCanvas(model)
            u.sendMessage(util.Message{TypeName: "setDisplayModeTwoPage"})
        } else if keyVal == gdk.KEY_3 {
            u.sendMessage(util.Message{TypeName: "setDisplayModeLongStrip"})
        } else if keyVal == gdk.KEY_grave {
            u.sendMessage(util.Message{TypeName: "toggleReadMode"})
        } else if keyVal == gdk.KEY_f {
            if model.Fullscreen {
                u.mainWindow.Unfullscreen()
            } else {
                u.mainWindow.Fullscreen()
            }
            u.sendMessage(util.Message{TypeName: "toggleFullscreen"})
        } else if keyVal == gdk.KEY_o {
            dlg, _ := gtk.FileChooserNativeDialogNew("Open", u.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
            dlg.SetCurrentFolder(model.BrowseDirectory)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "openFile", Data: f}
                u.sendMessage(*m)
            } else {
            }
            u.initCanvas(model)
        } else if keyVal == gdk.KEY_c {
            u.sendMessage(util.Message{TypeName: "closeFile"})
            u.initCanvas(model)
        } else if keyVal == gdk.KEY_r {
            u.sendMessage(util.Message{TypeName: "reflow"})
        } else if keyVal == gdk.KEY_e {
            dlg, _ := gtk.FileChooserNativeDialogNew("Save", u.mainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
            base := filepath.Base(model.Pages[model.SelectedPage].FilePath)
            dlg.SetCurrentFolder(model.BrowseDirectory)
            dlg.SetCurrentName(base)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "exportFile", Data: f}
                u.sendMessage(*m)
            } else {
            }
        } else if keyVal == gdk.KEY_n {
            u.initCanvas(model)
            u.sendMessage(util.Message{TypeName: "nextFile"})
        } else if keyVal == gdk.KEY_p {
            u.sendMessage(util.Message{TypeName: "previousFile"})
            u.initCanvas(model)
        } else if keyVal == gdk.KEY_space {
            u.sendMessage(util.Message{TypeName: "toggleBookmark"})
        } else if keyVal == gdk.KEY_l {
            u.sendMessage(util.Message{TypeName: "lastBookmark"})
        }
        //reset the hud hiding
        hudChan <-true
        if u.hudHidden {
            u.hud.ShowAll()
            u.hudHidden = false
        }
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
            lo := newTwoPageLayout(m, canvas, cr, leaf)
            renderTwoPageLayout(lo)
        } else if m.LeafMode == model.ONE_PAGE {
            lo := newOnePageLayout(canvas, cr, leaf.Pages[0])
            renderOnePageLayout(lo)
        } else {
            lo := newLongStripLayout(canvas, cr, leaf)
            renderLongStripLayout(m, u, lo)
        }
        w := u.mainWindow.GetAllocatedWidth() - 40
        u.hdrControl.container.SetSizeRequest(w, 8)
        u.navControl.container.SetSizeRequest(w, 8)
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

func hudHandler(ui *UI) {
    for {
        select {
        case <-hudTicker.C:
            if !ui.hudHidden {
                glib.IdleAdd(func(){
                    ui.hdrControl.container.Hide()
                    ui.navControl.container.Hide()
                    ui.mainWindow.QueueDraw()
                })
                ui.hudHidden = true;
            }
        case r := <-hudChan:
            if r == true {
                hudTicker = time.NewTicker(time.Second * 10)
                ui.hudHidden = false
            } else {
                return
            }
        }
    }
}

// Ticker to hide the HUD
var hudTicker *time.Ticker
var hudChan chan bool

type PagePosition int

type OnePageLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    page *model.Page
}

func newOnePageLayout(canvas *gtk.DrawingArea, cr *cairo.Context,
    page *model.Page) *OnePageLayout {
    return &OnePageLayout{canvas, cr, page}
}

type TwoPageLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    leftPage *model.Page
    rightPage *model.Page
}

// Create a two pg layout accounting for readmode
func newTwoPageLayout(m *model.Model, canvas *gtk.DrawingArea, cr *cairo.Context, leaf *model.Leaf) *TwoPageLayout {
    lo := &TwoPageLayout{}
    lo.canvas = canvas
    lo.cr = cr
    if m.ReadMode == model.LTR {
        lo.leftPage = leaf.Pages[0]
        if(len(leaf.Pages) > 1) {
            lo.rightPage = leaf.Pages[1]
        }
    } else {
        if(len(leaf.Pages) > 1) {
            lo.leftPage = leaf.Pages[1]
            lo.rightPage = leaf.Pages[0]
        } else {
            lo.leftPage = leaf.Pages[0]
        }
    }

    return lo
}

type LongStripLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    pages []*model.Page
}

func newLongStripLayout(canvas *gtk.DrawingArea, cr *cairo.Context, leaf *model.Leaf) *LongStripLayout {
    lo := &LongStripLayout{}
    lo.canvas = canvas
    lo.cr = cr
    lo.pages = leaf.Pages
    return lo
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
    if pos != CENTER {
        cW = canvas.GetAllocatedWidth() / 2
    } else {
        cW = canvas.GetAllocatedWidth()
    }
    cH := canvas.GetAllocatedHeight()
    pW := p.GetWidth()
    pH := p.GetHeight()

    if pos == CENTER {
        x = (cW - pW) / 2
    } else if pos == LEFT_ALIGN {
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

func imageToPixbuf(picture image.Image) (*gdk.Pixbuf, error) {
	w := picture.Bounds().Max.X
	h := picture.Bounds().Max.Y
	pixbuf, err := gdk.PixbufNew(gdk.COLORSPACE_RGB, true, 8, w, h)
	if nil != err {
		return nil, err
	}
	pixels := pixbuf.GetPixels()

	const bytesPerPixel = 4
	i := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			colour := picture.At(x, y)
			r, g, b, a := colour.RGBA()

			pixels[i] = componentToByte(r)
			pixels[i+1] = componentToByte(g)
			pixels[i+2] = componentToByte(b)
			pixels[i+3] = componentToByte(a)

			i += bytesPerPixel
		}
	}

	return pixbuf, nil
}

func componentToByte(component uint32) byte {
    //256/65536
    const ratio = 0.00390625
	byteValue := ratio * float64(component)
	if byteValue > 255 {
		return byte(255)
	}
	return byte(byteValue)
}

func renderOnePageLayout(layout *OnePageLayout) error {
	p, _ := imageToPixbuf(*layout.page.Image)
    cW := layout.canvas.GetAllocatedWidth()
    cH := layout.canvas.GetAllocatedHeight()

    p, err := scalePixbufToFit(layout.canvas, p, cW, cH)
    if err != nil {
        return err
    }

    x, y := positionPixbuf(layout.canvas, p, CENTER)
    renderPixbuf(layout.cr, p, x, y)
    p = nil
    return nil
}

// readmode (rtl or ltr) has already been accounted for
// so left and right here are literal
func renderTwoPageLayout(layout *TwoPageLayout) error {
    var err error
	lp, _ := imageToPixbuf(*layout.leftPage.Image)

	var x, y, cW, cH int
    if layout.rightPage != nil {
	    //put the left pg on the left, right-aligned
		cW = layout.canvas.GetAllocatedWidth() / 2
		cH = layout.canvas.GetAllocatedHeight()
        lp, err = scalePixbufToFit(layout.canvas, lp, cW, cH)
		if err != nil {
			return err
		}
        x, y = positionPixbuf(layout.canvas, lp, RIGHT_ALIGN)
        renderPixbuf(layout.cr, lp, x, y)

	    //put the right pg on the right, left-aligned
		rp, _ := imageToPixbuf(*layout.rightPage.Image)
        rp, err := scalePixbufToFit(layout.canvas, rp, cW, cH)
        if err != nil {
            return err
        }
        x, y = positionPixbuf(layout.canvas, rp, LEFT_ALIGN)
        renderPixbuf(layout.cr, rp, x, y)
    } else {
	    //there is no right page, then center the left page
		cW = layout.canvas.GetAllocatedWidth()
		cH = layout.canvas.GetAllocatedHeight()
        lp, err = scalePixbufToFit(layout.canvas, lp, cW, cH)
		if err != nil {
			return err
		}
		x, y = positionPixbuf(layout.canvas, lp, CENTER)
        renderPixbuf(layout.cr, lp, x, y)
    }

    return nil
}

func renderLongStripLayout(m *model.Model, ui *UI, layout *LongStripLayout) error {
    var x, y int
    if ui.longStripRender == nil {
        cW := layout.canvas.GetAllocatedWidth()

        for i := range layout.pages {
            page := layout.pages[i]
	        p, _ := imageToPixbuf(*page.Image)
            p, err := scalePixbufToWidth(layout.canvas, p, cW)
            if err != nil {
                return err
            }
            x = positionLongStripPixbuf(layout.canvas, p)
            renderPixbuf(layout.cr, p, x, y)
            y += p.GetHeight()
            ui.longStripRender = append(ui.longStripRender, p)
        }
    } else {
        for i := range ui.longStripRender {
            p := ui.longStripRender[i]
            x = positionLongStripPixbuf(layout.canvas, p)
            renderPixbuf(layout.cr, p, x, y)
            y += p.GetHeight()
        }
    }
    layout.canvas.SetSizeRequest(x, y)

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

