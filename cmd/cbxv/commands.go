package main

import (
	"path/filepath"
	"time"

	"example.com/cbxv-gotk3/internal/model"
	"example.com/cbxv-gotk3/internal/util"
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
        if m.CurrentSpread < len(m.Spreads) - 1 {
            m.CurrentSpread++
            m.SelectedPage = m.CalcVersoPage()
            go m.RefreshPages()
        } else {
            cmds.Commands["nextFile"]("")
        }
    }

    cmd = Command {
        Name: "previousPage",
        DisplayName: "Previous Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.CurrentSpread > 0 {
            m.CurrentSpread--
            m.SelectedPage = m.CalcVersoPage()
            go m.RefreshPages()
        } else {
            cmds.Commands["previousFile"]("")
        }
    }

    cmd = Command {
        Name: "firstPage",
        DisplayName: "First Page",
    }
    cmds.Commands [cmd.Name] = func(data string) {
        m.CurrentSpread = 0
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
    }

    cmd = Command {
        Name: "lastPage",
        DisplayName: "Last Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.CurrentSpread = (len(m.Spreads) - 1)
        m.SelectedPage = m.CalcVersoPage()
        m.RefreshPages()
    }

    cmd = Command {
        Name: "lastBookmark",
        DisplayName: "Last Bookmark",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        blen := (len(m.Bookmarks.Model.Bookmarks) - 1)
        if blen > -1 {
            bkmk := m.Bookmarks.Model.Bookmarks[blen]
            if bkmk.PageIndex > 0 && bkmk.PageIndex < len(m.Pages) {
                m.CurrentSpread = m.PageToSpread(bkmk.PageIndex)
                m.SelectedPage = bkmk.PageIndex 
            }
            m.RefreshPages()
        }
    }

    cmd = Command {
        Name: "selectPage",
        DisplayName: "Select Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.LayoutMode == model.TWO_PAGE {
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
        m.LayoutMode = model.ONE_PAGE
        m.CurrentSpread = m.PageToSpread(m.SelectedPage)
        m.NewSpreads()
        if m.CurrentSpread > len(m.Spreads) - 1 {
            m.CurrentSpread = len(m.Spreads) - 1
        }
        m.RefreshPages()
    }

    cmd = Command {
        Name: "setDisplayModeTwoPage",
        DisplayName: "Display Mode Two Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LayoutMode = model.TWO_PAGE
        m.CurrentSpread = m.PageToSpread(m.SelectedPage)
        m.NewSpreads()
        if m.CurrentSpread > len(m.Spreads) - 1 {
            m.CurrentSpread = len(m.Spreads) - 1
        }
        m.RefreshPages()
    }

    cmd = Command {
        Name: "setDisplayModeLongStrip",
        DisplayName: "Display Mode Long Strip",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LayoutMode = model.LONG_STRIP
        m.CurrentSpread = 0
        m.SelectedPage = m.CalcVersoPage()
        m.NewSpreads()
        m.RefreshPages()
    }

    cmd = Command {
        Name: "toggleDirection",
        DisplayName: "Toggle Read Mode",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        // Toggle the read mode
        if m.Direction == model.LTR {
            m.Direction = model.RTL
        } else {
            m.Direction = model.LTR
        }

        // Swap the keys
        // fixme: This means in rtl things are
        // named backward
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

        // Start loading stuff
        // See the model for details about
        // Error handling
        m.Loading = true
        m.LoadHash()
        go m.LoadCbxFile()
        go m.LoadSeriesList()

        m.SelectedPage = m.CalcVersoPage()
    }

    cmd = Command {
        Name: "closeFile",
        DisplayName: "Close File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.CloseCbxFile()
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
        Name: "toggleSpread",
        DisplayName: "toggleSpread",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.LayoutMode == model.TWO_PAGE {
            spg := m.SelectedPage
            p := &m.Pages[spg]
            if p.Orientation == model.PORTRAIT {
                p.Orientation = model.LANDSCAPE
            } else {
                p.Orientation = model.PORTRAIT
            }
            m.RefreshPages()
            m.NewSpreads()
            m.StoreLayout()
            m.CurrentSpread = m.PageToSpread(spg)
            vpg := m.CalcVersoPage() 
            if m.SelectedPage == vpg + 1 {
                m.SelectedPage = vpg + 1
            } else {
                m.SelectedPage = m.CalcVersoPage()
            }
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
    }

    return cmds
}

