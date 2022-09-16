package main

import (
	"fmt"
	"os"

	_ "github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"

    "example.com/cbxv-gotk3/internal/model"
    "example.com/cbxv-gotk3/internal/ui"
    "example.com/cbxv-gotk3/internal/util"
)

// Simple cbx application with a gui provided by gtk

// Some data loading utils
func loadCbxFile(m *model.Model, sendMessage util.Messenger) error {
    td, err := util.CreateTmpDir()
    if err != nil {
        return err
    }
    m.TmpDir = td

    ip, err := util.GetImagePaths(m.FilePath, m.TmpDir)
    if err != nil {
        return err
    }
    m.ImgPaths = ip
    m.NewPages()
    m.NewLeaves()
    m.CurrentLeaf = 0
    m.SelectedPage = 0
    m.RefreshPages()

    sendMessage(util.Message{TypeName: "render"})

    return nil
}

func closeCbxFile(m *model.Model) {
    os.RemoveAll(m.TmpDir)
    m.ImgPaths = nil
    m.Pages = nil
    m.Leaves = nil
    m.CurrentLeaf = 0
    m.SelectedPage = 0
    m.Bookmarks = nil
    m.SeriesList = nil
    m.SeriesIndex = 0
}

func loadSeriesList(m *model.Model) {
    s, err := util.ReadSeriesList(m.FilePath)
    if err != nil {
        fmt.Printf("Unable to load series list %s\n", err)
    }
    m.SeriesList = s

    for i := range s {
        if m.FilePath == s[i] {
            m.SeriesIndex = i
        }
    }
    m.SelectedPage = m.CalcVersoPage()
}

func quit() {
    gtk.MainQuit()
}

// Update listens for message on the message channel and
// handles messages by invoking commands which update the model
func update(model *model.Model, u *ui.UI, msgChan chan util.Message, commands *CommandList) {
    for m := range msgChan {
        cmd := commands.Commands[m.TypeName]
        if model.Leaves == nil && (
            m.TypeName != "quit" &&
            m.TypeName != "openFile") {
            continue
        }
        if cmd != nil {
            glib.IdleAdd(func(){
                cmd(m.Data)
                ui.Render(model, u)
            })
        }
    }
}

// Setup the model
// Setup the ui
// Create commands to modify the model
// Start the update message handler
// Open the main window, when it closes
// Shutdown UI threads
// Exit
func main() {
    msgChan := make(chan util.Message)
    messenger := func (m util.Message) { msgChan <- m }
    model := model.NewModel(messenger)
    u := &ui.UI{}
    commands := NewCommands(model)

    ui.InitUI(model, u, messenger)

    go update(model, u, msgChan, commands)

//    u.MainWindow.ShowAll()

    if len(os.Args) > 1 {
        commands.Commands["openFile"](os.Args[1])
    }

    gtk.Main()
    ui.StopUI(model, u)
}

