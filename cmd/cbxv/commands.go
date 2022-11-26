package main

import (
	"path/filepath"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mftb0/cbxv-gotk3/internal/model"
	"github.com/mftb0/cbxv-gotk3/internal/ui"
	"github.com/mftb0/cbxv-gotk3/internal/util"
)

type Environment struct {
	Vars map[string]string
	Args []string
}

type Command struct {
	Name        string
	DisplayName string
	BindKeys    []uint
	Env         *Environment
	Execute     func()
}

type CommandList struct {
	Names    map[string]func()
	KeyCodes map[uint]func()
}

func NewCommands(m *model.Model, u *ui.UI) *CommandList {
	cmds := &CommandList{Names: make(map[string]func())}

	cmd := Command{
		Name:        "rightPage",
		DisplayName: "Right Page",
		BindKeys:    []uint{gdk.KEY_d, gdk.KEY_Right, gdk.KEY_l},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "leftPage",
		DisplayName: "Left Page",
		BindKeys:    []uint{gdk.KEY_a, gdk.KEY_Left, gdk.KEY_h},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "firstPage",
		DisplayName: "First Page",
		BindKeys:    []uint{gdk.KEY_w, gdk.KEY_Up, gdk.KEY_k},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		switch v := u.View.(type) {
		case *ui.StripView:
			v.ScrollToTop()
		default:
			u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
		}
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "lastPage",
		DisplayName: "Last Page",
		BindKeys:    []uint{gdk.KEY_s, gdk.KEY_Down, gdk.KEY_j},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		switch v := u.View.(type) {
		case *ui.StripView:
			v.ScrollToBottom()
		default:
			u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
		}
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "lastBookmark",
		DisplayName: "Last Bookmark",
		BindKeys:    []uint{gdk.KEY_L},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "selectPage",
		DisplayName: "Select Page",
		BindKeys:    []uint{gdk.KEY_Tab},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "setLayoutModeOnePage",
		DisplayName: "Layout Mode One Page",
		BindKeys:    []uint{gdk.KEY_1},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.View.Disconnect(m, u)
		u.View = u.PageView
		u.View.Connect(m, u)
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "setLayoutModeTwoPage",
		DisplayName: "Layout Mode Two Page",
		BindKeys:    []uint{gdk.KEY_2},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.View.Disconnect(m, u)
		u.View = u.PageView
		u.View.Connect(m, u)
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "setLayoutModeLongStrip",
		DisplayName: "Layout Mode Long Strip",
		BindKeys:    []uint{gdk.KEY_3},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.View.Disconnect(m, u)
		u.View = u.StripView
		u.View.Connect(m, u)
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "toggleDirection",
		DisplayName: "Toggle Read Mode",
		BindKeys:    []uint{gdk.KEY_grave},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "toggleFullscreen",
		DisplayName: "Toggle Fullscreen",
		BindKeys:    []uint{gdk.KEY_f, gdk.KEY_F11},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		if m.Fullscreen {
			u.MainWindow.Unfullscreen()
		} else {
			u.MainWindow.Fullscreen()
		}
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "openFile",
		DisplayName: "Open File",
		BindKeys:    []uint{gdk.KEY_o},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		dlg, _ := gtk.FileChooserNativeDialogNew("Open",
			u.MainWindow, gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", "_Cancel")
		defer dlg.Destroy()

		dlg.SetCurrentFolder(m.BrowseDir)
		fltr, _ := gtk.FileFilterNew()
		fltr.AddPattern("*.cbz")
		fltr.AddPattern("*.cbr")
		fltr.AddPattern("*.cb7")
		fltr.AddPattern("*.cbt")
		fltr.SetName("cbx files")
		dlg.AddFilter(fltr)
		fltr, _ = gtk.FileFilterNew()
		fltr.AddPattern("*")
		fltr.SetName("All files")
		dlg.AddFilter(fltr)

		output := dlg.NativeDialog.Run()
		if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
			f := dlg.GetFilename()
			m := &util.Message{TypeName: cmd.Env.Args[0], Data: f}
			u.SendMessage(*m)
		}
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "closeFile",
		DisplayName: "Close File",
		BindKeys:    []uint{gdk.KEY_c},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		if u.View.(*ui.StripView) == u.StripView {
			u.View = nil
			u.StripView = nil
			u.StripView = ui.NewStripView(m, u, u.SendMessage)
			u.View = u.StripView
		}
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "nextFile",
		DisplayName: "Next File",
		BindKeys:    []uint{gdk.KEY_n},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "previousFile",
		DisplayName: "Previous File",
		BindKeys:    []uint{gdk.KEY_p},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "exportPage",
		DisplayName: "Export Page",
		BindKeys:    []uint{gdk.KEY_e},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		dlg, _ := gtk.FileChooserNativeDialogNew("Save",
			u.MainWindow, gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", "_Cancel")
		defer dlg.Destroy()

		base := filepath.Base(m.Pages[m.PageIndex].FilePath)
		dlg.SetCurrentFolder(m.ExportDir)
		dlg.SetCurrentName(base)
		dlg.SetDoOverwriteConfirmation(true)

		output := dlg.NativeDialog.Run()
		if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
			f := dlg.GetFilename()
			m := &util.Message{TypeName: cmd.Env.Args[0], Data: f}
			u.SendMessage(*m)
		}
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "toggleBookmark",
		DisplayName: "Toggle Bookmark",
		BindKeys:    []uint{gdk.KEY_space},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "toggleJoin",
		DisplayName: "toggleJoin",
		BindKeys:    []uint{gdk.KEY_r},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "hidePage",
		DisplayName: "Hide Page",
		BindKeys:    []uint{gdk.KEY_minus},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: cmd.Env.Args[0]})
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "showPage",
		DisplayName: "Show Page",
		BindKeys:    []uint{},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "loadAllPages",
		DisplayName: "Load All Pages",
		BindKeys:    []uint{},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "render",
		DisplayName: "Render",
		BindKeys:    []uint{},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "help",
		DisplayName: "Help",
		BindKeys:    []uint{gdk.KEY_question, gdk.KEY_F1},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		dlg := gtk.MessageDialogNewWithMarkup(u.MainWindow,
			gtk.DialogFlags(gtk.DIALOG_MODAL),
			gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, "Help")
		defer dlg.Destroy()

		dlg.SetTitle("Help")
		dlg.SetMarkup(util.HELP_TXT)
		css, _ := dlg.GetStyleContext()
		css.AddClass("msg-dlg")

		dlg.Run()
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "sendMessage",
		DisplayName: "Send Message",
		BindKeys:    []uint{},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		if len(cmd.Env.Args) > 1 {
			u.SendMessage(util.Message{TypeName: cmd.Env.Args[0], Data: cmd.Env.Args[1]})
		}
	}
	AddCommand(cmds, cmd)

	cmd = Command{
		Name:        "quit",
		DisplayName: "Quit",
		BindKeys:    []uint{gdk.KEY_q},
		Env:         &Environment{Vars: make(map[string]string)},
	}
	cmd.Env.Args[0] = cmd.Name
	cmd.Execute = func() {
		u.SendMessage(util.Message{TypeName: "quit"})
		u.Quit()
	}
	AddCommand(cmds, cmd)

	return cmds
}

func AddCommand(commands *CommandList, command Command) {
	commands.Names[command.Name] = command.Execute
	for i := range command.BindKeys {
		k := command.BindKeys[i]
		commands.KeyCodes[k] = command.Execute
	}
}

