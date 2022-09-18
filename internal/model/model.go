package model

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"os"
	"sort"

    "example.com/cbxv-gotk3/internal/util"
)

// Data model of a cbx application
// Composed of a handful of sub-models, collections and other standard types
type Model struct {
    SendMessage util.Messenger
    FilePath string
    TmpDir string
    Hash string
    Bookmarks *BookmarkList
    ImgPaths []string
    Pages []Page
    SelectedPage int
    Leaves []*Leaf
    CurrentLeaf int
    ReadMode ReadMode
    LeafMode LeafMode
    SeriesList []string
    SeriesIndex int
    BrowseDirectory string
    Fullscreen bool
}

func NewModel(messenger util.Messenger) *Model {
    m := &Model{}
    m.SendMessage = messenger
    m.BrowseDirectory, _ = os.Getwd()
    return m
}

// ReadMode of the model is either
// Left-To-Right or
// Right-To-Left
type ReadMode int

const (
    LTR ReadMode = iota
    RTL
)

// Maximum to load into the model
// Currently confusing because at init the units are pages
// Later it's leaves (so frequently double the number of pagesj
const (
    MAX_LOAD = 8
)

// Mark a place in the model by keeping track of an index in the pages slice
type Bookmark struct {
    PageIndex int `json:"pageIndex"`
    CreationTime int64 `json:"creationTime"`
}

// Just a little type for serialization
type BookmarkListComic struct {
    Hash string `json:"hash"`
    FilePath string `json:"filePath"`
}

// Type used for serialization/deserialization of the BookmarkList
type BookmarkListModel struct {
    FormatVersion string `json:"formatVersion"`
    Comic BookmarkListComic `json:"comic"`
    Bookmarks []Bookmark `json:"bookmarks"`
}

// Manage a list of bookmarks
type BookmarkList struct {
    Model BookmarkListModel
}

