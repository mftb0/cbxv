package ui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/gotk3/gotk3/gtk"

	"example.com/cbxv-gotk3/internal/model"
	"example.com/cbxv-gotk3/internal/util"
)

type StripViewHdrControl struct {
	container     *gtk.Grid
	spinner       *gtk.Spinner
	title         *gtk.Button
	helpControl   *gtk.Button
}

func NewStripViewHdrControl(m *model.Model, u *UI) *StripViewHdrControl {
	c := &StripViewHdrControl{}

	spn, err := gtk.SpinnerNew()
	if err != nil {
		fmt.Printf("Error creating label %s\n", err)
	}
	css, _ := spn.GetStyleContext()
	css.AddClass("nav-btn")

	t := util.CreateButton("Untitled", "nav-btn", util.S("Open File"))
	hc := util.CreateButton(APP_HLP_ICN, "nav-btn", util.S("Help"))

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
	container.SetHAlign(gtk.ALIGN_END)
	container.SetVAlign(gtk.ALIGN_START)
	container.SetHExpand(true)
	css, _ = container.GetStyleContext()
	css.AddClass("hdr-ctrl")
	container.Attach(spn, 0, 0, 1, 1)
	container.Attach(t, 1, 0, 1, 1)
	container.Attach(hc, 2, 0, 1, 1)
	container.SetSizeRequest(64, 32)
	c.spinner = spn
	c.title = t
	c.helpControl = hc
	c.container = container
	return c
}

func (c *StripViewHdrControl) Render(m *model.Model) {
	c.title.SetLabel("Untitled")

	if m.Loading {
		c.spinner.Start()
	} else {
		c.spinner.Stop()
	}

	if len(m.Spreads) < 1 || m.Bookmarks == nil {
		return
	} else {
		title := strings.TrimSuffix(filepath.Base(m.FilePath), filepath.Ext(m.FilePath))
		c.title.SetLabel(title)
	}
}
