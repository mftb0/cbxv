package model

import (
    "encoding/json"
    "fmt"
    "math"
    "os"
    "runtime/debug"
    "sort"

    "github.com/mftb0/cbxv/internal/util"
)

// Data model of a cbx application
// Composed of a handful of sub-models, collections and other standard types
type Model struct {
    SendMessage    util.Messenger
    FilePath       string
    TmpDir         string
    Hash           string
    Bookmarks      *BookmarkList
    ImgPaths       []string
    Pages          []Page
    PageIndex      int
    Spreads        []*Spread
    SpreadIndex    int
    Direction      Direction
    LayoutMode     LayoutMode
    SeriesList     []string
    SeriesIndex    int
    BrowseDir      string
    ExportDir      string
    HiddenPages    bool
    Fullscreen     bool
    Loading        bool
    ProgramName    string
    ProgramVersion string
}

func NewModel(md ProgramMetadata, messenger util.Messenger) *Model {
    m := &Model{}
    m.ProgramName = md.Name
    m.ProgramVersion = md.Version
    m.SendMessage = messenger
    m.BrowseDir, _ = os.Getwd()
    return m
}

type ProgramMetadata struct {
    Name    string
    Version string
}

// Direction of the model is either
// Left-To-Right or
// Right-To-Left
type Direction int

const (
    LTR Direction = iota
    RTL
)

// Maximum to load into the model
// Currently confusing because at init the units are pages
// Later it's spreads (so frequently double the number of pages
const (
    MAX_LOAD = 8
)

type ResultCode int

const (
    OK ResultCode = iota
    ERR
)

type Result struct {
    Code        ResultCode `json:"code"`
    Description string     `json:"description"`
}

// Mark a place in the model by keeping track of an index in the pages slice
type Bookmark struct {
    PageIndex    int   `json:"pageIndex"`
    CreationTime int64 `json:"creationTime"`
}

// Just a little type for serialization
type ComicData struct {
    Hash     string `json:"hash"`
    FilePath string `json:"filePath"`
}

// Type used for serialization/deserialization of the BookmarkList
type BookmarkListModel struct {
    FormatVersion string     `json:"formatVersion"`
    Comic         ComicData  `json:"comic"`
    Bookmarks     []Bookmark `json:"bookmarks"`
}

// Manage a list of bookmarks
type BookmarkList struct {
    Model BookmarkListModel
}

func NewBookmarkList(filePath string) *BookmarkList {
    b := BookmarkList{}
    m := BookmarkListModel{
        FormatVersion: "0.1",
    }
    c := ComicData{}
    c.Hash = ""
    c.FilePath = filePath
    m.Comic = c
    m.Bookmarks = make([]Bookmark, 0)
    b.Model = m
    return &b
}

func (l *BookmarkList) Add(b Bookmark) {
    for i := 0; i < len(l.Model.Bookmarks); i++ {
        if l.Model.Bookmarks[i].PageIndex == b.PageIndex {
            return
        }
    }
    l.Model.Bookmarks = append(l.Model.Bookmarks, b)
    sort.Slice(l.Model.Bookmarks, func(i, j int) bool {
        return l.Model.Bookmarks[i].PageIndex < l.Model.Bookmarks[j].PageIndex
    })
    l.Store()
}

func (l *BookmarkList) Remove(b Bookmark) *Bookmark {
    var x int
    var r Bookmark
    for i := 0; i < len(l.Model.Bookmarks); i++ {
        if l.Model.Bookmarks[i].PageIndex == b.PageIndex {
            r = l.Model.Bookmarks[i]
            x = i
            break
        }
    }
    l.Model.Bookmarks = append(l.Model.Bookmarks[:x], l.Model.Bookmarks[x+1:]...)
    l.Store()
    return &r
}

func (l *BookmarkList) Find(pageIndex int) *Bookmark {
    var r *Bookmark
    for i := 0; i < len(l.Model.Bookmarks); i++ {
        if l.Model.Bookmarks[i].PageIndex == pageIndex {
            r = &l.Model.Bookmarks[i]
            break
        }
    }
    return r
}

func (l *BookmarkList) Store() error {
    data, err := json.Marshal(l.Model)
    if err != nil {
        return err
    }
    err = util.WriteBookmarkList(l.Model.Comic.Hash, string(data))
    if err != nil {
        return err
    }
    return nil
}

