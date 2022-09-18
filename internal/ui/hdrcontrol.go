package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"
    "example.com/cbxv-gotk3/internal/model"
)

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

func (c *HdrControl) Render(m *model.Model) {
    vpn := m.CalcVersoPage()
    css, _ := c.leftBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    css, _ = c.rightBookmark.GetStyleContext()
    css.RemoveClass("marked")
    css.RemoveClass("transparent")
    c.title.SetText("")
    if len(m.Leaves) < 1 || m.Bookmarks == nil {
        c.title.SetText("Loading...")
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
        c.title.SetText(title)
    }
}

