package ui

import (
	"fmt"

	"github.com/gotk3/gotk3/gtk"
    "example.com/cbxv-gotk3/internal/model"
)

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

func NewNavControl() *NavControl {
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

func (c *NavControl) Render(m *model.Model) {
    if len(m.Leaves) < 1 {
        c.navBar.SetFraction(0)
        c.leftPageNum.SetText("")
        c.reflowControl.SetText("")
        if m.ReadMode == model.RTL {
            c.readModeControl.SetText("<")
        } else {
            c.readModeControl.SetText(">")
        }

        if m.LeafMode == model.ONE_PAGE {
            c.displayModeControl.SetText("1-Page")
        } else if m.LeafMode == model.TWO_PAGE {
            c.displayModeControl.SetText("2-Page")
        } else {
            c.displayModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.fullscreenControl.SetText("fullscreen")
        } else {
            c.fullscreenControl.SetText("")
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
            c.readModeControl.SetText("<")
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
            c.readModeControl.SetText(">")
        }

        if m.LeafMode == model.ONE_PAGE {
            c.displayModeControl.SetText("1-Page")
        } else if m.LeafMode == model.TWO_PAGE {
            c.displayModeControl.SetText("2-Page")
        } else {
            c.displayModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.fullscreenControl.SetText("fullscreen")
        } else {
            c.fullscreenControl.SetText("")
        }

        if leaf.Pages[0].Orientation == model.LANDSCAPE {
            c.reflowControl.SetText("-")
        } else {
            c.reflowControl.SetText("|")
            if len(leaf.Pages) > 1 {
                c.reflowControl.SetText("||")
            }
        }
    }
}



