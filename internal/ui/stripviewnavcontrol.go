package ui

import (
    "fmt"

    "github.com/gotk3/gotk3/glib"
    "github.com/gotk3/gotk3/gtk"
    "github.com/mftb0/cbxv-gotk3/internal/model"
    "github.com/mftb0/cbxv-gotk3/internal/util"
)

type StripViewNavControl struct {
    ui                *UI
    container         *gtk.Grid
    progName          *gtk.Label
    progVersion       *gtk.Label
    layoutModeControl *gtk.Label
    hpcSignalHandle   *glib.SignalHandle
    fullscreenControl *gtk.Button
}

func NewStripViewNavControl(m *model.Model, u *UI) *StripViewNavControl {
    nc := &StripViewNavControl{}
    nc.ui = u

    pn := util.CreateLabel("cbxv", "nav-btn", nil)
    pn.SetHAlign(gtk.ALIGN_START)
    pn.SetHExpand(false)

    pv := util.CreateLabel("v0.0.1", "nav-btn", nil)
    pv.SetHAlign(gtk.ALIGN_START)
    pv.SetHExpand(true)

    lmc := util.CreateLabel("Layout", "nav-btn", util.S("Layout"))

    fsc := util.CreateButton(util.FullscreenIcon(), "nav-btn", util.S("Fullscreen Toggle"))

    container, err := gtk.GridNew()
    if err != nil {
        fmt.Printf("Error creating grid %s\n", err)
    }
    container.SetHAlign(gtk.ALIGN_CENTER)
    container.SetVAlign(gtk.ALIGN_END)
    container.SetHExpand(true)
    css, _ := container.GetStyleContext()
    css.AddClass("nav-ctrl")

    fsc.Connect("clicked", func() {
        if m.Fullscreen {
            u.MainWindow.Unfullscreen()
        } else {
            u.MainWindow.Fullscreen()
        }
        u.SendMessage(util.Message{TypeName: "toggleFullscreen"})
    })

    container.Attach(pn, 0, 0, 1, 1)
    container.Attach(pv, 1, 0, 1, 1)
    container.Attach(lmc, 2, 0, 1, 1)
    container.Attach(fsc, 3, 0, 1, 1)
    container.SetSizeRequest(1024, 8)
    nc.container = container
    nc.progName = pn
    nc.progVersion = pv
    nc.layoutModeControl = lmc
    nc.fullscreenControl = fsc

    return nc
}

func (c *StripViewNavControl) Render(m *model.Model) {
    if len(m.Spreads) < 1 {
        if m.LayoutMode == model.ONE_PAGE {
            c.layoutModeControl.SetText("1-Page")
        } else if m.LayoutMode == model.TWO_PAGE {
            c.layoutModeControl.SetText("2-Page")
        } else {
            c.layoutModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.fullscreenControl.SetLabel(util.FullscreenIcon())
        } else {
            c.fullscreenControl.SetLabel(util.RestoreIcon())
        }

        return
    } else {
        if m.LayoutMode == model.ONE_PAGE {
            c.layoutModeControl.SetText("1-Page")
        } else if m.LayoutMode == model.TWO_PAGE {
            c.layoutModeControl.SetText("2-Page")
        } else {
            c.layoutModeControl.SetText("Strip")
        }

        if m.Fullscreen {
            c.container.SetSizeRequest(1400, 8)
            c.fullscreenControl.SetLabel(util.RestoreIcon())
        } else {
            c.container.SetSizeRequest(1024, 8)
            c.fullscreenControl.SetLabel(util.FullscreenIcon())
        }
    }
}