func (l *BookmarkList) Load(hash string) {
    l.Model.Comic.Hash = hash
    data, _ := util.ReadBookmarkList(l.Model.Comic.Hash)

    if data != nil {
        var m BookmarkListModel
        err := json.Unmarshal([]byte(*data), &m)
        if err != nil {
            fmt.Printf("e:%s\n", err)
        }
        l.Model = m
    }
}


// A page in this case is generally analogous to an image
// They are grouped on Spreads
type Page struct {
    FilePath string      `json:"filePath"`
    Width    int         `json:"width"`
    Height   int         `json:"height"`
    Span     int         `json:"span"`
    Hidden   bool        `json:"hidden"`
    Loaded   bool        `json:"loaded"`
    Image    *util.Img   `json:"-"`
}

func (p *Page) Load() {
    // Must be called from ui event dispatch thread or
    // it will leak. 
    f, err := util.ImgNewFromFile(p.FilePath)
    if err != nil {
        fmt.Printf("Warning unable to load file %s\n", err)
        return
    }
    p.Image = f
    p.Width = f.GetWidth()
    p.Height = f.GetHeight()
    p.Loaded = true
}

func (p *Page) LoadMeta() {
    _, w, h, err := util.ImgGetFileInfo(p.FilePath)
    if err != nil {
        fmt.Printf("Warning unable to load metadata for file %s\n", err)
        return
    }
    p.Width = w
    p.Height = h
    p.Loaded = false
}

// Creates pgs slice and loads it
func (m *Model) NewPages() {

    pages := make([]Page, len(m.ImgPaths))

    for i := range m.ImgPaths {
        pages[i].FilePath = m.ImgPaths[i]
        pages[i].Span = SINGLE
        pages[i].Loaded = false
        if i < MAX_LOAD {
            pages[i].Load()
        } else {
            pages[i].LoadMeta()
        }
    }

    m.Pages = pages
}

// How a page is oriented
type Span int

const (
    SINGLE = iota
    DOUBLE
)

// A Spread is an element of a layout
// It's essentially the pages you can
// see at a given time
type Spread struct {
    Pages    []*Page
    PageIdxs []int
}

// Creates spread slice based on pg slice and layout mode
func (m *Model) NewSpreads() {
    var spreads []*Spread

    pages := m.Pages
    if m.LayoutMode == ONE_PAGE {
        for i := range pages {
            spread := &Spread{}
            p := &pages[i]
            if p.Hidden {
                m.HiddenPages = true
                continue
            }
            spread.Pages = append(spread.Pages, p)
            spread.PageIdxs = append(spread.PageIdxs, i)
            spreads = append(spreads, spread)
        }
    } else if m.LayoutMode == TWO_PAGE {
        for i := 0; i < len(pages); i++ {
            // create spread add a page
            spread := &Spread{}
            p := &pages[i]
            if p.Hidden {
                m.HiddenPages = true
                continue
            }
            spread.Pages = append(spread.Pages, p)
            spread.PageIdxs = append(spread.PageIdxs, i)

            // if pg is landscape, spread done
            if p.Span == DOUBLE {
                spreads = append(spreads, spread)
                continue
            }
            // if pg is last page, spreads done
            if i == (len(pages) - 1) {
                spreads = append(spreads, spread)
                break
            }

            // on to the next page
            i++

            // skip hidden pages
            for ; i < len(pages); i++ {
                p = &pages[i]
                if p.Hidden {
                    m.HiddenPages = true
                    continue
                } else {
                    break
                }
            }

            // If all pages to the end were hidden, add spread, we're done
            if i >= len(pages) {
                spreads = append(spreads, spread)
                break
            }

            // if pg is landscape, make a new spread, spread done
            if p.Span == DOUBLE {
                spreads = append(spreads, spread)
                spread = &Spread{}
                spread.Pages = append(spread.Pages, p)
                spread.PageIdxs = append(spread.PageIdxs, i)
                spreads = append(spreads, spread)
                continue
            }

            // no more special cases, add recto page and add last spread
            spread.Pages = append(spread.Pages, p)
            spread.PageIdxs = append(spread.PageIdxs, i)
            spreads = append(spreads, spread)
        }
    } else {
        // Put all pages on one spread
        spread := &Spread{}
        for i := range pages {
            p := &pages[i]
            if !p.Loaded {
                p.Load()
            }

            if p.Hidden {
                m.HiddenPages = true
                continue
            }

            spread.Pages = append(spread.Pages, p)
            spread.PageIdxs = append(spread.PageIdxs, i)
        }
        spreads = append(spreads, spread)
    }

    m.Spreads = spreads
}

