package main

import (
	"fmt"
    "image"
	_ "image/color"
	"math"
	"path/filepath"
	 "strings"
	"time"

	_ "golang.org/x/image/colornames"

	"github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
)

const (
   LEFT_ALIGN = iota
   RIGHT_ALIGN
   CENTER
)

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

type HdrControl struct {
    container *gtk.Grid
    leftBookmark *gtk.Label
    title *gtk.Label
    rightBookmark *gtk.Label
}

func NewHdrControl() *HdrControl {
    c := &HdrControl{}

    lbkmk, err := gtk.LabelNew("")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    lbkmk.SetHAlign(gtk.ALIGN_START)
    lbkmk.SetHExpand(true)
    css, _ := lbkmk.GetStyleContext()
	css.AddClass("bkmk-btn")

	t, err := gtk.LabelNew("")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    css, _ = t.GetStyleContext()
	css.AddClass("nav-btn")

	rbkmk, err := gtk.LabelNew("")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    css, _ = rbkmk.GetStyleContext()
	css.AddClass("bkmk-btn")

    container, err := gtk.GridNew()
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    container.SetHAlign(gtk.ALIGN_CENTER)
	container.SetVAlign(gtk.ALIGN_START)
    container.SetHExpand(true)
	css, _ = container.GetStyleContext()
	css.AddClass("hdr-ctrl")
    container.Attach(lbkmk, 0, 0, 1, 1)
    container.Attach(t, 1, 0, 1, 1)
    container.Attach(rbkmk, 2, 0, 1, 1)
	container.SetSizeRequest(1000, 8)
    c.leftBookmark = lbkmk
    c.title = t
    c.rightBookmark = rbkmk
    c.container = container
    return c
}

type NavControl struct {
    container *gtk.Grid
    navBar *gtk.ProgressBar
    rightPageNum *gtk.Label
    reflowControl *gtk.Label
    readModeControl *gtk.Label
    displayModeControl *gtk.Label
    fullscreenControl *gtk.Label
    leftPageNum *gtk.Label
}

func newNavControl() *NavControl {
    nc := &NavControl{}

	nbc, err := gtk.ProgressBarNew()
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    nbc.SetHExpand(true)
	css, _ := nbc.GetStyleContext()
	css.AddClass("nav-bar")

	lpn, err := gtk.LabelNew("0")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    lpn.SetHAlign(gtk.ALIGN_START)
    lpn.SetHExpand(true)
	css, _ = lpn.GetStyleContext()
	css.AddClass("nav-btn")
	css.AddClass("page-num")

   	rc, err := gtk.LabelNew("reflow")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ = rc.GetStyleContext()
	css.AddClass("nav-btn")

	rmc, err := gtk.LabelNew("readmode")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ = rmc.GetStyleContext()
	css.AddClass("nav-btn")

	dmc, err := gtk.LabelNew("displaymode")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ = dmc.GetStyleContext()
	css.AddClass("nav-btn")

    fsc, err := gtk.LabelNew("fullscreen")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ = fsc.GetStyleContext()
	css.AddClass("nav-btn")

	rpn, err := gtk.LabelNew("1")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ = rpn.GetStyleContext()
	css.AddClass("nav-btn")
	css.AddClass("page-num")

    container, err := gtk.GridNew()
	if err != nil {
		fmt.Printf("Error creating grid %s\n", err)
	}
    container.SetHAlign(gtk.ALIGN_CENTER)
	container.SetVAlign(gtk.ALIGN_END)
    container.SetHExpand(true)
	css, _ = container.GetStyleContext()
	css.AddClass("nav-ctrl")

    container.Attach(nbc, 0, 0, 7, 1)
    container.Attach(lpn, 1, 1, 1, 1)
    container.Attach(rc, 2, 1, 1, 1)
    container.Attach(rmc, 3, 1, 1, 1)
    container.Attach(dmc, 4, 1, 1, 1)
    container.Attach(fsc, 5, 1, 1, 1)
    container.Attach(rpn, 6, 1, 1, 1)
	container.SetSizeRequest(1000, 8)
    nc.container = container
    nc.navBar = nbc
    nc.reflowControl = rc
    nc.readModeControl = rmc
    nc.leftPageNum = lpn
    nc.rightPageNum = rpn
    nc.fullscreenControl = fsc
    nc.displayModeControl = dmc

    return nc
}

