package ui

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

    "example.com/cbxv-gotk3/internal/util"
    "example.com/cbxv-gotk3/internal/model"
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

func initKBHandler(model *model.Model, ui *UI) {
    ui.mainWindow.Connect("key-press-event", func(widget *gtk.Window, event *gdk.Event) {
        keyEvent := gdk.EventKeyNewFromEvent(event)
        keyVal := keyEvent.KeyVal()
        if keyVal == gdk.KEY_q {
            ui.sendMessage(util.Message{TypeName: "quit"})
        } else if keyVal == gdk.KEY_d {
            ui.sendMessage(util.Message{TypeName: "nextPage"})
        } else if keyVal == gdk.KEY_a {
            ui.sendMessage(util.Message{TypeName: "previousPage"})
        } else if keyVal == gdk.KEY_w {
            ui.sendMessage(util.Message{TypeName: "firstPage"})
        } else if keyVal == gdk.KEY_s {
            ui.sendMessage(util.Message{TypeName: "lastPage"})
        } else if keyVal == gdk.KEY_Tab {
            ui.sendMessage(util.Message{TypeName: "selectPage"})
        } else if keyVal == gdk.KEY_1 {
            initCanvas(model, ui)
            ui.sendMessage(util.Message{TypeName: "setDisplayModeOnePage"})
        } else if keyVal == gdk.KEY_2 {
            initCanvas(model, ui)
            ui.sendMessage(util.Message{TypeName: "setDisplayModeTwoPage"})
        } else if keyVal == gdk.KEY_3 {
            ui.sendMessage(util.Message{TypeName: "setDisplayModeLongStrip"})
        } else if keyVal == gdk.KEY_grave {
            ui.sendMessage(util.Message{TypeName: "toggleReadMode"})
        } else if keyVal == gdk.KEY_f {
            if model.Fullscreen {
                ui.mainWindow.Unfullscreen()
            } else {
                ui.mainWindow.Fullscreen()
            }
            ui.sendMessage(util.Message{TypeName: "toggleFullscreen"})
        } else if keyVal == gdk.KEY_o {
            dlg, _ := gtk.FileChooserNativeDialogNew("Open", ui.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
            dlg.SetCurrentFolder(model.BrowseDirectory)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "openFile", Data: f}
                ui.sendMessage(*m)
            } else {
            }
            initCanvas(model, ui)
        } else if keyVal == gdk.KEY_c {
            ui.sendMessage(util.Message{TypeName: "closeFile"})
            initCanvas(model, ui)
        } else if keyVal == gdk.KEY_r {
            ui.sendMessage(util.Message{TypeName: "reflow"})
        } else if keyVal == gdk.KEY_e {
            dlg, _ := gtk.FileChooserNativeDialogNew("Save", ui.mainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
            base := filepath.Base(model.Pages[model.SelectedPage].FilePath)
            dlg.SetCurrentFolder(model.BrowseDirectory)
            dlg.SetCurrentName(base)
            output := dlg.NativeDialog.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "exportFile", Data: f}
                ui.sendMessage(*m)
            } else {
            }
        } else if keyVal == gdk.KEY_n {
            initCanvas(model, ui)
            ui.sendMessage(util.Message{TypeName: "nextFile"})
        } else if keyVal == gdk.KEY_p {
            ui.sendMessage(util.Message{TypeName: "previousFile"})
            initCanvas(model, ui)
        } else if keyVal == gdk.KEY_space {
            ui.sendMessage(util.Message{TypeName: "toggleBookmark"})
        }

        //reset the hud hiding
        hudChan <-true
        if ui.hudHidden {
            ui.hud.ShowAll()
            ui.hudHidden = false
        }
     })
}