func (s *Spread) VersoPage() int {
    return s.PageIdxs[0]
}

func (s *Spread) RectoPage() int {
    return s.PageIdxs[1]
}

// Layout mode determines the max pages per spread
// ONE_PAGE = 1 pg
// TWO_PAGE = up to 2 pgs
// LONG_STRIP = n pgs
type LayoutMode int

const (
    ONE_PAGE = iota
    TWO_PAGE
    LONG_STRIP
)

type Layout struct {
    FormatVersion string     `json:"formatVersion"`
    Comic         ComicData  `json:"comic"`
    Mode          LayoutMode `json:"mode"`
    Pages         []Page     `json:"pages"`
}

func (m *Model) LoadSeriesList() {
    s, err := util.ReadSeriesList(m.FilePath)
    if err != nil {
        fmt.Printf("Warning unable to load series list %s\n", err)
        return
    }
    m.SeriesList = s

    for i := range s {
        if m.FilePath == s[i] {
            m.SeriesIndex = i
        }
    }
}

/*
 * When the user fires the "openFile" event a process called "loading" starts.
 * There are two phases:
 *
 * The first phase "Opening" the cbx is asynchronous:
 * hash created
 * tmpDir created
 * cbx file opened
 * cbx file extracted
 * Errors during this phase are considered critical, and stop the process 
 * The ui is up and alive, but the user can't navigate until this phase signals
 * completion either success or failure. If the result is success LoadCbx is invoked,
 * see below
 *
 */
func (m *Model) OpenCbxFile() {
    m.SendMessage(util.Message{TypeName: "render"})

    hash, err := util.HashFile(m.FilePath)
    if err != nil {
        m.sendOpenFileResMsg(-1, fmt.Sprintf("Error opening file; %s", err))
        return
    }
    m.Hash = hash

    td, err := util.CreateTmpDir()
    if err != nil {
        m.sendOpenFileResMsg(-11, fmt.Sprintf("Error creating tmp dir; %s", err))
        return
    }
    m.TmpDir = td

    ip, err := util.GetImagePaths(m.FilePath, m.TmpDir)
    if err != nil {
        m.sendOpenFileResMsg(-21, fmt.Sprintf("Error extracting cbx file; %s", err))
        return
    }
    m.ImgPaths = ip

    m.sendOpenFileResMsg(0, "Success")
}

/*
 * The second phase of the "loading" process is the actual loading and 
 * its synchronous. We have a lot of stuff to load:
 * individual pages 
 * page metadata 
 * layout
 * bookmarks
 *
 * It has to be synchronous because the pixbufs that are loaded into the 
 * pages can't be touched on anything but the event dispatch thread or they 
 * leak.
 * 
 * Errors during this phase are treated as warnings. For two reasons, some of 
 * this stuff is optional and there could be hundreds of errors on a per page 
 * basis. If the program can be useful to the user in spite of that it tries.
 * 
 * Finally there is the serieslist which is just an exception in that it really
 * is optional if it can't be calculated the program just does without it, see 
 * loadSeriesList
 */
func (m *Model) LoadCbxFile() {
    m.NewPages()
    m.SpreadIndex = 0
    m.PageIndex = 0

    m.joinAll()
    lo := m.loadLayout(m.Hash)
    if lo != nil {
        m.applyLayout(lo)
    }

    m.NewSpreads()

    m.loadBookmarks()

    m.SendMessage(util.Message{TypeName: "render"})
}

func (m *Model) CloseCbxFile() {
    m.StoreLayout()
    os.RemoveAll(m.TmpDir)
    m.Hash = ""
    m.ImgPaths = nil
    m.Pages = nil
    m.Spreads = nil
    m.SpreadIndex = 0
    m.PageIndex = 0
    m.Bookmarks = nil
    m.SeriesList = nil
    m.SeriesIndex = 0
    debug.FreeOSMemory()
}

