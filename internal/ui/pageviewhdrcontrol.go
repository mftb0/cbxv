package ui

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/gotk3/gotk3/gtk"

    "github.com/mftb0/cbxv/internal/model"
    "github.com/mftb0/cbxv/internal/util"
)

const (
    APP_HLP_ICN = "?"
)

type PageViewHdrControl struct {
    container     *gtk.Grid
    leftBookmark  *gtk.Button
    spinner       *gtk.Spinner
    fileControl   *gtk.Button
    helpControl   *gtk.Button
    rightBookmark *gtk.Button
}

func NewHdrControl(m *model.Model, u *UI) *PageViewHdrControl {
    c := &PageViewHdrControl{}

    lbkmk := util.CreateButton("", "bkmk-btn", nil)
    lbkmk.SetHAlign(gtk.ALIGN_START)
    lbkmk.SetHExpand(true)

    spn, err := gtk.SpinnerNew()
    if err != nil {
        fmt.Printf("Error creating label %s\n", err)
    }
    css, _ := spn.GetStyleContext()
    css.AddClass("nav-btn")

    fc := util.CreateButton("File", "nav-btn", util.S("Open File"))
    hc := util.CreateButton(APP_HLP_ICN, "nav-btn", util.S("Help"))
    rbkmk := util.CreateButton("", "bkmk-btn", nil)

    lbkmk.Connect("clicked", func() bool {
        s := m.Spreads[m.SpreadIndex]
        if m.Direction == model.RTL {
            if len(s.Pages) > 1 {
                if s.VersoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
            u.Commands.Names["toggleBookmark"].Execute()
        } else {
            if len(s.Pages) > 1 {
                if s.RectoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
            u.Commands.Names["toggleBookmark"].Execute()
        }
        return true
    })

    fc.Connect("clicked", func() bool {
        u.Commands.Names["openFile"].Execute()
        return true
    })

    hc.Connect("clicked", func() bool {
        u.Commands.Names["help"].Execute()
        return true
    })

    rbkmk.Connect("clicked", func() bool {
        s := m.Spreads[m.SpreadIndex]
        if m.Direction == model.RTL {
            if len(s.Pages) > 1 {
                if s.RectoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
            u.Commands.Names["toggleBookmark"].Execute()
        } else {
            if len(s.Pages) > 1 {
                if s.VersoPage() == m.PageIndex {
                    u.Commands.Names["selectPage"].Execute()
                }
            }
            u.Commands.Names["toggleBookmark"].Execute()
        }
        return true
    })

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
    container.Attach(spn, 1, 0, 1, 1)
    container.Attach(fc, 2, 0, 1, 1)
    container.Attach(hc, 3, 0, 1, 1)
    container.Attach(rbkmk, 4, 0, 1, 1)
    container.SetSizeRequest(1000, 8)

    c.leftBookmark = lbkmk
    c.spinner = spn
    c.fileControl = fc
    c.helpControl = hc
    c.rightBookmark = rbkmk
    c.container = container
    return c
}

func (c *PageViewHdrControl) Render(m *model.Model) {
    css, _ := c.leftBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    css, _ = c.rightBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    c.fileControl.SetLabel("File")

    if m.Loading {
        c.spinner.Start()
    } else {
        c.spinner.Stop()
    }

    if len(m.Spreads) < 1 || m.Bookmarks == nil {
        return
    } else {
        lbkmkcss, _ := c.leftBookmark.GetStyleContext()
        rbkmkcss, _ := c.rightBookmark.GetStyleContext()
        spread := m.Spreads[m.SpreadIndex]
        title := strings.TrimSuffix(filepath.Base(m.FilePath), filepath.Ext(m.FilePath))

        if m.Direction == model.RTL {
            if len(spread.Pages) > 1 {
                b := m.Bookmarks.Find(spread.VersoPage())
                if b != nil {
                    rbkmkcss.AddClass("marked")
                }

                b = m.Bookmarks.Find(spread.RectoPage())
                if b != nil {
                    lbkmkcss.AddClass("marked")
                }
            } else {
                rbkmkcss.AddClass("transparent")
                b := m.Bookmarks.Find(spread.VersoPage())
                if b != nil {
                    lbkmkcss.AddClass("marked")
                }
            }
        } else {
            if len(spread.Pages) > 1 {
                b := m.Bookmarks.Find(spread.VersoPage())
                if b != nil {
                    lbkmkcss.AddClass("marked")
                }

                b = m.Bookmarks.Find(spread.RectoPage())
                if b != nil {
                    rbkmkcss.AddClass("marked")
                }
            } else {
                lbkmkcss.AddClass("transparent")
                b := m.Bookmarks.Find(spread.VersoPage())
                if b != nil {
                    rbkmkcss.AddClass("marked")
                }
            }
        }
        c.fileControl.SetLabel(title)
    }
}

