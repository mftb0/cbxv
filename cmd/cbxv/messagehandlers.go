package main

import (
    "path/filepath"
    "strconv"
    "time"

    "github.com/mftb0/cbxv-gotk3/internal/model"
    "github.com/mftb0/cbxv-gotk3/internal/util"
)

/* 
Messages are just the generic way to communicate with the app
They can be looked up by name and take a single argument
which can contain whatever data as a string, structured or 
unstructured. 

Currently the MessageHandlers only have access to the model, because that's 
all thats been needed. Generally the UI communicates to the app and the app 
takes action by updating the model.

There are a couple minor exceptions; quit, where the app terminates
itself and render which the model sends to the app, but right now
since the app always calls render after every message its 
essentially a noop. This may change in the future though and I'll
add access to the UI
 */
 
type MessageHandlerList struct {
    List map[string]func(data string)
}

func NewMessageHandlers(m *model.Model) *MessageHandlerList {
    handlers := &MessageHandlerList{List: make(map[string]func(data string))}

    handlers.List["rightPage"] = func(data string) {
        if m.SpreadIndex < len(m.Spreads)-1 {
            m.SpreadIndex++
            m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
            m.RefreshPages()
        } else {
            handlers.List["nextFile"]("")
        }
    }

    handlers.List["leftPage"] = func(data string) {
        if m.SpreadIndex > 0 {
            m.SpreadIndex--
            m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
            m.RefreshPages()
        } else {
            handlers.List["previousFile"]("")
        }
    }

    handlers.List["firstPage"] = func(data string) {
        m.SpreadIndex = 0
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        m.RefreshPages()
    }

    handlers.List["lastPage"] = func(data string) {
        m.SpreadIndex = (len(m.Spreads) - 1)
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        m.RefreshPages()
    }

    handlers.List["lastBookmark"] = func(data string) {
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

    handlers.List["selectPage"] = func(data string) {
        if m.LayoutMode == model.TWO_PAGE {
            s := m.Spreads[m.SpreadIndex]
            if len(s.Pages) > 1 {
                if m.PageIndex == s.VersoPage() {
                    m.PageIndex = s.RectoPage()
                } else {
                    m.PageIndex = s.VersoPage()
                }
            }
        }
    }

    handlers.List["setLayoutModeOnePage"] = func(data string) {
        m.LayoutMode = model.ONE_PAGE
        m.SpreadIndex = m.PageToSpread(m.PageIndex)
        m.NewSpreads()
        if m.SpreadIndex > len(m.Spreads)-1 {
            m.SpreadIndex = len(m.Spreads) - 1
        }
        m.RefreshPages()
    }

    handlers.List["setLayoutModeTwoPage"] = func(data string) {
        m.LayoutMode = model.TWO_PAGE
        m.SpreadIndex = m.PageToSpread(m.PageIndex)
        m.NewSpreads()
        if m.SpreadIndex > len(m.Spreads)-1 {
            m.SpreadIndex = len(m.Spreads) - 1
        }
        m.RefreshPages()
    }

    handlers.List["setLayoutModeLongStrip"] = func(data string) {
        m.LayoutMode = model.LONG_STRIP
        m.SpreadIndex = 0
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
        m.NewSpreads()
        m.RefreshPages()
    }

    handlers.List["toggleDirection"] = func(data string) {
        // Toggle the read mode
        if m.Direction == model.LTR {
            m.Direction = model.RTL
        } else {
            m.Direction = model.LTR
        }

        // Swap what these do, so they continue to do what they say 0_o
        r := handlers.List["rightPage"]
        l := handlers.List["leftPage"]
        handlers.List["rightPage"] = l
        handlers.List["leftPage"] = r
    }

    handlers.List["toggleFullscreen"] = func(data string) {
        if m.Fullscreen == true {
            m.Fullscreen = false
        } else {
            m.Fullscreen = true
        }
    }

    handlers.List["openFile"] = func(data string) {
        handlers.List["closeFile"]("")
        m.FilePath = data
        m.BrowseDir = filepath.Dir(data)

        // Start loading stuff
        // See the model for details about
        // Error handling
        m.Loading = true
        m.LoadHash()
        m.LoadCbxFile()
        go m.LoadSeriesList()

        m.PageIndex = 0
    }

    handlers.List["closeFile"] = func(data string) {
        m.CloseCbxFile()
    }

    handlers.List["nextFile"] = func(data string) {
        if m.SeriesIndex < (len(m.SeriesList) - 1) {
            m.SeriesIndex++
            filePath := m.SeriesList[m.SeriesIndex]
            handlers.List["closeFile"]("")
            handlers.List["openFile"](filePath)
        }
    }

    handlers.List["previousFile"] = func(data string) {
        if m.SeriesIndex > 0 {
            m.SeriesIndex--
            filePath := m.SeriesList[m.SeriesIndex]
            handlers.List["closeFile"]("")
            handlers.List["openFile"](filePath)
        }
    }

    handlers.List["exportPage"] = func(data string) {
        srcPath := m.Pages[m.PageIndex].FilePath
        dstPath := data
        m.ExportDir = filepath.Dir(dstPath)
        util.ExportPage(srcPath, dstPath)
    }

    handlers.List["toggleBookmark"] = func(data string) {
        p := m.PageIndex
        b := m.Bookmarks.Find(p)
        if b != nil {
            m.Bookmarks.Remove(*b)
        } else {
            b = &model.Bookmark{PageIndex: p, CreationTime: time.Now().UnixMilli()}
            m.Bookmarks.Add(*b)
        }
    }

    handlers.List["toggleJoin"] = func(data string) {
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

    handlers.List["hidePage"] = func(data string) {
        pi := m.PageIndex
        p := &m.Pages[pi]
        p.Hidden = true
        m.RefreshPages()
        m.NewSpreads()
        m.StoreLayout()
        m.SpreadIndex = m.PageToSpread(pi)
    }

    handlers.List["showPage"] = func(data string) {
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

    handlers.List["loadAllPages"] = func(data string) {
        m.RefreshPages()
        m.NewSpreads()
    }

    handlers.List["render"] = func(data string) {
        //noop render always gets called after cmd
    }

    handlers.List["quit"] = func(data string) {
        // because of orchestration with gtk's
        // thread this no longer works at shutdown
        // Mostly doesn't matter, but we do need
        // to clean up the last tmpDir, moved to
        // end of main
        handlers.List["closeFile"]("")
    }

    return handlers
}

