package main

import (
    "os"

    "github.com/mftb0/cbxv-gotk3/internal/model"
    "github.com/mftb0/cbxv-gotk3/internal/ui"
    "github.com/mftb0/cbxv-gotk3/internal/util"
)

// Simple cbx application with a gui provided by gtk

const (
    NAME    = "cbxv"
    VERSION = "0.0.17"
)

// Update listens for message on the message channel and
// handles messages by invoking commands which update the model
func update(m *model.Model, u *ui.UI, msgChan chan util.Message, commands *CommandList) {
    for msg := range msgChan {
        cmd := commands.Commands[msg.TypeName]
        if m.Spreads == nil && (msg.TypeName != "quit" &&
            msg.TypeName != "openFile") {
            continue
        }
        if cmd != nil {
            u.RunFunc(func() {
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
    md := model.ProgramMetadata{Name: NAME, Version: VERSION}
    messenger := func(m util.Message) { msgChan <- m }
    m := model.NewModel(md, messenger)
    commands := NewCommands(m)

    u := ui.NewUI(m, messenger)

    go update(m, u, msgChan, commands)

    u.RunFunc(func() {
        //default to 2-page display
        commands.Commands["setLayoutModeTwoPage"]("")
        if len(os.Args) > 1 {
            commands.Commands["openFile"](os.Args[1])
        }
    })

    u.Run()

    // At the end of all things the closeFile command
    // can't work because we've orchestrated for it only
    // to be run on the UI thread and the UI thread is dead so
    // we have to close the last cbx file here, to get
    // rid of any tmpDir
    m.CloseCbxFile()
}
