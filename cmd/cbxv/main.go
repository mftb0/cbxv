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
    VERSION = "0.4.7"
)

// Update listens for messages on the message channel and
// handles messages by invoking messageHandlers
func update(m *model.Model, u *ui.UI, msgChan chan util.Message, msgHandlers *MessageHandlerList) {
    for msg := range msgChan {
        msgHandler := msgHandlers.List[msg.TypeName]
        if m.Spreads == nil &&
            (msg.TypeName != "quit" &&
            msg.TypeName != "openFile" &&
            msg.TypeName != "openFileResult" &&
            msg.TypeName != "toggleFullscreen") {
            continue
        }
        if msgHandler != nil {
            u.RunFunc(func() {
                msgHandler(msg.Data)
                u.Render(m)
            })
        }
        runtime.GC()
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

    // At the end of all things the closeFile message
    // can't work because we've orchestrated for it only
    // to be run on the UI thread and the UI thread is dead so
    // we have to close the last cbx file here, to get
    // rid of any tmpDir
    m.CloseCbxFile()
}

