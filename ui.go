package main

import (
	"fmt"
	_ "image/color"
	"math"
	_ "path/filepath"
	_ "strings"
	_"time"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	gotk3extra "github.com/jamesrr39/go-gtk-extra/gotk3-extra"
)

const (
   LEFT_ALIGN = iota
   RIGHT_ALIGN
   CENTER
)

type PagePosition int

func NewHeaderControl(ui *UI, title string)  {
}

func NewNavBar(model *Model)  {
}

func NewNavControl() {
}

func NewHUD(ui *UI, title string) *gtk.Overlay {
	o, _ := gtk.OverlayNew()

	pn, err := gtk.LabelNew("0")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	pn.SetHAlign(gtk.ALIGN_END)
	pn.SetVAlign(gtk.ALIGN_END)
	o.AddOverlay(pn)

    return o
}

type UI struct {
    mainWindow *gtk.Window
    hud *gtk.Overlay
    spread int
    canvas *gtk.DrawingArea
    scrollbars *gtk.ScrolledWindow
    view int
    headerControl int
    navControl int
    longStripRender []*gdk.Pixbuf
}

func InitKBHandler(model *Model, ui *UI) {
    ui.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
        keyEvent := gdk.EventKeyNewFromEvent(event)
        keyVal := keyEvent.KeyVal()
        if keyVal == gdk.KEY_q {
            m := &Message{typeName: "quit"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_d {
            m := &Message{typeName: "nextPage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_a {
            m := &Message{typeName: "previousPage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_w {
            m := &Message{typeName: "firstPage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_s {
            m := &Message{typeName: "lastPage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_Tab {
            m := &Message{typeName: "selectPage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_1 {
            InitCanvas(model, ui)
            m := &Message{typeName: "setDisplayModeOnePage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_2 {
            InitCanvas(model, ui)
            m := &Message{typeName: "setDisplayModeTwoPage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_3 {
            m := &Message{typeName: "setDisplayModeLongStrip"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_grave {
            m := &Message{typeName: "toggleReadMode"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_f {
            if model.fullscreen {
                ui.mainWindow.Unfullscreen()
            } else {
                ui.mainWindow.Fullscreen()
            }
            m := &Message{typeName: "toggleFullscreen"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_o {
            dlg, _ := gtk.FileChooserNativeDialogNew("Open", ui.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
            dlg.SetCurrentFolder(model.browseDirectory)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                d := dlg
                f := d.GetFilename()
                m := &Message{typeName: "openFile", data: f}
                sendMessage(*m)
            } else {
            }
            InitCanvas(model, ui)
        } else if keyVal == gdk.KEY_c {
            m := &Message{typeName: "closeFile"}
            sendMessage(*m)
            InitCanvas(model, ui)
        } else if keyVal == gdk.KEY_r {
            m := &Message{typeName: "reflow"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_e {
            m := &Message{typeName: "exportFile"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_n {
            InitCanvas(model, ui)
            m := &Message{typeName: "nextFile"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_p {
            m := &Message{typeName: "previousFile"}
            sendMessage(*m)
            InitCanvas(model, ui)
        } else if keyVal == gdk.KEY_e {
            m := &Message{typeName: "toggleBookmark"}
            sendMessage(*m)
        }
        //reset the hud hiding
//        hudTicker.Reset(time.Second * 5)
//        if ui.hud.Hidden {
//            ui.hud.Show()
//        }
     })
}

func renderNavControl(model *Model, ui *UI) {
}

func renderBookmark(model *Model, ui *UI, pageIndex int, pos PagePosition) {
}

func renderHeaderControl(model *Model, ui *UI) {
}

type OnePageLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    page *Page
}

func NewOnePageLayout(canvas *gtk.DrawingArea, cr *cairo.Context,
    page *Page) *OnePageLayout {
    return &OnePageLayout{canvas, cr, page}
}

type TwoPageLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    leftPage *Page
    rightPage *Page
}

// Create a two pg layout accounting for readmode
func NewTwoPageLayout(model *Model, canvas *gtk.DrawingArea, cr *cairo.Context, leaf *Leaf) *TwoPageLayout {

    lo := &TwoPageLayout{}
    lo.canvas = canvas
    lo.cr = cr
    if model.readMode == LTR {
        lo.leftPage = leaf.pages[0]
        if(len(leaf.pages) > 1) {
            lo.rightPage = leaf.pages[1]
        }
    } else {
        if(len(leaf.pages) > 1) {
            lo.leftPage = leaf.pages[1]
            lo.rightPage = leaf.pages[0]
        } else {
            lo.leftPage = leaf.pages[0]
        }
    }

    return lo
}

type LongStripLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    pages []*Page
}

func NewLongStripLayout(canvas *gtk.DrawingArea, cr *cairo.Context, leaf *Leaf) *LongStripLayout {

    lo := &LongStripLayout{}
    lo.canvas = canvas
    lo.cr = cr
    lo.pages = leaf.pages
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

func renderOnePageLayout(layout *OnePageLayout) error {
	p, _ := gotk3extra.PixBufFromImage(*layout.page.Image)
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
	lp, _ := gotk3extra.PixBufFromImage(*layout.leftPage.Image)

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
		rp, _ := gotk3extra.PixBufFromImage(*layout.rightPage.Image)
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

func renderLongStripLayout(model *Model, ui *UI, layout *LongStripLayout) error {
    var x, y int
    if ui.longStripRender == nil {
        cW := layout.canvas.GetAllocatedWidth()

        for i := range layout.pages {
            page := layout.pages[i]
	        p, _ := gotk3extra.PixBufFromImage(*page.Image)
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

func InitRenderer(model *Model, ui *UI) {
    ui.longStripRender = nil
    ui.canvas.Connect("draw", func(canvas *gtk.DrawingArea, cr *cairo.Context) {
        cr.SetSourceRGB(0,0,0)
        cr.Rectangle(0,0,float64(ui.canvas.GetAllocatedWidth()), float64(ui.canvas.GetAllocatedHeight()))
        cr.Fill()
        if model.leaves == nil {
            return
        }

        leaf := model.leaves[model.currentLeaf]
        if(model.leafMode == TWO_PAGE) {
            lo := NewTwoPageLayout(model, canvas, cr, leaf)
            renderTwoPageLayout(lo)
        } else if model.leafMode == ONE_PAGE {
            lo := NewOnePageLayout(canvas, cr, leaf.pages[0])
            renderOnePageLayout(lo)
        } else {
            lo := NewLongStripLayout(canvas, cr, leaf)
            renderLongStripLayout(model, ui, lo)
        }

        renderHeaderControl(model, ui)
    })
}

func InitCanvas(model *Model, ui *UI) {
    if ui.canvas != nil {
        ui.hud.Remove(ui.canvas)
        ui.canvas.Destroy()
        ui.canvas = nil
    }

    ui.canvas, _ = gtk.DrawingAreaNew()
	ui.hud.Add(ui.canvas)
    InitRenderer(model, ui)
    ui.mainWindow.ShowAll()
}

func InitCss() {
	css, err := gtk.CssProviderNew()
    if err != nil {
		fmt.Printf("css error %s\n", err)
	}

    data, err := loadTextFile("assets/index.css")
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

func InitUI(model *Model, ui *UI) {
    gtk.Init(nil)
    ui.mainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    ui.mainWindow.SetPosition(gtk.WIN_POS_CENTER)
    ui.mainWindow.SetTitle("cbxv")
    ui.mainWindow.Connect("destroy", func() {
        gtk.MainQuit()
    })
    ui.mainWindow.SetSizeRequest(1024, 768)

    InitCss()

    ui.hud = NewHUD(ui, "")
    ui.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)
    ui.scrollbars.Add(ui.hud)
	ui.mainWindow.Add(ui.scrollbars)

    InitKBHandler(model, ui)

    InitCanvas(model, ui)
}

