package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"

    "example.com/cbxv-gotk3/internal/util"
    "example.com/cbxv-gotk3/internal/model"
)

const (
    APP_HLP_ICN = "?"
)

type HdrControl struct {
    container *gtk.Grid
    leftBookmark *gtk.Label
    spinner *gtk.Spinner
    title *gtk.Button
    helpControl *gtk.Button
    rightBookmark *gtk.Label
}

func NewHdrControl(m *model.Model, u *UI) *HdrControl {
    c := &HdrControl{}

    lbkmk, err := gtk.LabelNew("")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    lbkmk.SetHAlign(gtk.ALIGN_START)
    lbkmk.SetHExpand(true)
    css, _ := lbkmk.GetStyleContext()
	css.AddClass("bkmk-btn")

	spn, err := gtk.SpinnerNew()
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    css, _ = spn.GetStyleContext()
	css.AddClass("nav-btn")

	t, err := gtk.ButtonNewWithLabel("Untitled")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    t.SetTooltipText("Open File")
    css, _ = t.GetStyleContext()
	css.AddClass("nav-btn")

	hc, err := gtk.ButtonNewWithLabel(APP_HLP_ICN)
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    t.SetTooltipText("Help")
    css, _ = hc.GetStyleContext()
	css.AddClass("nav-btn")

	rbkmk, err := gtk.LabelNew("")
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
    css, _ = rbkmk.GetStyleContext()
	css.AddClass("bkmk-btn")

    t.Connect("clicked", func() bool { 
        fmt.Printf("openFile\n")
        // fixme: this code just copy/pasted from UI
        // should add a concept of UICommand
        dlg, _ := gtk.FileChooserNativeDialogNew("Open", u.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
        dlg.SetCurrentFolder(m.BrowseDirectory)
        output := dlg.NativeDialog.Run()
        if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
            f := dlg.GetFilename()
            m := &util.Message{TypeName: "openFile", Data: f}
            u.sendMessage(*m)
        }
        u.initCanvas(m)
        return true
    })

    hc.Connect("clicked", func() bool { 
        // fixme: this code just copy/pasted from UI
        // should add a concept of UICommand
        dlg := gtk.MessageDialogNewWithMarkup(u.mainWindow, 
            gtk.DialogFlags(gtk.DIALOG_MODAL), 
            gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, "Help")
        dlg.SetTitle("Help")
        dlg.SetMarkup(util.HELP_TXT)
        css, _ := dlg.GetStyleContext()
        css.AddClass("msg-dlg")
        dlg.Run()
        dlg.Destroy()
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
    container.Attach(t, 2, 0, 1, 1)
    container.Attach(hc, 3, 0, 1, 1)
    container.Attach(rbkmk, 4, 0, 1, 1)
	container.SetSizeRequest(1000, 8)
    c.leftBookmark = lbkmk
    c.spinner = spn
    c.title = t
    c.helpControl = hc
    c.rightBookmark = rbkmk
    c.container = container
    return c
}

func (c *HdrControl) Render(m *model.Model) {
    vpn := m.CalcVersoPage()
    css, _ := c.leftBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    css, _ = c.rightBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    c.title.SetLabel("Untitled")

    if m.Loading {
        c.spinner.Start()
    } else {
        c.spinner.Stop()
    }

    if len(m.Leaves) < 1 || m.Bookmarks == nil {
        return 
    } else {
        lbkmkcss, _ := c.leftBookmark.GetStyleContext()
        rbkmkcss, _ := c.rightBookmark.GetStyleContext()
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
        c.title.SetLabel(title)
    }
}

