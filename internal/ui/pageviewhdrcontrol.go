package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"example.com/cbxv-gotk3/internal/model"
	"example.com/cbxv-gotk3/internal/util"
)

const (
	APP_HLP_ICN = "?"
)

type PageViewHdrControl struct {
	container     *gtk.Grid
	leftBookmark  *gtk.Label
	spinner       *gtk.Spinner
	title         *gtk.Button
	helpControl   *gtk.Button
	rightBookmark *gtk.Label
}

func NewHdrControl(m *model.Model, u *UI) *PageViewHdrControl {
	c := &PageViewHdrControl{}

	lbkmk := util.CreateLabel("", "bkmk-btn", nil)
	lbkmk.SetHAlign(gtk.ALIGN_START)
	lbkmk.SetHExpand(true)

	spn, err := gtk.SpinnerNew()
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ := spn.GetStyleContext()
	css.AddClass("nav-btn")

	t := util.CreateButton("Untitled", "nav-btn", util.S("Open File"))
	hc := util.CreateButton(APP_HLP_ICN, "nav-btn", util.S("Help"))
	rbkmk := util.CreateLabel("", "bkmk-btn", nil)

	t.Connect("clicked", func() bool {
		// fixme: this code just copy/pasted from UI
		// should add a concept of UICommand
		dlg, _ := gtk.FileChooserNativeDialogNew("Open", u.mainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
		dlg.SetCurrentFolder(m.BrowseDir)
		output := dlg.NativeDialog.Run()
		if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
			f := dlg.GetFilename()
			m := &util.Message{TypeName: "openFile", Data: f}
			u.sendMessage(*m)
		}
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

func (c *PageViewHdrControl) Render(m *model.Model) {
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

	if len(m.Spreads) < 1 || m.Bookmarks == nil {
		return
	} else {
		lbkmkcss, _ := c.leftBookmark.GetStyleContext()
		rbkmkcss, _ := c.rightBookmark.GetStyleContext()
		spread := m.Spreads[m.SpreadIndex]
		title := strings.TrimSuffix(filepath.Base(m.FilePath), filepath.Ext(m.FilePath))

		if m.Direction == model.RTL {
			b := m.Bookmarks.Find(spread.VersoPage())
			if b != nil {
				if len(spread.Pages) > 1 {
					rbkmkcss.AddClass("marked")
				} else {
					rbkmkcss.AddClass("transparent")
					lbkmkcss.AddClass("marked")
				}
			}

			if len(spread.Pages) > 1 {
				b = m.Bookmarks.Find(spread.RectoPage())
				if b != nil {
					lbkmkcss.AddClass("marked")
				}
			}
		} else {
			b := m.Bookmarks.Find(spread.VersoPage())
			if b != nil {
				if len(spread.Pages) > 1 {
					lbkmkcss.AddClass("marked")
				} else {
					lbkmkcss.AddClass("transparent")
					rbkmkcss.AddClass("marked")
				}
			}

			if len(spread.Pages) > 1 {
				b = m.Bookmarks.Find(spread.RectoPage())
				if b != nil {
					rbkmkcss.AddClass("marked")
				}
			}
		}
		c.title.SetLabel(title)
	}
}
