package main

import (
	"os"

    "example.com/cbxv-gotk3/internal/model"
    "example.com/cbxv-gotk3/internal/ui"
    "example.com/cbxv-gotk3/internal/util"
)

// Simple cbx application with a gui provided by gtk

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

