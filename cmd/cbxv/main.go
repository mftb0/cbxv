package main

import (
	"os"

    "example.com/cbxv-gotk3/internal/model"
    "example.com/cbxv-gotk3/internal/ui"
    "example.com/cbxv-gotk3/internal/util"
)

// Simple cbx application with a gui provided by gtk

// Update listens for message on the message channel and
// handles messages by invoking commands which update the model
func update(m *model.Model, u *ui.UI, msgChan chan util.Message, commands *CommandList) {
    for msg := range msgChan {
        cmd := commands.Commands[msg.TypeName]
        if m.Spreads == nil && (
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
        u.RunFunc(func(){
            commands.Commands["openFile"](os.Args[1])
        })
    }

    u.Run()

    // At the end of all things the closeFile command
    // can't work because we've orchestrated for it only
    // to be run on the UI thread and the UI thread is dead so
    // we have to close the last cbx file here, to get
    // rid of any tmpDir
    m.CloseCbxFile()
}

