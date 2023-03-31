package main

import (
    "encoding/json"
    "fmt"
    "path/filepath"
    "strconv"
    "time"

    "github.com/mftb0/cbxv/internal/model"
    "github.com/mftb0/cbxv/internal/ui"
    "github.com/mftb0/cbxv/internal/util"
)

/* 
 * Messages are just the generic way to communicate with the app
 * They can be looked up by name and take a single argument
 * which can contain whatever data as a string, structured or 
 * unstructured. 
 *
 * MessageHandlers can handle messages from both the model and the UI
 */
 
type MessageHandlerList struct {
    List map[string]func(data string)
}

func NewMessageHandlers(m *model.Model, u *ui.UI) *MessageHandlerList {
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

    // The first part of opening a cbx file is handled
    // async, because it's the single most expensive 
    // thing in the program. When the first half completes
    // the model will emit an "openFileResult", see below
    handlers.List["openFile"] = func(data string) {
        handlers.List["closeFile"]("")
        m.FilePath = data
        m.BrowseDir = filepath.Dir(data)

        // Start loading
        m.Loading = true
        go m.OpenCbxFile()
        go m.LoadSeriesList()
    }

    // Opening a cbx, Success or failure resolves here
    // By the end of this handler the loading of the
    // cbx will also be finished success or fail 
    // It's possible that loading the SeriesList may still
    // be outstanding or even fail, it's non-critical
    handlers.List["openFileResult"] = func(data string) {
		var r model.Result
		err := json.Unmarshal([]byte(data), &r)
		if err != nil {
            msg := fmt.Sprintf("Error unable to decode openFileResult: %s", err)
            u.DisplayErrorDlg(msg)
		} else {
            if r.Code == model.OK {
                m.LoadCbxFile()
                if len(m.Spreads) > 0 {
                    m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
                } else {
                    // If there are no spreads, bad file?
                    m.PageIndex = 0
                }
            } else {
                msg := fmt.Sprintf("%s, %d", r.Description, r.Code)
                u.DisplayErrorDlg(msg)
            }
        }
        // End loading
	    m.Loading = false
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

    // Hide the currently selected page
    handlers.List["hidePage"] = func(data string) {
        // Note current SpreadIndex
        si := m.PageToSpread(m.PageIndex)

        // Hide page
        p := &m.Pages[m.PageIndex]
        p.Hidden = true

        // Recalculate layout
        m.RefreshPages()
        m.NewSpreads()
        m.StoreLayout()

        // It's possible we lost a spread so
        // make sure we cap the SpreadIndex
        if si >= len(m.Spreads) {
            si = len(m.Spreads) - 1
        }

        // Restore the SpreadIndex
        m.SpreadIndex = si
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
    }

    // Unhide a page (doesn't even have to be on screen, let alone selected)
    handlers.List["showPage"] = func(data string) {
        i, err := strconv.Atoi(data)
        if err != nil {
            return
        }

        if i < 0 || i > len(m.Pages)-1 {
            return
        }

        // Note current PageIndex
        pi := m.PageIndex

        // Recalculate layout
        p := &m.Pages[i]
        p.Hidden = false
        m.RefreshPages()
        m.NewSpreads()
        m.StoreLayout()

        // Set the spread index based on the noted page
        m.SpreadIndex = m.PageToSpread(pi)
        m.PageIndex = m.Spreads[m.SpreadIndex].VersoPage()
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