func renderNavControl(m *model.Model, ui *UI) {
    if len(m.Leaves) < 1 {
        ui.navControl.navBar.SetFraction(0)
        ui.navControl.leftPageNum.SetText("")
        ui.navControl.reflowControl.SetText("")
        if m.ReadMode == model.RTL {
            ui.navControl.readModeControl.SetText("<")
        } else {
            ui.navControl.readModeControl.SetText(">")
        }

        if m.LeafMode == model.ONE_PAGE {
            ui.navControl.displayModeControl.SetText("1-Page")
        } else if m.LeafMode == model.TWO_PAGE {
            ui.navControl.displayModeControl.SetText("2-Page")
        } else {
            ui.navControl.displayModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            ui.navControl.fullscreenControl.SetText("fullscreen")
        } else {
            ui.navControl.fullscreenControl.SetText("")
        }

        ui.navControl.rightPageNum.SetText("")

        return 
    } else {
        leaf := m.Leaves[m.CurrentLeaf]
        vpn := m.CalcVersoPage()
        np := len(m.ImgPaths)
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

        if m.ReadMode == model.RTL {
            if np > 0 {
                ui.navControl.navBar.SetInverted(true)
                ui.navControl.navBar.SetFraction((float64(vpn)+float64(len(leaf.Pages)))/float64(np))
            }

            if len(leaf.Pages) > 1 {
                ui.navControl.rightPageNum.SetText(fmt.Sprintf("%d", vpn))
                ui.navControl.leftPageNum.SetText(fmt.Sprintf("%d", vpn+1))
                if m.SelectedPage == vpn {
                    rpncss.AddClass("bordered")
                } else if m.SelectedPage == vpn+1 {
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
                ui.navControl.navBar.SetFraction((float64(vpn)+float64(len(leaf.Pages)))/float64(np))
            }

            if len(leaf.Pages) > 1 {
                ui.navControl.leftPageNum.SetText(fmt.Sprintf("%d", vpn))
                ui.navControl.rightPageNum.SetText(fmt.Sprintf("%d", vpn+1))
                if m.SelectedPage == vpn {
                    lpncss.AddClass("bordered")
                } else if m.SelectedPage == vpn+1 {
                    rpncss.AddClass("bordered")
                }
            } else {
                lpncss.AddClass("transparent")
                ui.navControl.rightPageNum.SetText(fmt.Sprintf("%d", vpn))
                rpncss.AddClass("bordered")
            }
            ui.navControl.readModeControl.SetText(">")
        }

        if m.LeafMode == model.ONE_PAGE {
            ui.navControl.displayModeControl.SetText("1-Page")
        } else if m.LeafMode == model.TWO_PAGE {
            ui.navControl.displayModeControl.SetText("2-Page")
        } else {
            ui.navControl.displayModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            ui.navControl.fullscreenControl.SetText("fullscreen")
        } else {
            ui.navControl.fullscreenControl.SetText("")
        }

        if leaf.Pages[0].Orientation == model.LANDSCAPE {
            ui.navControl.reflowControl.SetText("-")
        } else {
            ui.navControl.reflowControl.SetText("|")
            if len(leaf.Pages) > 1 {
                ui.navControl.reflowControl.SetText("||")
            }
        }
    }
}

func renderHdrControl(m *model.Model, ui *UI) {
    vpn := m.CalcVersoPage()
    css, _ := ui.hdrControl.leftBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    css, _ = ui.hdrControl.rightBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    ui.hdrControl.title.SetText("")
    if len(m.Leaves) < 1 || m.Bookmarks == nil {
        ui.hdrControl.title.SetText("Loading...")
        return 
    } else {
        lbkmkcss, _ := ui.hdrControl.leftBookmark.GetStyleContext()
        rbkmkcss, _ := ui.hdrControl.rightBookmark.GetStyleContext()
        leaf := m.Leaves[m.CurrentLeaf]
        title := strings.TrimSuffix(filepath.Base(m.FilePath), filepath.Ext(m.FilePath))

        if m.ReadMode == model.RTL {
            b := m.Bookmarks.Find(vpn)
            if b != nil {
                if len(leaf.Pages) > 1 {
                    rbkmkcss.AddClass("marked")
                } else {
                    rbkmkcss.AddClass("transparent")
                    lbkmkcss.AddClass("marked")
                }
            } 

            if len(leaf.Pages) > 1 {
                b = m.Bookmarks.Find(vpn+1)
                if b != nil {
                    lbkmkcss.AddClass("marked")
                }
            }
        } else {
            b := m.Bookmarks.Find(vpn)
            if b != nil {
                if len(leaf.Pages) > 1 {
                    lbkmkcss.AddClass("marked")
                } else {
                    lbkmkcss.AddClass("transparent")
                    rbkmkcss.AddClass("marked")
                }
            } 

            if len(leaf.Pages) > 1 {
                b = m.Bookmarks.Find(vpn+1)
                if b != nil {
                    rbkmkcss.AddClass("marked")
                }
            }
        }
        ui.hdrControl.title.SetText(title)
    }
}

func renderHud(m *model.Model, ui *UI) {
	renderHdrControl(m, ui)
	renderNavControl(m, ui)
}

func Render(m *model.Model, ui *UI) {
    glib.IdleAdd(func(){
        renderHud(m, ui)
        ui.mainWindow.QueueDraw()
    })
}

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

func initRenderer(m *model.Model, ui *UI) {
    ui.longStripRender = nil
    ui.canvas.Connect("draw", func(canvas *gtk.DrawingArea, cr *cairo.Context) {
        cr.SetSourceRGB(0,0,0)
        cr.Rectangle(0,0,float64(ui.canvas.GetAllocatedWidth()), float64(ui.canvas.GetAllocatedHeight()))
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
            renderLongStripLayout(m, ui, lo)
        }
        w := ui.mainWindow.GetAllocatedWidth() - 40
        ui.hdrControl.container.SetSizeRequest(w, 8)
        ui.navControl.container.SetSizeRequest(w, 8)
    })
}

func initCanvas(m *model.Model, ui *UI) {
    if ui.canvas != nil {
        ui.hud.Remove(ui.canvas)
        ui.canvas.Destroy()
        ui.canvas = nil
    }

    ui.canvas, _ = gtk.DrawingAreaNew()
	ui.hud.Add(ui.canvas)
    initRenderer(m, ui)
    ui.mainWindow.ShowAll()
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

func InitUI(m *model.Model, ui *UI, messenger util.Messenger) {
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

    initKBHandler(m, ui)

    initCanvas(m, ui)

    hudChan = make(chan bool)
    hudTicker = time.NewTicker(time.Second * 10)
    defer hudTicker.Stop()

    go hudHandler(ui)

    ui.mainWindow.ShowAll()
}

func StopUI(model *model.Model, ui *UI) {
    hudChan <-false
}

