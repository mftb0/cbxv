package main

import (
	"encoding/json"
	"fmt"
	"image"
	"math"
	"sort"
)

// Data model of a cbx application
// Composed of a handful of sub-models, collections and other standard types
type Model struct {
    filePath string
    tmpDir string
    hash string
    bookmarks *BookmarkList
    imgPaths []string
    pages []Page
    selectedPage int
    leaves []*Leaf
    currentLeaf int
    readMode ReadMode
    leafMode LeafMode
    seriesList []string
    seriesIndex int
    browseDirectory string
    fullscreen bool
}

func NewModel() *Model {
    m := &Model{}
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
    model BookmarkListModel
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
    b.model = m
    return &b
}

func (l *BookmarkList) Add(b Bookmark) {
    for i := 0; i < len(l.model.Bookmarks); i++ {
        if l.model.Bookmarks[i].PageIndex == b.PageIndex {
            return
        }
    }
    l.model.Bookmarks = append(l.model.Bookmarks, b)
    sort.Slice(l.model.Bookmarks, func(i, j int) bool {
        return l.model.Bookmarks[i].PageIndex < l.model.Bookmarks[j].PageIndex
    })
    l.Store()
}

func (l *BookmarkList) Remove(b Bookmark) *Bookmark {
    var x int
    var r Bookmark
    for i := 0; i < len(l.model.Bookmarks); i++ {
        if l.model.Bookmarks[i].PageIndex == b.PageIndex {
            r = l.model.Bookmarks[i]
            x = i
            break;
        }
    }
    l.model.Bookmarks = append(l.model.Bookmarks[:x], l.model.Bookmarks[x+1:]...)
    l.Store()
    return &r
}

func (l *BookmarkList) Find(pageIndex int) *Bookmark {
    var r *Bookmark
    for i := 0; i < len(l.model.Bookmarks); i++ {
        if l.model.Bookmarks[i].PageIndex == pageIndex {
            r = &l.model.Bookmarks[i]
            break;
        }
    }
    return r
}

func (l *BookmarkList) Store() error {
    data, err := json.Marshal(l.model)
    if err != nil {
        return err
    }
    err = WriteBookmarkList(l.model.Comic.Hash, string(data))
    if err != nil {
        return err
    }
    return nil
}

func (l *BookmarkList) Load(hash string) {
    l.model.Comic.Hash = hash
    data, _ := ReadBookmarkList(l.model.Comic.Hash)

    if data != nil {
        var m BookmarkListModel
        err := json.Unmarshal([]byte(*data), &m)
        if err != nil {
            fmt.Printf("e:%s\n", err)
        }
        l.model = m
    }
}

// A page in this case is generally analogous to an image
// They are grouped on Leaves
type Page struct {
    filePath string
    Width int
    Height int
    Orientation int
    Loaded bool
    Image *image.Image
}

func (p *Page) Load() {
    f, err := loadImageFile(p.filePath)
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

// Creates pgs slice and loads it
func NewPages(model *Model) []Page {
    pages := make([]Page, len(model.imgPaths))

    for i := range model.imgPaths {
        pages[i].filePath = model.imgPaths[i]
        pages[i].Orientation = PORTRAIT
        pages[i].Loaded = false
        if i < MAX_LOAD {
            pages[i].Load()
        }
    }

    return pages
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
    pages []*Page
}

// Creates leaf slice based on pg slice and leaf mode
func NewLeaves(model *Model) []*Leaf {
    var leaves []*Leaf

    pages := model.pages
    if(model.leafMode == ONE_PAGE) {
        for i := range pages {
            leaf := &Leaf{}
            p := &pages[i]
            leaf.pages = append(leaf.pages, p)
            leaves = append(leaves, leaf)
        }
    } else if model.leafMode == TWO_PAGE {
        for i := 0; i < len(pages); i++ {
            // create leaf add a page
            leaf := &Leaf{}
            p := &pages[i]
            leaf.pages = append(leaf.pages, p)

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
                leaf.pages = append(leaf.pages, p)
                leaves = append(leaves, leaf)
                continue
            }

            // No special cases, so leaf with 2 pages
            leaf.pages = append(leaf.pages, p)
            leaves = append(leaves, leaf)
        }
    } else {
        // Put all pages on one leaf
        leaf := &Leaf{}
        for i := range pages {
            p := &pages[i]
            leaf.pages = append(leaf.pages, p)
        }
        leaves = append(leaves, leaf)
    }

    return leaves
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
func RefreshPages(model *Model) {
    if model.leafMode != LONG_STRIP {
        // load the current leaf
        leaf := model.leaves[model.currentLeaf]
        for i := range leaf.pages {
            if !leaf.pages[i].Loaded {
                leaf.pages[i].Load()
            }
        }

        // buffer nearby leaves
        start := int(math.Max(0, float64(model.currentLeaf-(MAX_LOAD/2))))
        end := int(math.Min(float64(model.currentLeaf+(MAX_LOAD/2)), float64(len(model.leaves))))
        for j := start; j < end; j++  {
            leaf = model.leaves[j]
            for i := range leaf.pages {
                if !leaf.pages[i].Loaded {
                    leaf.pages[i].Load()
                }
            }
        }

        // remove distant leaves
        for j := range model.leaves {
            if j < start || j > end {
                leaf := model.leaves[j]
                for i := range leaf.pages {
                    if leaf.pages[i].Loaded {
                        leaf.pages[i].Image = nil
                        leaf.pages[i].Loaded = false
                    }
                }
            }
        }

        c := 0
        for x := range model.pages {
            if !model.pages[x].Loaded {
                if model.pages[x].Image != nil {
                    fmt.Printf("Page %d shouldn't be loaded\n", x)
                    model.pages[x].Image = nil
                }
            } else {
                c++
            }
        }
    } else {
        // load all pages
        for i := range model.pages {
            model.pages[i].Load()
        }
    }
}

func calcVersoPage(model *Model) int {
    r := 0
    if(model.leafMode == ONE_PAGE) {
        r = model.currentLeaf
    } else if model.leafMode == TWO_PAGE {
        if model.leaves == nil {
            return 0
        }
        for i := 0; i < model.currentLeaf; i++ {
            leaf := model.leaves[i]
            r += len(leaf.pages)
        }
    } else {
        r = model.currentLeaf
    }
    return r
}

