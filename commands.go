package main

import (
	"math"
    "path/filepath"
    "time"
)

type Command struct {
    Name string
    DisplayName string
    BindKey []string
}

type CommandList struct {
    Commands map[string]func(data string)
}

func NewCommands(model *Model) *CommandList {
    cmds := &CommandList{Commands: make(map[string]func(data string)),}

    cmd := Command {
        Name: "nextPage",
        DisplayName: "Next Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if model.currentLeaf < len(model.leaves) - 1 {
            model.currentLeaf++
            model.selectedPage = calcVersoPage(model)
            RefreshPages(model)
            model.leaves = NewLeaves(model)
        } else {
            cmds.Commands["nextFile"]("")
        }
    }

    cmd = Command {
        Name: "previousPage",
        DisplayName: "Previous Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if model.currentLeaf > 0 {
            model.currentLeaf--
            model.selectedPage = calcVersoPage(model)
            RefreshPages(model)
            model.leaves = NewLeaves(model)
        } else {
            cmds.Commands["previousFile"]("")
        }
    }

    cmd = Command {
        Name: "firstPage",
        DisplayName: "First Page",
    }
    cmds.Commands [cmd.Name] = func(data string) {
        model.currentLeaf = 0
        model.selectedPage = calcVersoPage(model)
        RefreshPages(model)
        model.leaves = NewLeaves(model)
    }

    cmd = Command {
        Name: "lastPage",
        DisplayName: "Last Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        model.currentLeaf = (len(model.leaves) - 1)
        model.selectedPage = calcVersoPage(model)
        RefreshPages(model)
        model.leaves = NewLeaves(model)
    }

    cmd = Command {
        Name: "selectPage",
        DisplayName: "Select Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if model.leafMode == TWO_PAGE {
            if model.selectedPage == calcVersoPage(model) {
                model.selectedPage++
            } else {
                model.selectedPage = calcVersoPage(model)
            }
        }
    }

    cmd = Command {
        Name: "setDisplayModeOnePage",
        DisplayName: "Display Mode One Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        model.leafMode = ONE_PAGE
        model.leaves = NewLeaves(model)
        c := model.currentLeaf*2
        if c < 0 {
            c = 0
        } else if c > len(model.leaves) - 1 {
            c = len(model.leaves) - 1
        }
        model.currentLeaf = c
        model.selectedPage = calcVersoPage(model)
        RefreshPages(model)
        model.leaves = NewLeaves(model)
    }

    cmd = Command {
        Name: "setDisplayModeTwoPage",
        DisplayName: "Display Mode Two Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        model.leafMode = TWO_PAGE
        model.leaves = NewLeaves(model)
        c := int(math.Floor(float64(model.currentLeaf)/2))
        if c < 0 {
            c = 0
        } else if c > len(model.leaves) - 1 {
            c = len(model.leaves) - 1
        }
        model.currentLeaf = c
        model.selectedPage = calcVersoPage(model)
        RefreshPages(model)
        model.leaves = NewLeaves(model)
    }

    cmd = Command {
        Name: "setDisplayModeLongStrip",
        DisplayName: "Display Mode Long Strip",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        model.leafMode = LONG_STRIP
        model.leaves = NewLeaves(model)
        model.currentLeaf = 0
        model.selectedPage = calcVersoPage(model)
        RefreshPages(model)
        model.leaves = NewLeaves(model)
    }

    cmd = Command {
        Name: "toggleReadMode",
        DisplayName: "Toggle Read Mode",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        // Toggle the read mode
        if model.readMode == LTR {
            model.readMode = RTL
        } else {
            model.readMode = LTR
        }

        // Swap the keys
        n := cmds.Commands["nextPage"]
        p := cmds.Commands["previousPage"]
        cmds.Commands["nextPage"] = p
        cmds.Commands["previousPage"] = n
    }

    cmd = Command {
        Name: "toggleFullscreen",
        DisplayName: "Toggle Fullscreen",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        // Toggle the read mode
        if model.fullscreen == true {
            model.fullscreen = false
        } else {
            model.fullscreen = true
        }
    }

    cmd = Command {
        Name: "openFile",
        DisplayName: "Open File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        model.filePath = data
        model.browseDirectory = filepath.Dir(data)

        go loadHash(model)
        go loadCbxFile(model)
        go loadSeriesList(model)
        model.selectedPage = calcVersoPage(model)
    }

    cmd = Command {
        Name: "closeFile",
        DisplayName: "Close File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        closeCbxFile(model)
    }

    cmd = Command {
        Name: "nextFile",
        DisplayName: "Next File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if model.seriesIndex < (len(model.seriesList) - 1) {
            model.seriesIndex++
            filePath := model.seriesList[model.seriesIndex]
            cmds.Commands["closeFile"]("")
            cmds.Commands["openFile"](filePath)
        }
    }

    cmd = Command {
        Name: "previousFile",
        DisplayName: "Previous File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if model.seriesIndex > 0 {
            model.seriesIndex--
            filePath := model.seriesList[model.seriesIndex]
            cmds.Commands["closeFile"]("")
            cmds.Commands["openFile"](filePath)
        }
    }

    cmd = Command {
        Name: "exportFile",
        DisplayName: "Export File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        srcPath := model.pages[model.selectedPage].filePath
        dstPath := data
        model.browseDirectory = filepath.Dir(dstPath)
        ExportFile(srcPath, dstPath)
    }

    cmd = Command {
        Name: "toggleBookmark",
        DisplayName: "Toggle Bookmark",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        p := model.selectedPage
        b := model.bookmarks.Find(p)
        if b != nil {
            model.bookmarks.Remove(*b)
        } else {
            b = &Bookmark{PageIndex: p, CreationTime: time.Now().UnixMilli()}
            model.bookmarks.Add(*b)
        }
    }

    cmd = Command {
        Name: "reflow",
        DisplayName: "Reflow",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if model.leafMode == TWO_PAGE {
            p := &model.pages[model.selectedPage]
            if p.Orientation == PORTRAIT {
                p.Orientation = LANDSCAPE
            } else {
                p.Orientation = PORTRAIT
            }
            RefreshPages(model)
            model.leaves = NewLeaves(model)
        }
    }

    cmd = Command {
        Name: "render",
        DisplayName: "Render",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        //noop render always gets called after cmd
    }

    cmd = Command {
        Name: "quit",
        DisplayName: "Quit",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        cmds.Commands["closeFile"]("")
        quit()
    }

    return cmds
}

