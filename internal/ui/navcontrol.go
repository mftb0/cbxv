package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
    "example.com/cbxv-gotk3/internal/model"
    "example.com/cbxv-gotk3/internal/util"
)

const (
    DIR_LTR_ICN = "⯈"   // u+2bc8
    DIR_RTL_ICN = "⯇"   // u+2bc7
    FS_MAX_ICN  = "⛶ "  // u+26f6 - square four corners
    FS_RES_ICN  = "🮻 "  // u+1fbbb 
    SD_ONE_ICN  = "Ⅰ"   // u+2160 - roman numeral 1
    SD_TWO_ICN  = "Ⅱ"   // u+2161 - roman numeral 2
    SD_DBL_ICN  = "█"   // u+2588
    APP_CLS_ICN = "⮽ "  // u+2bbd
    CBX_CLS_ICN = "⮾ "  // u+2bbe
)

type NavControl struct {
    container *gtk.Grid
    navBar *gtk.ProgressBar
    rightPageNum *gtk.Label
    progName *gtk.Label
    progVersion *gtk.Label
    spreadControl *gtk.Button
    readModeControl *gtk.Button
    displayModeControl *gtk.Label
    fullscreenControl *gtk.Button
    leftPageNum *gtk.Label
}

func NewNavControl(m *model.Model, u *UI) *NavControl {
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
	css, _ = lpn.GetStyleContext()
	css.AddClass("nav-btn")
	css.AddClass("page-num")

	pn, err := gtk.LabelNew("cbxv")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    pn.SetHAlign(gtk.ALIGN_START)
	css, _ = pn.GetStyleContext()
	css.AddClass("nav-btn")

	pv, err := gtk.LabelNew("v0.0.1")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    pv.SetHAlign(gtk.ALIGN_START)
    pv.SetHExpand(true)
	css, _ = pv.GetStyleContext()
	css.AddClass("nav-btn")

	dc, err := gtk.ButtonNewWithLabel(DIR_LTR_ICN)
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    dc.SetTooltipText("Direction Toggle")
	css, _ = dc.GetStyleContext()
	css.AddClass("nav-btn")

	dmc, err := gtk.LabelNew("Layout")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    dmc.SetTooltipText("Layout")
	css, _ = dmc.GetStyleContext()
	css.AddClass("nav-btn")

   	sc, err := gtk.ButtonNewWithLabel("Spread Toggle")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    sc.SetTooltipText("Spread Toggle")
	css, _ = sc.GetStyleContext()
	css.AddClass("nav-btn")

    fsc, err := gtk.ButtonNewWithLabel(FS_MAX_ICN) 
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    sc.SetTooltipText("Fullscreen Toggle")
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

    sc.Connect("clicked", func() { 
        u.sendMessage(util.Message{TypeName: "spread"})
    })

    dc.Connect("clicked", func() { 
        u.sendMessage(util.Message{TypeName: "toggleReadMode"})
    })

    fsc.Connect("clicked", func() { 
        if m.Fullscreen {
            u.mainWindow.Unfullscreen()
        } else {
            u.mainWindow.Fullscreen()
        }
        u.sendMessage(util.Message{TypeName: "toggleFullscreen"})
    })

    container.Attach(nbc, 0, 0, 9, 1)
    container.Attach(lpn, 1, 1, 1, 1)
    container.Attach(pn, 2, 1, 1, 1)
    container.Attach(pv, 3, 1, 1, 1)
    container.Attach(dc, 4, 1, 1, 1)
    container.Attach(dmc, 5, 1, 1, 1)
    container.Attach(sc, 6, 1, 1, 1)
    container.Attach(fsc, 7, 1, 1, 1)
    container.Attach(rpn, 8, 1, 1, 1)
	container.SetSizeRequest(1000, 8)
    nc.container = container
    nc.navBar = nbc
    nc.leftPageNum = lpn
    nc.progName = pn
    nc.progVersion = pv
    nc.readModeControl = dc
    nc.displayModeControl = dmc
    nc.spreadControl = sc
    nc.fullscreenControl = fsc
    nc.rightPageNum = rpn
    
    return nc
}

