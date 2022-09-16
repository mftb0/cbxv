package main

import (
	"math"
    "path/filepath"
    "time"

    "example.com/cbxv-gotk3/internal/util"
    "example.com/cbxv-gotk3/internal/model"
)

type Command struct {
    Name string
    DisplayName string
    BindKey []string
}

type CommandList struct {
    Commands map[string]func(data string)
}

func NewCommands(m *model.Model) *CommandList {
    cmds := &CommandList{Commands: make(map[string]func(data string)),}

    cmd := Command {
        Name: "nextPage",
        DisplayName: "Next Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.CurrentLeaf < len(m.Leaves) - 1 {
            m.CurrentLeaf++
            m.SelectedPage = m.CalcVersoPage()
            m.RefreshPages()
            m.NewLeaves()
        } else {
            cmds.Commands["nextFile"]("")
        }
    }

    cmd = Command {
        Name: "previousPage",
        DisplayName: "Previous Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.CurrentLeaf > 0 {
            m.CurrentLeaf--
            m.SelectedPage = m.CalcVersoPage()
            m.RefreshPages()
            m.NewLeaves()
        } else {
            cmds.Commands["previousFile"]("")
        }
    }

    cmd = Command {
        Name: "firstPage",
        DisplayName: "First Page",
    }
    cmds.Commands [cmd.Name] = func(data string) {
        m.CurrentLeaf = 0
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
        m.NewLeaves()
    }

    cmd = Command {
        Name: "lastPage",
        DisplayName: "Last Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.CurrentLeaf = (len(m.Leaves) - 1)
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
        m.NewLeaves()
    }

    cmd = Command {
        Name: "selectPage",
        DisplayName: "Select Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.LeafMode == model.TWO_PAGE {
            if m.SelectedPage == m.CalcVersoPage() {
                m.SelectedPage++
            } else {
                m.SelectedPage = m.CalcVersoPage()
            }
        }
    }

    cmd = Command {
        Name: "setDisplayModeOnePage",
        DisplayName: "Display Mode One Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LeafMode = model.ONE_PAGE
        m.NewLeaves()
        c := m.CurrentLeaf*2
        if c < 0 {
            c = 0
        } else if c > len(m.Leaves) - 1 {
            c = len(m.Leaves) - 1
        }
        m.CurrentLeaf = c
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
        m.NewLeaves()
    }

    cmd = Command {
        Name: "setDisplayModeTwoPage",
        DisplayName: "Display Mode Two Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LeafMode = model.TWO_PAGE
        m.NewLeaves()
        c := int(math.Floor(float64(m.CurrentLeaf)/2))
        if c < 0 {
            c = 0
        } else if c > len(m.Leaves) - 1 {
            c = len(m.Leaves) - 1
        }
        m.CurrentLeaf = c
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
        m.NewLeaves()
    }

    cmd = Command {
        Name: "setDisplayModeLongStrip",
        DisplayName: "Display Mode Long Strip",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LeafMode = model.LONG_STRIP
        m.NewLeaves()
        m.CurrentLeaf = 0
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
        m.NewLeaves()
    }

    cmd = Command {
        Name: "toggleReadMode",
        DisplayName: "Toggle Read Mode",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        // Toggle the read mode
        if m.ReadMode == model.LTR {
            m.ReadMode = model.RTL
        } else {
            m.ReadMode = model.LTR
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
        if m.Fullscreen == true {
            m.Fullscreen = false
        } else {
            m.Fullscreen = true
        }
    }

    cmd = Command {
        Name: "openFile",
        DisplayName: "Open File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.FilePath = data
        m.BrowseDirectory = filepath.Dir(data)

        go m.LoadHash()
        go loadCbxFile(m, m.SendMessage)
        go loadSeriesList(m)

        m.SelectedPage = m.CalcVersoPage()
    }

    cmd = Command {
        Name: "closeFile",
        DisplayName: "Close File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        closeCbxFile(m)
    }

    cmd = Command {
        Name: "nextFile",
        DisplayName: "Next File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.SeriesIndex < (len(m.SeriesList) - 1) {
            m.SeriesIndex++
            filePath := m.SeriesList[m.SeriesIndex]
            cmds.Commands["closeFile"]("")
            cmds.Commands["openFile"](filePath)
        }
    }

    cmd = Command {
        Name: "previousFile",
        DisplayName: "Previous File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.SeriesIndex > 0 {
            m.SeriesIndex--
            filePath := m.SeriesList[m.SeriesIndex]
            cmds.Commands["closeFile"]("")
            cmds.Commands["openFile"](filePath)
        }
    }

    cmd = Command {
        Name: "exportFile",
        DisplayName: "Export File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        srcPath := m.Pages[m.SelectedPage].FilePath
        dstPath := data
        m.BrowseDirectory = filepath.Dir(dstPath)
        util.ExportFile(srcPath, dstPath)
    }

    cmd = Command {
        Name: "toggleBookmark",
        DisplayName: "Toggle Bookmark",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        p := m.SelectedPage
        b := m.Bookmarks.Find(p)
        if b != nil {
            m.Bookmarks.Remove(*b)
        } else {
            b = &model.Bookmark{PageIndex: p, CreationTime: time.Now().UnixMilli()}
            m.Bookmarks.Add(*b)
        }
    }

    cmd = Command {
        Name: "reflow",
        DisplayName: "Reflow",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.LeafMode == model.TWO_PAGE {
            p := &m.Pages[m.SelectedPage]
            if p.Orientation == model.PORTRAIT {
                p.Orientation = model.LANDSCAPE
            } else {
                p.Orientation = model.PORTRAIT
            }
            m.RefreshPages()
            m.NewLeaves()
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

