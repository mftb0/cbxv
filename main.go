package main

import (
	"embed"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"time"

	_ "github.com/gotk3/gotk3/cairo"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

// Simple cbx application with a gui provided by gtk

//go:embed assets
var assets embed.FS

func loadTextFile(filePath string) (*string, error) {
    b, err := assets.ReadFile(filePath)
    if(err != nil) {
        return nil, err
    }

    s := string(b)
    return &s, nil
}

// Utility for pages to load images
func loadImageFile(filePath string) (image.Image, error) {
    f, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer f.Close()

    img, _, err := image.Decode(f)
    if err != nil {
        return nil, err
    }
    return img, nil
}

// Some data loading utils
func loadCbxFile(model *Model) error {

    td, err := CreateTmpDir()
    if err != nil {
        return err
    }
    model.tmpDir = td

    ip, err := GetImagePaths(model.filePath, model.tmpDir)
    if err != nil {
        return err
    }
    model.imgPaths = ip
    model.pages = NewPages(model)
    model.leaves = NewLeaves(model)
    model.currentLeaf = 0
    model.selectedPage = 0
    RefreshPages(model)

    m := &Message{typeName: "render"}
    sendMessage(*m)

    return nil
}

func closeCbxFile(model *Model) {

    os.RemoveAll(model.tmpDir)
    model.imgPaths = nil
    model.pages = nil
    model.leaves = nil
    model.currentLeaf = 0
    model.selectedPage = 0
    model.bookmarks = nil
    model.seriesList = nil
    model.seriesIndex = 0
}

func loadSeriesList(model *Model) {
    s, err := ReadSeriesList(model.filePath)
    if err != nil {
        fmt.Printf("Unable to load series list %s\n", err)
    }
    model.seriesList = s

    for i := range s {
        if model.filePath == s[i] {
            model.seriesIndex = i
        }
    }
    model.selectedPage = calcVersoPage(model)
}

func loadHash(model *Model) {
    hash, err := HashFile(model.filePath)
    if err != nil {
        fmt.Printf("Unable to compute file hash %s\n", err)
    }
    model.hash = hash
}

func loadBookmarks(model *Model) {
    model.bookmarks = NewBookmarkList(model.filePath)
    model.bookmarks.Load(model.hash)
}

func quit() {
    gtk.MainQuit()
}

// Stuff to handle messages from the ui
type Message struct {
    typeName string
    data string
}

func sendMessage(m Message) {
    msgChan <- m
}

var msgChan = make(chan Message)

// Update listens for message on the message channel and
// handles messages by invoking commands which update the model
func update(model *Model, ui *UI, commands *CommandList) {
    for m := range msgChan {
        cmd := commands.Commands[m.typeName]
        if model.leaves == nil && (
            m.typeName != "quit" &&
            m.typeName != "openFile") {
            continue
        }
        if cmd != nil {
            glib.IdleAdd(func(){
                cmd(m.data)
                render(model, ui)
            })
        }
    }
}

func hudHandler(ui *UI) {
    for {
        select {
        case <-hudTicker.C:
            if !ui.hudHidden {
                glib.IdleAdd(func(){
                    ui.hdrControl.container.Hide()
                    ui.navControl.container.Hide()
                    ui.mainWindow.QueueDraw()
                })
                ui.hudHidden = true;
            }
        case r := <-hudChan:
            if r == true {
                hudTicker = time.NewTicker(time.Second * 10)
                ui.hudHidden = false
            } else {
                return
            }
        }
    }
}

// Ticker to hide the HUD
var hudTicker *time.Ticker
var hudChan chan bool

// Setup the model
// Setup the ui
// Create commands to modify the model
// Start the update message handler
// Open the main window, when it closes program exits
func main() {
    var model = NewModel()
    var ui = &UI{}
    commands := NewCommands(model)

    InitUI(model, ui)

    hudChan = make(chan bool)
    hudTicker = time.NewTicker(time.Second * 15)
    defer hudTicker.Stop()

    go hudHandler(ui)
    go update(model, ui, commands)

    ui.mainWindow.ShowAll()

    if len(os.Args) > 1 {
        commands.Commands["openFile"](os.Args[1])
    }

    gtk.Main()
    hudChan <-false
}