// Walk the layout and load/unload pages as needed
func (m *Model) RefreshPages() {
    if m.LayoutMode != LONG_STRIP {
        start := int(math.Max(0, float64(m.SpreadIndex-(MAX_LOAD/2)+1)))
        end := int(math.Min(float64(m.SpreadIndex+(MAX_LOAD/2)-1), float64(len(m.Spreads)-1)))

        // iterate over all spreads
        // load/unload pgs as needed
        for i := range m.Spreads {
            spread := m.Spreads[i]
            if i < start || i > end {
                for j := range spread.Pages {
                    if spread.Pages[j].Loaded {
                        spread.Pages[j].Image = nil
                        spread.Pages[j].Loaded = false
                    }
                }
            } else {
                for j := range spread.Pages {
                    if !spread.Pages[j].Loaded {
                        spread.Pages[j].Load()
                    }
                }
            }
        }

        if util.DEBUG {
            m.printLoaded()
        }
    } else {
        // load all pages
        for i := range m.Pages {
            page := &m.Pages[i]
            if !page.Loaded {
                page.Load()
            }
        }
    }
}

// Returns 0 if spreads are nil or 
// page can't be found 
// Otherwise it guesses
// fixme: Don't like it, but there's no point in telling a user about any of it 
// because errors detected here are almost certainly the result of a 
// programming error elsewhere in the program
func (m *Model) PageToSpread(n int) int {
    if m.Spreads == nil {
        util.Log("p2s: spreads nil %d\n", n)
        return 0
    } 

    if n < 0 {
        util.Log("p2s: page out of range %d\n", n)
        return 0
    }

    var pagesNil bool
    if m.Pages == nil {
        pagesNil = true
        util.Log("p2s: pages nil %d\n", n)
    }

    if !pagesNil && n > len(m.Pages) - 1 {
        max := len(m.Pages) - 1
        util.Log("p2s: page out of range max: %d, n:%d\n", max, n)
        return len(m.Spreads) - 1
    }

    for i := range m.Spreads {
        spread := m.Spreads[i]
        for j := range spread.PageIdxs {
            if n == spread.PageIdxs[j] {
                return i
            }
        }
    }

    util.Log("p2s: page not found %d\n", n)
    return  0
}

func (m *Model) loadBookmarks() {
    m.Bookmarks = NewBookmarkList(m.FilePath)
    m.Bookmarks.Load(m.Hash)
    m.SendMessage(util.Message{TypeName: "render"})
}

func (m *Model) loadLayout(hash string) *Layout {
    data, _ := util.ReadLayout(hash)

    if data != nil {
        var lo Layout
        err := json.Unmarshal([]byte(*data), &lo)
        if err != nil {
            fmt.Printf("e:%s\n", err)
        }
        return &lo
    }
    return nil
}

func (m *Model) joinAll() {
    for i := range m.Pages {
        p := m.Pages[i]
        if p.Width >= p.Height {
            p.Span = DOUBLE
        }
        m.Pages[i] = p
    }
}

func (m *Model) applyLayout(layout *Layout) {
    for i := range layout.Pages {
        p := layout.Pages[i]
        mp := m.Pages[i]
        mp.Span = p.Span
        mp.Hidden = p.Hidden
        m.Pages[i] = mp
    }
}

func (m *Model) StoreLayout() error {
    layout := Layout{
        FormatVersion: "0.1",
    }
    c := ComicData{}
    c.Hash = m.Hash
    c.FilePath = m.FilePath
    layout.Comic = c
    layout.Mode = m.LayoutMode
    layout.Pages = m.Pages

    data, err := json.Marshal(layout)
    if err != nil {
        return err
    }

    err = util.WriteLayout(m.Hash, string(data))
    if err != nil {
        return err
    }
    return nil
}

// Make sure we always send a result message, no errors allowed
func (m *Model) sendOpenFileResMsg(code ResultCode, description string) {
    var d string
    r := Result{code, description}
    buf, err := json.Marshal(r)
    if err != nil {
        d = fmt.Sprintf("{\"code\":%d,\"result\":\"%s\"}", r.Code, r.Description)
    } else {
        d = string(buf)
    }

    m.SendMessage(util.Message{TypeName: "openFileResult", Data: d})
}

// dbg
func (m *Model) checkSpreads() {
    c := 0
    for x := range m.Pages {
        if !m.Pages[x].Loaded {
            if m.Pages[x].Image != nil {
                fmt.Printf("Page %d shouldn't be loaded\n", x)
                m.Pages[x].Image = nil
            }
        } else {
            c++
        }
    }
}

// dbg
func (m *Model) printLoaded() {
    var buf string
    for i := range m.Pages {
        if !m.Pages[i].Loaded {
            buf += "0"
        } else {
            if i == m.PageIndex {
                buf += "_"
            } else {
                buf += "1"
            }
        }
    }
    fmt.Printf("%s\n", buf)
}

