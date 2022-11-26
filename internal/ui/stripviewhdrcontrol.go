package ui

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/gotk3/gotk3/gtk"

    "github.com/mftb0/cbxv-gotk3/internal/model"
    "github.com/mftb0/cbxv-gotk3/internal/util"
)

type StripViewHdrControl struct {
    container   *gtk.Grid
    spinner     *gtk.Spinner
    fileControl *gtk.Button
    helpControl *gtk.Button
}

func NewStripViewHdrControl(m *model.Model, u *UI) *StripViewHdrControl {
    c := &StripViewHdrControl{}

    spn, err := gtk.SpinnerNew()
    if err != nil {
        fmt.Printf("Error creating label %s\n", err)
    }
    css, _ := spn.GetStyleContext()
    css.AddClass("nav-btn")

    fc := util.CreateButton("File", "nav-btn", util.S("Open File"))
    hc := util.CreateButton(APP_HLP_ICN, "nav-btn", util.S("Help"))

    fc.Connect("clicked", func() bool {
        u.Commands.Names["fileOpen"].Execute()
        return true
    })

    hc.Connect("clicked", func() bool {
        u.Commands.Names["help"].Execute()
        return true
    })

    container, err := gtk.GridNew()
    if err != nil {
        fmt.Printf("Error creating label %s\n", err)
    }
    container.SetHAlign(gtk.ALIGN_END)
    container.SetVAlign(gtk.ALIGN_START)
    container.SetHExpand(true)
    css, _ = container.GetStyleContext()
    css.AddClass("hdr-ctrl")
    container.Attach(spn, 0, 0, 1, 1)
    container.Attach(fc, 1, 0, 1, 1)
    container.Attach(hc, 2, 0, 1, 1)
    container.SetSizeRequest(64, 32)
    c.spinner = spn
    c.fileControl = fc
    c.helpControl = hc
    c.container = container
    return c
}

func (c *StripViewHdrControl) Render(m *model.Model) {
    c.fileControl.SetLabel("File")

    if m.Loading {
        c.spinner.Start()
    } else {
        c.spinner.Stop()
    }

    if len(m.Spreads) < 1 || m.Bookmarks == nil {
        return
    } else {
        title := strings.TrimSuffix(filepath.Base(m.FilePath), filepath.Ext(m.FilePath))
        c.fileControl.SetLabel(title)
    }
}