func (c *NavControl) Render(m *model.Model) {
    if len(m.Leaves) < 1 {
        c.navBar.SetFraction(0)
        c.leftPageNum.SetText("")
        c.spreadControl.SetLabel("")
        if m.ReadMode == model.RTL {
            c.readModeControl.SetLabel(DIR_RTL_ICN)
        } else {
            c.readModeControl.SetLabel(DIR_LTR_ICN)
        }

        if m.LeafMode == model.ONE_PAGE {
            c.displayModeControl.SetText("1-Page")
        } else if m.LeafMode == model.TWO_PAGE {
            c.displayModeControl.SetText("2-Page")
        } else {
            c.displayModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.fullscreenControl.SetLabel(FS_MAX_ICN)
        } else {
            c.fullscreenControl.SetLabel(FS_RES_ICN)
        }

        c.rightPageNum.SetText("")

        return 
    } else {
        leaf := m.Leaves[m.CurrentLeaf]
        vpn := m.CalcVersoPage()
        np := len(m.ImgPaths)
        c.leftPageNum.SetText("")
        c.rightPageNum.SetText("")
        lpncss, _ := c.leftPageNum.GetStyleContext()
        rpncss, _ := c.rightPageNum.GetStyleContext()
        lpncss.RemoveClass("bordered")
        rpncss.RemoveClass("bordered")
        lpncss.RemoveClass("transparent")
        rpncss.RemoveClass("transparent")
        c.leftPageNum.Show()
        c.rightPageNum.Show()

        if m.ReadMode == model.RTL {
            if np > 0 {
                c.navBar.SetInverted(true)
                c.navBar.SetFraction((float64(vpn)+float64(len(leaf.Pages)))/float64(np))
            }

            if len(leaf.Pages) > 1 {
                c.rightPageNum.SetText(fmt.Sprintf("%d", vpn))
                c.leftPageNum.SetText(fmt.Sprintf("%d", vpn+1))
                if m.SelectedPage == vpn {
                    rpncss.AddClass("bordered")
                } else if m.SelectedPage == vpn+1 {
                    lpncss.AddClass("bordered")
                }
            } else {
                rpncss.AddClass("transparent")
                c.leftPageNum.SetText(fmt.Sprintf("%d", vpn))
                lpncss.AddClass("bordered")
            }
            c.readModeControl.SetLabel(DIR_RTL_ICN)
        } else {
            if np > 0 {
                c.navBar.SetInverted(false)
                c.navBar.SetFraction((float64(vpn)+float64(len(leaf.Pages)))/float64(np))
            }

            if len(leaf.Pages) > 1 {
                c.leftPageNum.SetText(fmt.Sprintf("%d", vpn))
                c.rightPageNum.SetText(fmt.Sprintf("%d", vpn+1))
                if m.SelectedPage == vpn {
                    lpncss.AddClass("bordered")
                } else if m.SelectedPage == vpn+1 {
                    rpncss.AddClass("bordered")
                }
            } else {
                lpncss.AddClass("transparent")
                c.rightPageNum.SetText(fmt.Sprintf("%d", vpn))
                rpncss.AddClass("bordered")
            }
            c.readModeControl.SetLabel(DIR_LTR_ICN)
        }

        if m.LeafMode == model.ONE_PAGE {
            c.displayModeControl.SetText("1-Page")
        } else if m.LeafMode == model.TWO_PAGE {
            c.displayModeControl.SetText("2-Page")
        } else {
            c.displayModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.fullscreenControl.SetLabel(FS_RES_ICN)
        } else {
            c.fullscreenControl.SetLabel(FS_MAX_ICN)
        }

        if leaf.Pages[0].Orientation == model.LANDSCAPE {
            c.spreadControl.SetLabel(SD_DBL_ICN) 
        } else {
            c.spreadControl.SetLabel(SD_ONE_ICN)
            if len(leaf.Pages) > 1 {
                c.spreadControl.SetLabel(SD_TWO_ICN)
            }
        }
    }
}



