package util

import (
    "fmt"

	"github.com/gotk3/gotk3/gdk"
	"github.com/gotk3/gotk3/glib"
	"github.com/gotk3/gotk3/gtk"
)

type Img = gdk.Pixbuf
type ImgFormat = gdk.PixbufFormat

func ImgNewFromFile(path string) (*Img, error) {
    return gdk.PixbufNewFromFile(path)
}

func ImgGetFileInfo(path string) (*ImgFormat, int, int, error) {
    return gdk.PixbufGetFileInfo(path)
}

type PBCache struct {
    max int
    evict string
    items map[string]*gdk.Pixbuf
}

func NewPBCache(max int) *PBCache {
    return &PBCache{
        max: max,
        items: make(map[string]*gdk.Pixbuf, max),
    }
}

func (c *PBCache) Put(key string, pb *gdk.Pixbuf) {
    if len(c.items) == c.max {
        delete(c.items, c.evict)
        Log("evicted %s\n", c.evict)
    }

    c.items[key] = pb

    if len(c.items) == c.max {
        c.evict = key
    }
}

func (c *PBCache) Get(key string) *gdk.Pixbuf {
    return c.items[key]
}

func (c *PBCache) Clear() {
    for k := range c.items {delete(c.items, k)}
}

func CreateLabel(text string, cssClass string, toolTip *string) *gtk.Label {
    c, err := gtk.LabelNew(text)
    if err != nil {
        fmt.Printf("Error creating %s label %s\n", text, err)
    }
    if toolTip != nil {
        c.SetTooltipMarkup(*toolTip)
    }
    css, _ := c.GetStyleContext()
    css.AddClass(cssClass)
    return c
}

func CreateButton(text string, cssClass string, toolTip *string) *gtk.Button {
    c, err := gtk.ButtonNewWithLabel(text)
    if err != nil {
        fmt.Printf("Error creating label %s\n", err)
    }
    c.SetCanFocus(false)
    if toolTip != nil {
        c.SetTooltipText(*toolTip)
    }
    css, _ := c.GetStyleContext()
    css.AddClass(cssClass)
    return c
}

func assertEventDispatchThread() bool {
    if glib.MainContextDefault().IsOwner() {
        return true
    }
    return false
}

