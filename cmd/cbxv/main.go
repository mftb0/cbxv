package main

import (
	"os"
	"runtime"

	"github.com/mftb0/cbxv/internal/model"
	"github.com/mftb0/cbxv/internal/ui"
	"github.com/mftb0/cbxv/internal/util"
)

// Simple cbx application with a gui provided by gtk

const (
    NAME    = "cbxv"
    VERSION = "0.5.0"
)

// Update listens for messages on the message channel and
// handles messages by invoking messageHandlers
func update(m *model.Model, u *ui.UI, msgChan chan util.Message, msgHandlers *MessageHandlerList) {
    for msg := range msgChan {
        msgHandler := msgHandlers.List[msg.TypeName]

        // If spreads are nil, the list below are the only commands
        // that are allowed to run
        if m.Spreads == nil &&
            (msg.TypeName != "quit" &&
            msg.TypeName != "openFile" &&
            msg.TypeName != "openFileResult" &&
            msg.TypeName != "toggleFullscreen") {
            continue
        }

        // We have a handler, schedule it to run on event dispatch thread
        if msgHandler != nil {
            u.RunFunc(func() {
                msgHandler(msg.Data)
            })
        }

        // Render after every command, except refreshSpreads
        if msg.TypeName != "refreshSpreads" {
            u.Render(m)
        }

        runtime.GC()

        // Handling the quit message above 
        // scheduled any open file to be closed and
        // the GUI to be shut down which will let the program
        // exit. In the meanwhile this routine is no longer
        // needed so we can break the loop allowing it to 
        // return
        if msg.TypeName == "quit" {
            break
        }
    }
}

// Setup the model
// Setup the ui
// Create messageHandlers
// Start the update message handler
// Open the main window, when it closes
// Shutdown UI threads
// Exit
func main() {
    msgChan := make(chan util.Message)
    md := model.ProgramMetadata{Name: NAME, Version: VERSION}
    messenger := func(m util.Message) { msgChan <- m }
    m := model.NewModel(md, messenger)
    u := ui.NewUI(m, messenger)
    msgHandlers := NewMessageHandlers(m, u)

    go update(m, u, msgChan, msgHandlers)

    u.RunFunc(func() {
        //default to 2-page display
        msgHandlers.List["setLayoutModeTwoPage"]("")
        if len(os.Args) > 1 {
            messenger(util.Message{TypeName: "openFile", Data: os.Args[1]})
        }
    })

    u.Run()
}