func NewBookmarkList(filePath string) *BookmarkList {
    b := BookmarkList{}
    m := BookmarkListModel {
        FormatVersion : "0.1",
    }
    c := BookmarkListComic{}
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
            break;
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
            break;
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
// They are grouped on Leaves
type Page struct {
    FilePath string `json:"filePath"`
    Width int`json:"width"`
    Height int`json:"height"`
    Orientation int `json:"orientation"`
    Loaded bool `json:"loaded"`
    Image *image.Image `json:"-"` 
}

func (p *Page) Load() {
    f, err := util.LoadImageFile(p.FilePath)
    if err != nil {
        fmt.Printf("Error loading file %s\n", err)
    }
    p.Image = &f
    b := f.Bounds()
    p.Width = b.Dx()
    p.Height = b.Dy()
    if p.Width >= p.Height {
        p.Orientation = LANDSCAPE
    }
    p.Loaded = true
}

func (p *Page) LoadMeta() {
    f, err := util.LoadImageFileMeta(p.FilePath)
    if err != nil {
        fmt.Printf("Error loading file %s\n", err)
    }
    p.Width = f.Width
    p.Height = f.Height
    if p.Width >= p.Height {
        p.Orientation = LANDSCAPE
    }
    p.Loaded = false
}

// Creates pgs slice and loads it
func (m *Model) NewPages() {
    pages := make([]Page, len(m.ImgPaths))

    for i := range m.ImgPaths {
        pages[i].FilePath = m.ImgPaths[i]
        pages[i].Orientation = PORTRAIT
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
type Orientation int

const (
    PORTRAIT = iota
    LANDSCAPE
)

// Leaf here is borrowed from publishing
// and can generally be thought of like a sheet of paper,
// on which any number of pages may be printed.
// In a folio binding for instance there are generrally
// 4 pages on each leaf (2 on the front, 2 on the back)
// In our case the leafMode determines how many pages per leaf
type Leaf struct {
    Pages []*Page
}

// Creates leaf slice based on pg slice and leaf mode
func (m *Model) NewLeaves() {
    var leaves []*Leaf

    pages := m.Pages
    if(m.LeafMode == ONE_PAGE) {
        for i := range pages {
            leaf := &Leaf{}
            p := &pages[i]
            leaf.Pages = append(leaf.Pages, p)
            leaves = append(leaves, leaf)
        }
    } else if m.LeafMode == TWO_PAGE {
        for i := 0; i < len(pages); i++ {
            // create leaf add a page
            leaf := &Leaf{}
            p := &pages[i]
            leaf.Pages = append(leaf.Pages, p)

            // if pg is landscape, leaf done
            if p.Orientation == LANDSCAPE {
                leaves = append(leaves, leaf)
                continue
            }
            // if pg is last page, leaves done
            if i == (len(pages) - 1) {
                leaves = append(leaves, leaf)
                break
            }

            // on to the next page
            i++
            p = &pages[i]

            // if pg is landscape, make a new leaf, leaf done
            if p.Orientation == LANDSCAPE {
                leaves = append(leaves, leaf)
                leaf = &Leaf{}
                leaf.Pages = append(leaf.Pages, p)
                leaves = append(leaves, leaf)
                continue
            }

            // No special cases, so leaf with 2 pages
            leaf.Pages = append(leaf.Pages, p)
            leaves = append(leaves, leaf)
        }
    } else {
        // Put all pages on one leaf
        leaf := &Leaf{}
        for i := range pages {
            p := &pages[i]
            leaf.Pages = append(leaf.Pages, p)
        }
        leaves = append(leaves, leaf)
    }

    m.Leaves = leaves
}

// Leaf mode determines how many pages per leaf
// ONE_PAGE = 1 pg
// TWO_PAGE = 2 pgs
// LONG_STRIP = n pgs
type LeafMode int

const (
    ONE_PAGE = iota
    TWO_PAGE
    LONG_STRIP
)

// Uses leaves to determine which pages to load or unload
func (m *Model) RefreshPages() {
    if m.LeafMode != LONG_STRIP {
        // load the current leaf
        leaf := m.Leaves[m.CurrentLeaf]
        for i := range leaf.Pages {
            if !leaf.Pages[i].Loaded {
                leaf.Pages[i].Load()
            }
        }

        // buffer nearby leaves
        start := int(math.Max(0, float64(m.CurrentLeaf-(MAX_LOAD/2))))
        end := int(math.Min(float64(m.CurrentLeaf+(MAX_LOAD/2)), float64(len(m.Leaves))))
        for j := start; j < end; j++  {
            leaf = m.Leaves[j]
            for i := range leaf.Pages {
                if !leaf.Pages[i].Loaded {
                    leaf.Pages[i].Load()
                }
            }
        }

        // remove distant leaves
        for j := range m.Leaves {
            if j < start || j > end {
                leaf := m.Leaves[j]
                for i := range leaf.Pages {
                    if leaf.Pages[i].Loaded {
                        //leave image metadata alone
                        util.Log("Unload pg %d\n", j);
                        leaf.Pages[i].Image = nil
                        leaf.Pages[i].Loaded = false
                    }
                }
            }
        }

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
    } else {
        // load all pages
        for i := range m.Pages {
            m.Pages[i].Load()
        }
    }
}

// for the current leaf find the "lesser" or
// verso page number
func (m *Model) CalcVersoPage() int {
    r := 0
    if(m.LeafMode == ONE_PAGE) {
        r = m.CurrentLeaf
    } else if m.LeafMode == TWO_PAGE {
        if m.Leaves == nil {
            return 0
        }
        for i := 0; i < m.CurrentLeaf; i++ {
            leaf := m.Leaves[i]
            r += len(leaf.Pages)
        }
    } else {
        r = m.CurrentLeaf
    }
    return r
}

// page index to leaf index
func (m *Model) PageToLeaf(n int) int {
    if m.Leaves == nil {
        return 0
    } else if m.LeafMode == TWO_PAGE {
        var p = 0
        for i := range m.Leaves {
            leaf := m.Leaves[i]
            p += len(leaf.Pages)
            if p > n {
                return i
            }
        }
    }
    return n;
}

func (m *Model) loadBookmarks() {
    m.Bookmarks = NewBookmarkList(m.FilePath)
    m.Bookmarks.Load(m.Hash)
    msg := &util.Message{TypeName: "render"}
    m.SendMessage(*msg)
}

func (m *Model) LoadHash() {
    hash, err := util.HashFile(m.FilePath)
    if err != nil {
        fmt.Printf("Unable to compute file hash %s\n", err)
    }
    m.Hash = hash
    m.loadBookmarks()
}

