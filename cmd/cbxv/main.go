package main

import (
	"fmt"
	"os"

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
    //noop currently called in case app ever wants to take
    //action before ui shutsdown
}

// Update listens for message on the message channel and
// handles messages by invoking commands which update the model
func update(m *model.Model, u *ui.UI, msgChan chan util.Message, commands *CommandList) {
    for msg := range msgChan {
        cmd := commands.Commands[msg.TypeName]
        if m.Leaves == nil && (
            msg.TypeName != "quit" &&
            msg.TypeName != "openFile") {
            continue
        }
        if cmd != nil {
            u.RunFunc(func(){
                cmd(msg.Data)
                u.Render(m)
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
    m := model.NewModel(messenger)
    commands := NewCommands(m)

    u := ui.NewUI(m, messenger)

    go update(m, u, msgChan, commands)

    if len(os.Args) > 1 {
        commands.Commands["openFile"](os.Args[1])
    }

    u.Run()
    u.Dispose()
}

