package main

import (
    "path/filepath"
    "strconv"
    "time"

    "github.com/mftb0/cbxv-gotk3/internal/model"
    "github.com/mftb0/cbxv-gotk3/internal/util"
)

type Command struct {
    Name        string
    DisplayName string
    BindKey     []string
}

type CommandList struct {
    Commands map[string]func(data string)
}

func NewCommands(m *model.Model) *CommandList {
    cmds := &CommandList{Commands: make(map[string]func(data string))}

    cmd := Command{
        Name:        "rightPage",
        DisplayName: "Right Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.SpreadIndex < len(m.Spreads)-1 {
            m.SpreadIndex++
            m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
            go m.RefreshPages()
        } else {
            cmds.Commands["nextFile"]("")
        }
    }

    cmd = Command{
        Name:        "leftPage",
        DisplayName: "Left Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.SpreadIndex > 0 {
            m.SpreadIndex--
            m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
            go m.RefreshPages()
        } else {
            cmds.Commands["previousFile"]("")
        }
    }

    cmd = Command{
        Name:        "firstPage",
        DisplayName: "First Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.SpreadIndex = 0
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        m.RefreshPages()
    }

    cmd = Command{
        Name:        "lastPage",
        DisplayName: "Last Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.SpreadIndex = (len(m.Spreads) - 1)
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        m.RefreshPages()
    }

    cmd = Command{
        Name:        "lastBookmark",
        DisplayName: "Last Bookmark",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        blen := (len(m.Bookmarks.Model.Bookmarks) - 1)
        if blen > -1 {
            bkmk := m.Bookmarks.Model.Bookmarks[blen]
            if bkmk.PageIndex > 0 && bkmk.PageIndex < len(m.Pages) {
                m.SpreadIndex = m.PageToSpread(bkmk.PageIndex)
                m.PageIndex = bkmk.PageIndex
            }
            m.RefreshPages()
        }
    }

    cmd = Command{
        Name:        "selectPage",
        DisplayName: "Select Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.LayoutMode == model.TWO_PAGE {
            if m.PageIndex == m.Spreads[m.SpreadIndex].VersoPage() {
                m.PageIndex++
            } else {
                m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
            }
        }
    }

    cmd = Command{
        Name:        "setLayoutModeOnePage",
        DisplayName: "Layout Mode One Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LayoutMode = model.ONE_PAGE
        m.SpreadIndex = m.PageToSpread(m.PageIndex)
        m.NewSpreads()
        if m.SpreadIndex > len(m.Spreads)-1 {
            m.SpreadIndex = len(m.Spreads) - 1
        }
        m.RefreshPages()
    }

    cmd = Command{
        Name:        "setLayoutModeTwoPage",
        DisplayName: "Layout Mode Two Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LayoutMode = model.TWO_PAGE
        m.SpreadIndex = m.PageToSpread(m.PageIndex)
        m.NewSpreads()
        if m.SpreadIndex > len(m.Spreads)-1 {
            m.SpreadIndex = len(m.Spreads) - 1
        }
        m.RefreshPages()
    }

    cmd = Command{
        Name:        "setLayoutModeLongStrip",
        DisplayName: "Layout Mode Long Strip",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.LayoutMode = model.LONG_STRIP
        m.SpreadIndex = 0
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        m.NewSpreads()
        m.RefreshPages()
    }

    cmd = Command{
        Name:        "toggleDirection",
        DisplayName: "Toggle Read Mode",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        // Toggle the read mode
        if m.Direction == model.LTR {
            m.Direction = model.RTL
        } else {
            m.Direction = model.LTR
        }

        // Swap what these do, so they continue to do what they say 0_o
        r := cmds.Commands["rightPage"]
        l := cmds.Commands["leftPage"]
        cmds.Commands["rightPage"] = l
        cmds.Commands["leftPage"] = r
    }

    cmd = Command{
        Name:        "toggleFullscreen",
        DisplayName: "Toggle Fullscreen",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.Fullscreen == true {
            m.Fullscreen = false
        } else {
            m.Fullscreen = true
        }
    }

    cmd = Command{
        Name:        "openFile",
        DisplayName: "Open File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        cmds.Commands["closeFile"]("")
        m.FilePath = data
        m.BrowseDir = filepath.Dir(data)

        // Start loading stuff
        // See the model for details about
        // Error handling
        m.Loading = true
        m.LoadHash()
        go m.LoadCbxFile()
        go m.LoadSeriesList()

        m.PageIndex = 0
    }

    cmd = Command{
        Name:        "closeFile",
        DisplayName: "Close File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.CloseCbxFile()
    }

    cmd = Command{
        Name:        "nextFile",
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

    cmd = Command{
        Name:        "previousFile",
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

    cmd = Command{
        Name:        "exportFile",
        DisplayName: "Export File",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        srcPath := m.Pages[m.PageIndex].FilePath
        dstPath := data
        m.ExportDir = filepath.Dir(dstPath)
        util.ExportFile(srcPath, dstPath)
    }

    cmd = Command{
        Name:        "toggleBookmark",
        DisplayName: "Toggle Bookmark",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        p := m.PageIndex
        b := m.Bookmarks.Find(p)
        if b != nil {
            m.Bookmarks.Remove(*b)
        } else {
            b = &model.Bookmark{PageIndex: p, CreationTime: time.Now().UnixMilli()}
            m.Bookmarks.Add(*b)
        }
    }

    cmd = Command{
        Name:        "toggleJoin",
        DisplayName: "toggleJoin",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        if m.LayoutMode == model.TWO_PAGE {
            pi := m.PageIndex
            p := &m.Pages[pi]
            if p.Span == model.SINGLE {
                p.Span = model.DOUBLE
            } else {
                p.Span = model.SINGLE
            }
            m.RefreshPages()
            m.NewSpreads()
            m.StoreLayout()
            m.SpreadIndex = m.PageToSpread(pi)
            m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        }
    }

    cmd = Command{
        Name:        "hidePage",
        DisplayName: "Hide Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        pi := m.PageIndex
        p := &m.Pages[pi]
        p.Hidden = true
        m.RefreshPages()
        m.NewSpreads()
        m.StoreLayout()
        m.SpreadIndex = m.PageToSpread(pi)
    }

    cmd = Command{
        Name:        "showPage",
        DisplayName: "Show Page",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        i, err := strconv.Atoi(data)
        if err != nil {
            return
        }

        if i < 0 || i > len(m.Pages)-1 {
            return
        }

        pi := m.PageIndex

        if i < pi {
            pi++
        }

        p := &m.Pages[i]
        p.Hidden = false
        m.RefreshPages()
        m.NewSpreads()
        m.StoreLayout()
        m.SpreadIndex = m.PageToSpread(pi)
    }

    cmd = Command{
        Name:        "loadAllPages",
        DisplayName: "Load All Pages",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        m.RefreshPages()
        m.NewSpreads()
    }

    cmd = Command{
        Name:        "render",
        DisplayName: "Render",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        //noop render always gets called after cmd
    }

    cmd = Command{
        Name:        "quit",
        DisplayName: "Quit",
    }
    cmds.Commands[cmd.Name] = func(data string) {
        // because of orchestration with gtk's
        // thread this no longer works at shutdown
        // Mostly doesn't matter, but we do need
        // to clean up the last tmpDir, moved to
        // end of main
        cmds.Commands["closeFile"]("")
    }

    return cmds
}