func newHUD(ui *UI, title string) *gtk.Overlay {
	o, _ := gtk.OverlayNew()

    ui.hdrControl = NewHdrControl()
    ui.navControl = newNavControl()
	o.AddOverlay(ui.hdrControl.container)
	o.AddOverlay(ui.navControl.container)
    ui.hudHidden = false

    return o
}

type UI struct {
    sendMessage Messenger
    mainWindow *gtk.Window
    hud *gtk.Overlay
    hudHidden bool
    canvas *gtk.DrawingArea
    scrollbars *gtk.ScrolledWindow
    hdrControl *HdrControl
    navControl *NavControl
    longStripRender []*gdk.Pixbuf
}

func initKBHandler(model *Model, ui *UI) {
    ui.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
        keyEvent := gdk.EventKeyNewFromEvent(event)
        keyVal := keyEvent.KeyVal()
        if keyVal == gdk.KEY_q {
            m := &Message{typeName: "quit"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_d {
            m := &Message{typeName: "nextPage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_a {
            m := &Message{typeName: "previousPage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_w {
            m := &Message{typeName: "firstPage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_s {
            m := &Message{typeName: "lastPage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_Tab {
            m := &Message{typeName: "selectPage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_1 {
            initCanvas(model, ui)
            m := &Message{typeName: "setDisplayModeOnePage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_2 {
            initCanvas(model, ui)
            m := &Message{typeName: "setDisplayModeTwoPage"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_3 {
            m := &Message{typeName: "setDisplayModeLongStrip"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_grave {
            m := &Message{typeName: "toggleReadMode"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_f {
            if model.fullscreen {
                ui.mainWindow.Unfullscreen()
            } else {
                ui.mainWindow.Fullscreen()
            }
            m := &Message{typeName: "toggleFullscreen"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_o {
            dlg, _ := gtk.FileChooserNativeDialogNew("Open", ui.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
            dlg.SetCurrentFolder(model.browseDirectory)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &Message{typeName: "openFile", data: f}
                ui.sendMessage(*m)
            } else {
            }
            initCanvas(model, ui)
        } else if keyVal == gdk.KEY_c {
            m := &Message{typeName: "closeFile"}
            ui.sendMessage(*m)
            initCanvas(model, ui)
        } else if keyVal == gdk.KEY_r {
            m := &Message{typeName: "reflow"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_e {
            dlg, _ := gtk.FileChooserNativeDialogNew("Save", ui.mainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
            base := filepath.Base(model.pages[model.selectedPage].filePath)
            dlg.SetCurrentFolder(model.browseDirectory)
            dlg.SetCurrentName(base)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &Message{typeName: "exportFile", data: f}
                ui.sendMessage(*m)
            } else {
            }
        } else if keyVal == gdk.KEY_n {
            initCanvas(model, ui)
            m := &Message{typeName: "nextFile"}
            ui.sendMessage(*m)
        } else if keyVal == gdk.KEY_p {
            m := &Message{typeName: "previousFile"}
            ui.sendMessage(*m)
            initCanvas(model, ui)
        } else if keyVal == gdk.KEY_space {
            m := &Message{typeName: "toggleBookmark"}
            ui.sendMessage(*m)
        }

        //reset the hud hiding
        hudChan <-true
        if ui.hudHidden {
            ui.hud.ShowAll()
            ui.hudHidden = false
        }
     })
}

func renderNavControl(model *Model, ui *UI) {
    if len(model.leaves) < 1 {
        ui.navControl.navBar.SetFraction(0)
        ui.navControl.leftPageNum.SetText("")
        ui.navControl.reflowControl.SetText("")
        if model.readMode == RTL {
            ui.navControl.readModeControl.SetText("<")
        } else {
            ui.navControl.readModeControl.SetText(">")
        }

        if model.leafMode == ONE_PAGE {
            ui.navControl.displayModeControl.SetText("1-Page")
        } else if model.leafMode == TWO_PAGE {
            ui.navControl.displayModeControl.SetText("2-Page")
        } else {
            ui.navControl.displayModeControl.SetText("Strip")
        }

        if model.fullscreen {
            ui.navControl.fullscreenControl.SetText("fullscreen")
        } else {
            ui.navControl.fullscreenControl.SetText("")
        }

        ui.navControl.rightPageNum.SetText("")

        return 
    } else {
        leaf := model.leaves[model.currentLeaf]
        vpn := calcVersoPage(model)
        np := len(model.imgPaths)
        ui.navControl.leftPageNum.SetText("")
        ui.navControl.rightPageNum.SetText("")
        lpncss, _ := ui.navControl.leftPageNum.GetStyleContext()
        rpncss, _ := ui.navControl.rightPageNum.GetStyleContext()
        lpncss.RemoveClass("bordered")
        rpncss.RemoveClass("bordered")
        lpncss.RemoveClass("transparent")
        rpncss.RemoveClass("transparent")
        ui.navControl.leftPageNum.Show()
        ui.navControl.rightPageNum.Show()

        if model.readMode == RTL {
            if np > 0 {
                ui.navControl.navBar.SetInverted(true)
                ui.navControl.navBar.SetFraction((float64(vpn)+float64(len(leaf.pages)))/float64(np))
            }

            if len(leaf.pages) > 1 {
                ui.navControl.rightPageNum.SetText(fmt.Sprintf("%d", vpn))
                ui.navControl.leftPageNum.SetText(fmt.Sprintf("%d", vpn+1))
                if model.selectedPage == vpn {
                    rpncss.AddClass("bordered")
                } else if model.selectedPage == vpn+1 {
                    lpncss.AddClass("bordered")
                }
            } else {
                rpncss.AddClass("transparent")
                ui.navControl.leftPageNum.SetText(fmt.Sprintf("%d", vpn))
                lpncss.AddClass("bordered")
            }
            ui.navControl.readModeControl.SetText("<")
        } else {
            if np > 0 {
                ui.navControl.navBar.SetInverted(false)
                ui.navControl.navBar.SetFraction((float64(vpn)+float64(len(leaf.pages)))/float64(np))
            }

            if len(leaf.pages) > 1 {
                ui.navControl.leftPageNum.SetText(fmt.Sprintf("%d", vpn))
                ui.navControl.rightPageNum.SetText(fmt.Sprintf("%d", vpn+1))
                if model.selectedPage == vpn {
                    lpncss.AddClass("bordered")
                } else if model.selectedPage == vpn+1 {
                    rpncss.AddClass("bordered")
                }
            } else {
                lpncss.AddClass("transparent")
                ui.navControl.rightPageNum.SetText(fmt.Sprintf("%d", vpn))
                rpncss.AddClass("bordered")
            }
            ui.navControl.readModeControl.SetText(">")
        }

        if model.leafMode == ONE_PAGE {
            ui.navControl.displayModeControl.SetText("1-Page")
        } else if model.leafMode == TWO_PAGE {
            ui.navControl.displayModeControl.SetText("2-Page")
        } else {
            ui.navControl.displayModeControl.SetText("Strip")
        }

        if model.fullscreen {
            ui.navControl.fullscreenControl.SetText("fullscreen")
        } else {
            ui.navControl.fullscreenControl.SetText("")
        }

        if leaf.pages[0].Orientation == LANDSCAPE {
            ui.navControl.reflowControl.SetText("-")
        } else {
            ui.navControl.reflowControl.SetText("|")
            if len(leaf.pages) > 1 {
                ui.navControl.reflowControl.SetText("||")
            }
        }
    }
}

func renderHdrControl(model *Model, ui *UI) {
    vpn := calcVersoPage(model)
    css, _ := ui.hdrControl.leftBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    css, _ = ui.hdrControl.rightBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    ui.hdrControl.title.SetText("")
    if len(model.leaves) < 1 || model.bookmarks == nil {
        ui.hdrControl.title.SetText("Loading...")
        return 
    } else {
        lbkmkcss, _ := ui.hdrControl.leftBookmark.GetStyleContext()
        rbkmkcss, _ := ui.hdrControl.rightBookmark.GetStyleContext()
        leaf := model.leaves[model.currentLeaf]
        title := strings.TrimSuffix(filepath.Base(model.filePath), filepath.Ext(model.filePath))

        if model.readMode == RTL {
            b := model.bookmarks.Find(vpn)
            if b != nil {
                if len(leaf.pages) > 1 {
                    rbkmkcss.AddClass("marked")
                } else {
                    rbkmkcss.AddClass("transparent")
                    lbkmkcss.AddClass("marked")
                }
            } 

            if len(leaf.pages) > 1 {
                b = model.bookmarks.Find(vpn+1)
                if b != nil {
                    lbkmkcss.AddClass("marked")
                }
            }
        } else {
            b := model.bookmarks.Find(vpn)
            if b != nil {
                if len(leaf.pages) > 1 {
                    lbkmkcss.AddClass("marked")
                } else {
                    lbkmkcss.AddClass("transparent")
                    rbkmkcss.AddClass("marked")
                }
            } 

            if len(leaf.pages) > 1 {
                b = model.bookmarks.Find(vpn+1)
                if b != nil {
                    rbkmkcss.AddClass("marked")
                }
            }
        }
        ui.hdrControl.title.SetText(title)
    }
}

func renderHud(model *Model, ui *UI) {
	renderHdrControl(model, ui)
	renderNavControl(model, ui)
}

func Render(model *Model, ui *UI) {
    glib.IdleAdd(func(){
        renderHud(model, ui)
        ui.mainWindow.QueueDraw()
    })
}

type OnePageLayout struct {
    canvas *gtk.DrawingArea
    cr *cairo.Context
    page *Page
}

func newOnePageLayout(canvas *gtk.DrawingArea, cr *cairo.Context,
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
func newTwoPageLayout(model *Model, canvas *gtk.DrawingArea, cr *cairo.Context, leaf *Leaf) *TwoPageLayout {

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

func newLongStripLayout(canvas *gtk.DrawingArea, cr *cairo.Context, leaf *Leaf) *LongStripLayout {

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

func renderLongStripLayout(model *Model, ui *UI, layout *LongStripLayout) error {
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

func initRenderer(model *Model, ui *UI) {
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
            lo := newTwoPageLayout(model, canvas, cr, leaf)
            renderTwoPageLayout(lo)
        } else if model.leafMode == ONE_PAGE {
            lo := newOnePageLayout(canvas, cr, leaf.pages[0])
            renderOnePageLayout(lo)
        } else {
            lo := newLongStripLayout(canvas, cr, leaf)
            renderLongStripLayout(model, ui, lo)
        }
        w := ui.mainWindow.GetAllocatedWidth() - 40
        ui.hdrControl.container.SetSizeRequest(w, 8)
        ui.navControl.container.SetSizeRequest(w, 8)
    })
}

func initCanvas(model *Model, ui *UI) {
    if ui.canvas != nil {
        ui.hud.Remove(ui.canvas)
        ui.canvas.Destroy()
        ui.canvas = nil
    }

    ui.canvas, _ = gtk.DrawingAreaNew()
	ui.hud.Add(ui.canvas)
    initRenderer(model, ui)
    ui.mainWindow.ShowAll()
}

func initCss() {
	css, err := gtk.CssProviderNew()
    if err != nil {
		fmt.Printf("css error %s\n", err)
	}

    data, err := LoadTextFile("assets/index.css")
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

func InitUI(model *Model, ui *UI, messenger Messenger) {
    gtk.Init(nil)
    ui.sendMessage = messenger
    ui.mainWindow, _ = gtk.WindowNew(gtk.WINDOW_TOPLEVEL)
    ui.mainWindow.SetPosition(gtk.WIN_POS_CENTER)
    ui.mainWindow.SetTitle("cbxv")
    ui.mainWindow.Connect("destroy", func() {
        gtk.MainQuit()
    })
    ui.mainWindow.SetSizeRequest(1024, 768)

    ui.mainWindow.Connect("configure-event", func() {
        w := ui.mainWindow.GetAllocatedWidth() - 40
        ui.hdrControl.container.SetSizeRequest(w, 8)
        ui.navControl.container.SetSizeRequest(w, 8)
    })

    initCss()

    ui.hud = newHUD(ui, "")
    ui.scrollbars, _ = gtk.ScrolledWindowNew(nil, nil)
    ui.scrollbars.Add(ui.hud)
	ui.mainWindow.Add(ui.scrollbars)

    initKBHandler(model, ui)

    initCanvas(model, ui)

    hudChan = make(chan bool)
    hudTicker = time.NewTicker(time.Second * 10)
    defer hudTicker.Stop()

    go hudHandler(ui)
}

func StopUI(model *Model, ui *UI) {
    hudChan <-false
}

