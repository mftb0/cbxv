package main

import (
	_ "image/color"
	"math"
	_ "path/filepath"
	_ "strings"
	"time"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
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
	b, _ := gtk.LabelNew("0")
	b.SetHAlign(gtk.ALIGN_END)
	b.SetVAlign(gtk.ALIGN_END)
	o.AddOverlay(b)

    return o
}

type UI struct {
    mainWindow *gtk.Window
    hud *gtk.Overlay
    spread int
    canvas *gtk.DrawingArea
    view int
    headerControl int
    navControl int
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
            m := &Message{typeName: "setDisplayModeOnePage"}
            sendMessage(*m)
        } else if keyVal == gdk.KEY_2 {
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
        hudTicker.Reset(time.Second * 5)
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

func NewTwoPageLayout(model *Model, canvas *gtk.DrawingArea, cr *cairo.Context,
    leaf *Leaf) *TwoPageLayout {

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

func renderPixbuf(cr *cairo.Context, p *gdk.Pixbuf, x, y int) {
    gtk.GdkCairoSetSourcePixBuf(cr, p, float64(x), float64(y))
    cr.Paint()
}

func renderOnePageLayout(layout *OnePageLayout) error {
    p := layout.page.Image.pixbuf
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

func renderTwoPageLayout(layout *TwoPageLayout) error {
    var err error
	lp := layout.leftPage.Image.pixbuf

	//put the left pg on the left, right-aligned unless 
	//there is no right page, then center the left page
	var x, y, cW, cH int
    if layout.rightPage != nil {
		cW = layout.canvas.GetAllocatedWidth() / 2
		cH = layout.canvas.GetAllocatedHeight()
        lp, err = scalePixbufToFit(layout.canvas, lp, cW, cH)
		if err != nil {
			return err
		}
        x, y = positionPixbuf(layout.canvas, lp, RIGHT_ALIGN)
    } else {
		cW = layout.canvas.GetAllocatedWidth()
		cH = layout.canvas.GetAllocatedHeight()
        lp, err = scalePixbufToFit(layout.canvas, lp, cW, cH)
		if err != nil {
			return err
		}
		x, y = positionPixbuf(layout.canvas, lp, CENTER)
    }
    renderPixbuf(layout.cr, lp, x, y)

    if layout.rightPage != nil {
		rp := layout.rightPage.Image.pixbuf

        rp, err := scalePixbufToFit(layout.canvas, rp, cW, cH)
        if err != nil {
            return err
        }

        x, y = positionPixbuf(layout.canvas, rp, LEFT_ALIGN)
        renderPixbuf(layout.cr, rp, x, y)
    }
    return nil
}

func InitRenderer(model *Model, ui *UI) {
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
            lo.cr = nil
        } else if model.leafMode == ONE_PAGE {
            lo := NewOnePageLayout(canvas, cr, leaf.pages[0])
            renderOnePageLayout(lo)
        } else {
        }

        renderHeaderControl(model, ui)
    })
}

func InitCanvas(model *Model, ui *UI) {
    if ui.canvas != nil {
        ui.hud.Remove(ui.canvas)
        ui.canvas.Unref()
        ui.canvas.Destroy()
        ui.canvas = nil
    }

    ui.canvas, _ = gtk.DrawingAreaNew()
	ui.hud.Add(ui.canvas)
    InitRenderer(model, ui)
    ui.mainWindow.ShowAll()
}

func InitUI(model *Model, ui *UI) {
    gtk.Init(nil)
    ui.mainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    ui.mainWindow.SetPosition(gtk.WIN_POS_CENTER)
    ui.mainWindow.SetTitle("cbxs")
    ui.mainWindow.Connect("destroy", func() {
        gtk.MainQuit()
    })
    ui.mainWindow.SetSizeRequest(1024, 768)

    ui.hud = NewHUD(ui, "")
	ui.mainWindow.Add(ui.hud)

    InitKBHandler(model, ui)

    InitCanvas(model, ui)
}

