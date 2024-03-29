package ui

import (
    "fmt"

    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/gtk"
    "github.com/mftb0/cbxv/internal/model"
    "github.com/mftb0/cbxv/internal/util"
)

const (
    DIR_LTR_ICN = "▶" // u+25b6
    DIR_RTL_ICN = "◀" // u+25c0
    SD_ONE_ICN = "|"
    SD_TWO_ICN = "||"
    SD_DBL_ICN = "█" // u+2588
)

type PageViewNavControl struct {
    ui                *UI
    container         *gtk.Grid
    navBar            *gtk.ProgressBar
    rightPageNum      *gtk.Button
    progName          *gtk.Label
    progVersion       *gtk.Label
    directionControl  *gtk.Button
    layoutModeControl *gtk.Label
    spreadControl     *gtk.Button
    hiddenPageControl *gtk.ComboBoxText
    hpcSignalHandle   *glib.SignalHandle
    fullscreenControl *gtk.Button
    leftPageNum       *gtk.Button
}

func NewNavControl(m *model.Model, u *UI) *PageViewNavControl {
    nc := &PageViewNavControl{}
    nc.ui = u

    nbc, err := gtk.ProgressBarNew()
    if err != nil {
        fmt.Printf("Error creating label %s\n", err)
    }
    nbc.SetHExpand(true)
    css, _ := nbc.GetStyleContext()
    css.AddClass("nav-bar")

    lpn := util.CreateButton("0", "nav-btn", util.S("Left Page Index"))
    lpn.SetHAlign(gtk.ALIGN_START)
    css.AddClass("page-num")

    pn := util.CreateLabel(m.ProgramName, "nav-btn", nil)
    pn.SetHAlign(gtk.ALIGN_START)

    pv := util.CreateLabel(m.ProgramVersion, "nav-btn", nil)
    pv.SetHAlign(gtk.ALIGN_START)
    pv.SetHExpand(true)

    dc := util.CreateButton(DIR_LTR_ICN, "nav-btn", util.S("Direction Toggle"))
    lmc := util.CreateLabel("Layout", "nav-btn", util.S("Layout"))
    jc := util.CreateButton("Join", "nav-btn", util.S("Join Toggle"))

    hpc, err := gtk.ComboBoxTextNew()
    if err != nil {
        fmt.Printf("Error creating control %s\n", err)
    }
    css, _ = hpc.GetStyleContext()
    css.AddClass("nav-btn")

    fsc := util.CreateButton(util.FullscreenIcon(), "nav-btn", util.S("Fullscreen Toggle"))

    rpn := util.CreateButton("1", "nav-btn", util.S("Right Page Index"))
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

    lpn.Connect("clicked", func() bool {
        s := m.Spreads[m.SpreadIndex]
        if m.Direction == model.RTL {
            if len(s.Pages) > 1 {
                if s.VersoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
        } else {
            if len(s.Pages) > 1 {
                if s.RectoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
        }
        return true
    })

    jc.Connect("clicked", func() {
        u.Commands.Names["toggleJoin"].Execute()
    })

    dc.Connect("clicked", func() {
        u.Commands.Names["toggleDirection"].Execute()
    })

    fsc.Connect("clicked", func() {
        u.Commands.Names["toggleFullscreen"].Execute()
    })

    rpn.Connect("clicked", func() bool {
        s := m.Spreads[m.SpreadIndex]
        if m.Direction == model.RTL {
            if len(s.Pages) > 1 {
                if s.RectoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
        } else {
            if len(s.Pages) > 1 {
                if s.VersoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
        }
        return true
    })

    container.Attach(nbc, 0, 0, 10, 1)
    container.Attach(lpn, 1, 1, 1, 1)
    container.Attach(pn, 2, 1, 1, 1)
    container.Attach(pv, 3, 1, 1, 1)
    container.Attach(dc, 4, 1, 1, 1)
    container.Attach(lmc, 5, 1, 1, 1)
    container.Attach(jc, 6, 1, 1, 1)
    container.Attach(hpc, 7, 1, 1, 1)
    container.Attach(fsc, 8, 1, 1, 1)
    container.Attach(rpn, 9, 1, 1, 1)
    container.SetSizeRequest(1000, 8)
    nc.container = container
    nc.navBar = nbc
    nc.leftPageNum = lpn
    nc.progName = pn
    nc.progVersion = pv
    nc.directionControl = dc
    nc.layoutModeControl = lmc
    nc.spreadControl = jc
    nc.hiddenPageControl = hpc
    nc.fullscreenControl = fsc
    nc.rightPageNum = rpn

    return nc
}

func (c *PageViewNavControl) Render(m *model.Model) {
    if len(m.Spreads) < 1 {
        c.navBar.SetFraction(0)
        c.leftPageNum.SetLabel("")
        c.spreadControl.SetLabel("")
        if m.Direction == model.RTL {
            c.directionControl.SetLabel(DIR_RTL_ICN)
        } else {
            c.directionControl.SetLabel(DIR_LTR_ICN)
        }

        if m.LayoutMode == model.ONE_PAGE {
            c.layoutModeControl.SetText("1-Page")
        } else if m.LayoutMode == model.TWO_PAGE {
            c.layoutModeControl.SetText("2-Page")
        } else {
            c.layoutModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.fullscreenControl.SetLabel(util.RestoreIcon())
        } else {
            c.fullscreenControl.SetLabel(util.FullscreenIcon())
        }

        c.rightPageNum.SetLabel("")

        return
    } else {
        spread := m.Spreads[m.SpreadIndex]
        np := len(m.Spreads)
        c.leftPageNum.SetLabel("")
        c.rightPageNum.SetLabel("")
        lpncss, _ := c.leftPageNum.GetStyleContext()
        rpncss, _ := c.rightPageNum.GetStyleContext()
        lpncss.RemoveClass("bordered")
        rpncss.RemoveClass("bordered")
        lpncss.RemoveClass("transparent")
        rpncss.RemoveClass("transparent")
        c.leftPageNum.Show()
        c.rightPageNum.Show()

        if m.Direction == model.RTL {
            if np > 0 {
                c.navBar.SetInverted(true)
                c.navBar.SetFraction((float64(m.SpreadIndex) + 1) / float64(np))
            }

            if len(spread.Pages) > 1 {
                c.rightPageNum.SetLabel(fmt.Sprintf("%d", spread.VersoPage()))
                c.leftPageNum.SetLabel(fmt.Sprintf("%d", spread.RectoPage()))
                if m.PageIndex == spread.VersoPage() {
                    rpncss.AddClass("bordered")
                } else if m.PageIndex == spread.RectoPage() {
                    lpncss.AddClass("bordered")
                }
            } else {
                rpncss.AddClass("transparent")
                c.leftPageNum.SetLabel(fmt.Sprintf("%d", spread.VersoPage()))
                lpncss.AddClass("bordered")
            }
            c.directionControl.SetLabel(DIR_RTL_ICN)
        } else {
            if np > 0 {
                c.navBar.SetInverted(false)
                c.navBar.SetFraction((float64(m.SpreadIndex) + 1) / float64(np))
            }

            if len(spread.Pages) > 1 {
                c.leftPageNum.SetLabel(fmt.Sprintf("%d", spread.VersoPage()))
                c.rightPageNum.SetLabel(fmt.Sprintf("%d", spread.RectoPage()))
                if m.PageIndex == spread.VersoPage() {
                    lpncss.AddClass("bordered")
                } else if m.PageIndex == spread.RectoPage() {
                    rpncss.AddClass("bordered")
                }
            } else {
                lpncss.AddClass("transparent")
                c.rightPageNum.SetLabel(fmt.Sprintf("%d", spread.VersoPage()))
                rpncss.AddClass("bordered")
            }
            c.directionControl.SetLabel(DIR_LTR_ICN)
        }

        if m.LayoutMode == model.ONE_PAGE {
            c.layoutModeControl.SetText("1-Page")
        } else if m.LayoutMode == model.TWO_PAGE {
            c.layoutModeControl.SetText("2-Page")
        } else {
            c.layoutModeControl.SetText("Strip")
        }

        // If there was a hpc signal handler clean it out
        // and disconnect
        if c.hpcSignalHandle != nil {
            c.hiddenPageControl.HandlerDisconnect(*c.hpcSignalHandle)
            c.hiddenPageControl.RemoveAll()
        }

        // If an hpc control is needed populate and
        // hook it up
        if m.HiddenPages == true {
            for i := len(m.Pages) - 1; i > -1; i-- {
                p := m.Pages[i]
                if p.Hidden {
                    v := fmt.Sprintf("%d", i)
                    id := fmt.Sprintf("Page %d", i)
                    c.hiddenPageControl.Append(id, v)
                    c.hiddenPageControl.SetActiveID(id)
                }
            }
            c.hiddenPageControl.Append("hidden", "Hidden")
            c.hiddenPageControl.SetActiveID("hidden")

            hndl := c.hiddenPageControl.Connect("changed", func() {
                v := c.hiddenPageControl.GetActiveText()
                if v == "Hidden" {
                    return
                }
                c.ui.Commands.Names["showPage"].Execute(v)
            })
            c.hpcSignalHandle = &hndl
        }

        if m.Fullscreen {
            c.fullscreenControl.SetLabel(util.RestoreIcon())
        } else {
            c.fullscreenControl.SetLabel(util.FullscreenIcon())
        }

        if spread.Pages[0].Span == model.DOUBLE {
            c.spreadControl.SetLabel(SD_DBL_ICN)
        } else {
            c.spreadControl.SetLabel(SD_ONE_ICN)
            if len(spread.Pages) > 1 {
                c.spreadControl.SetLabel(SD_TWO_ICN)
            }
        }
    }
}
