package model

import (
    "encoding/json"
    "fmt"
    _ "image"
    "math"
    "os"
    "sort"

    "github.com/gotk3/gotk3/gdk"
    "github.com/mftb0/cbxv-gotk3/internal/util"
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
    Image    *gdk.Pixbuf `json:"-"`
}

func (p *Page) Load() {
    //    f, frmt, err := util.LoadImageFile(p.FilePath)
    f, err := gdk.PixbufNewFromFile(p.FilePath)
    if err != nil {
        fmt.Printf("Error loading file %s\n", err)
    }
    p.Image = f
    p.Width = f.GetWidth()
    p.Height = f.GetHeight()
    p.Loaded = true
}

func (p *Page) LoadMeta() {
    _, w, h, err := gdk.PixbufGetFileInfo(p.FilePath)
    if err != nil {
        fmt.Printf("Error loading file %s\n", err)
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
    m.Loading = false
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

            // skip hidden
            for ; i < len(pages); i++ {
                p = &pages[i]
                if p.Hidden {
                    m.HiddenPages = true
                    continue
                } else {
                    break
                }
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

            // No special cases, so spread with 2 pages
            spread.Pages = append(spread.Pages, p)
            spread.PageIdxs = append(spread.PageIdxs, i)
            spreads = append(spreads, spread)
        }
    } else {
        // Put all pages on one spread
        spread := &Spread{}
        for i := range pages {
            p := &pages[i]
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

/*
 * We have a lot of stuff to load hash, bookmarks, serieslist, cbx,
 * individual pages, page metadata and layout.
 * Error handling - All Load functions are responsible for handling any
 * errors they encounter. They currently handle them by logging to the console.
 * The load process starts in the openFile Command when the Loading property
 * is set true on the Model. It ends in NewPages after we've attemped to load
 * the inital pages and metadata, we set Loading to false. There may still be
 * some stuff loading async, but by that point the user should be able to know
 * that either the attempt failed or that they can reasonably start paging
 * through their comic.
 */
func (m *Model) loadBookmarks() {
    m.Bookmarks = NewBookmarkList(m.FilePath)
    m.Bookmarks.Load(m.Hash)
    msg := &util.Message{TypeName: "render"}
    m.SendMessage(*msg)
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

func (m *Model) LoadHash() {
    hash, err := util.HashFile(m.FilePath)
    if err != nil {
        fmt.Printf("Unable to compute file hash %s\n", err)
    }
    m.Hash = hash
    m.loadBookmarks()
}

func (m *Model) LoadSeriesList() {
    s, err := util.ReadSeriesList(m.FilePath)
    if err != nil {
        fmt.Printf("Unable to load series list %s\n", err)
    }
    m.SeriesList = s

    for i := range s {
        if m.FilePath == s[i] {
            m.SeriesIndex = i
        }
    }
    m.PageIndex = m.CalcVersoPage()
}

func (m *Model) LoadCbxFile() {
    td, err := util.CreateTmpDir()
    if err != nil {
        fmt.Printf("Unable to create tmp dir %s\n", err)
        return
    }
    m.TmpDir = td

    ip, err := util.GetImagePaths(m.FilePath, m.TmpDir)
    if err != nil {
        fmt.Printf("Unable to load cbx file %s\n", err)
    }
    m.ImgPaths = ip
    m.NewPages()
    m.SpreadIndex = 0
    m.PageIndex = 0
    m.joinAll()
    lo := m.loadLayout(m.Hash)
    if lo != nil {
        util.Log("Applying layout\n")
        m.applyLayout(lo)
    }
    m.NewSpreads()

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
            m.Pages[i].Load()
        }
    }
}

// For the current spread find the "lesser" or
// verso page number
func (m *Model) CalcVersoPage() int {
    r := 0
    if m.LayoutMode == ONE_PAGE {
        r = m.SpreadIndex
    } else if m.LayoutMode == TWO_PAGE {
        if m.Spreads == nil {
            return 0
        }

        for i := 0; i < m.SpreadIndex; i++ {
            spread := m.Spreads[i]
            r += len(spread.Pages)
        }
    } else {
        r = m.SpreadIndex
    }

    return r
}

// page index to spread index
func (m *Model) PageToSpread(n int) int {
    if m.Spreads == nil {
        return 0
    } else if m.LayoutMode == TWO_PAGE {
        var p = 0
        for i := range m.Spreads {
            spread := m.Spreads[i]
            p += len(spread.Pages)
            if p > n {
                return i
            }
        }
        // Couldn't find out of bounds
        if n > len(m.Spreads)-1 {
            n = len(m.Spreads) - 1
        }
    }
    return n
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
