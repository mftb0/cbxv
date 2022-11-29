package ui

import (
	"path/filepath"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/gtk"
	"github.com/mftb0/cbxv-gotk3/internal/model"
	"github.com/mftb0/cbxv-gotk3/internal/util"
)

type Command struct {
	Name        string
	DisplayName string
	BindKeys    []uint
	Execute      func()
}

func NewCommand(name string, displayName string, bindKeys []uint, callback func()) *Command {
	cmd := &Command{
		Name:        name,
		DisplayName: displayName,
		BindKeys:    bindKeys,
	}
	cmd.Execute = callback
    return cmd
}

type CommandList struct {
	Names    map[string]*Command
	KeyCodes map[uint]*Command
}

func NewCommands(m *model.Model, u *UI) *CommandList {
	cmds := &CommandList{Names: make(map[string]*Command), KeyCodes: make(map[uint]*Command)}

    AddCommand(cmds, NewCommand("rightPage", "Right Page", 
        []uint{gdk.KEY_d, gdk.KEY_Right, gdk.KEY_l}, 
        func() {
		    u.SendMessage(util.Message{TypeName: "rightPage"})
	    }))

    AddCommand(cmds, NewCommand("leftPage", "Left Page", 
        []uint{gdk.KEY_a, gdk.KEY_Left, gdk.KEY_h}, 
        func() {
		    u.SendMessage(util.Message{TypeName: "leftPage"})
	    }))

    AddCommand(cmds, NewCommand("firstPage", "First Page",
		[]uint{gdk.KEY_w, gdk.KEY_Up, gdk.KEY_k},
        func() {
            switch v := u.View.(type) {
            case *StripView:
                v.ScrollToTop()
            default:
                u.SendMessage(util.Message{TypeName: "firstPage"})
            }
	    }))

    AddCommand(cmds, NewCommand("lastPage", "Last Page",
		[]uint{gdk.KEY_s, gdk.KEY_Down, gdk.KEY_j},
	    func() {
            switch v := u.View.(type) {
            case *StripView:
                v.ScrollToBottom()
            default:
                u.SendMessage(util.Message{TypeName: "lastPage"})
            }
        }))

    AddCommand(cmds, NewCommand("lastBookmark", "Last Bookmark",
		[]uint{gdk.KEY_L},
	    func() {
		    u.SendMessage(util.Message{TypeName: "lastBookmark"})
	    }))

    AddCommand(cmds, NewCommand("selectPage", "Select Page",
		[]uint{gdk.KEY_Tab},
	    func() {
		    u.SendMessage(util.Message{TypeName: "selectPage"})
	    }))

    AddCommand(cmds, NewCommand("setLayoutModeOnePage", "Layout Mode One Page",
		[]uint{gdk.KEY_1},
	    func() {
            u.View.Disconnect(m, u)
            u.View = u.PageView
            u.View.Connect(m, u)
            u.SendMessage(util.Message{TypeName: "setLayoutModeOnePage"})
        }))

    AddCommand(cmds, NewCommand("setLayoutModeTwoPage", "Layout Mode Two Page",
		[]uint{gdk.KEY_2},
	    func() {
            u.View.Disconnect(m, u)
            u.View = u.PageView
            u.View.Connect(m, u)
            u.SendMessage(util.Message{TypeName: "setLayoutModeTwoPage"})
        }))

    AddCommand(cmds, NewCommand("setLayoutModeLongStrip", "Layout Mode Long Strip",
		[]uint{gdk.KEY_3},
        func() {
            u.View.Disconnect(m, u)
            u.View = u.StripView
            u.View.Connect(m, u)
            u.SendMessage(util.Message{TypeName: "setLayoutModeLongStrip"})
        }))

    AddCommand(cmds, NewCommand("toggleDirection", "Toggle Read Mode",
		[]uint{gdk.KEY_grave},
	    func() {
		    u.SendMessage(util.Message{TypeName: "toggleDirection"})
	    }))

    AddCommand(cmds, NewCommand("toggleFullscreen", "Toggle Fullscreen",
		[]uint{gdk.KEY_f, gdk.KEY_F11},
	    func() {
            if m.Fullscreen {
                u.MainWindow.Unfullscreen()
            } else {
                u.MainWindow.Fullscreen()
            }
            u.SendMessage(util.Message{TypeName: "toggleFullscreen"})
        }))

    AddCommand(cmds, NewCommand("openFile", "Open File",
		[]uint{gdk.KEY_o},
	    func() {
            dlg, _ := gtk.FileChooserDialogNewWith2Buttons("Open", u.MainWindow,
                gtk.FILE_CHOOSER_ACTION_OPEN, "_Open", gtk.RESPONSE_ACCEPT,
                "_Cancel", gtk.RESPONSE_CANCEL)
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

            output := dlg.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "openFile", Data: f}
                u.SendMessage(*m)
            }
        }))

    AddCommand(cmds, NewCommand("closeFile", "Close File",
		[]uint{gdk.KEY_c},
	    func() {
            u.SendMessage(util.Message{TypeName: "closeFile"})
        }))

    AddCommand(cmds, NewCommand("nextFile", "Next File",
		[]uint{gdk.KEY_n},
        func() {
            u.SendMessage(util.Message{TypeName: "nextFile"})
        }))

    AddCommand(cmds, NewCommand("previousFile", "Previous File",
		[]uint{gdk.KEY_p},
        func() {
            u.SendMessage(util.Message{TypeName: "previousFile"})
        }))

    AddCommand(cmds, NewCommand("exportPage", "Export Page",
		[]uint{gdk.KEY_e},
        func() {
            dlg, _ := gtk.FileChooserDialogNewWith2Buttons("Save", u.MainWindow,
                gtk.FILE_CHOOSER_ACTION_SAVE, "_Save", gtk.RESPONSE_ACCEPT,
                "_Cancel", gtk.RESPONSE_CANCEL)
            defer dlg.Destroy()

            base := filepath.Base(m.Pages[m.PageIndex].FilePath)
            dlg.SetCurrentFolder(m.ExportDir)
            dlg.SetCurrentName(base)
            dlg.SetDoOverwriteConfirmation(true)

            output := dlg.Run()
            if gtk.ResponseType(output) == gtk.RESPONSE_ACCEPT {
                f := dlg.GetFilename()
                m := &util.Message{TypeName: "exportPage", Data: f}
                u.SendMessage(*m)
            }
        }))

    AddCommand(cmds, NewCommand("toggleBookmark", "Toggle Bookmark",
		[]uint{gdk.KEY_space},
        func() {
            u.SendMessage(util.Message{TypeName: "toggleBookmark"})
        }))

    AddCommand(cmds, NewCommand("toggleJoin", "toggleJoin",
		[]uint{gdk.KEY_r},
        func() {
            u.SendMessage(util.Message{TypeName: "toggleJoin"})
        }))

    AddCommand(cmds, NewCommand("hidePage", "Hide Page",
		[]uint{gdk.KEY_minus},
        func() {
            u.SendMessage(util.Message{TypeName: "hidePage"})
        }))

    AddCommand(cmds, NewCommand("showPage", "Show Page",
		[]uint{},
        func() {
            u.SendMessage(util.Message{TypeName: "showPage"})
        }))

    AddCommand(cmds, NewCommand("loadAllPages", "Load All Pages",
		[]uint{},
        func() {
            u.SendMessage(util.Message{TypeName: "loadAllPages"})
        }))

    AddCommand(cmds, NewCommand("render", "Render",
		[]uint{},
        func() {
        }))

    AddCommand(cmds, NewCommand("help", "Help",
		[]uint{gdk.KEY_question, gdk.KEY_F1},
        func() {
            dlg := gtk.MessageDialogNewWithMarkup(u.MainWindow,
                gtk.DialogFlags(gtk.DIALOG_MODAL),
                gtk.MESSAGE_INFO, gtk.BUTTONS_CLOSE, "Help")
            defer dlg.Destroy()

            dlg.SetTitle("Help")
            dlg.SetMarkup(util.HELP_TXT)
            css, _ := dlg.GetStyleContext()
            css.AddClass("msg-dlg")

            dlg.Run()
        }))

    AddCommand(cmds, NewCommand("quit", "Quit",
		[]uint{gdk.KEY_q},
        func() {
            u.SendMessage(util.Message{TypeName: "quit"})
            u.Quit()
        }))

	return cmds
}

func AddCommand(commands *CommandList, command *Command) {
	commands.Names[command.Name] = command
	for i := range command.BindKeys {
		k := command.BindKeys[i]
		commands.KeyCodes[k] = command
	}
}

