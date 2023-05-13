module github.com/mftb0/cbxv

go 1.18

require (
	github.com/gen2brain/go-unarr v0.1.6
	github.com/gotk3/gotk3 v0.6.2
	github.com/pdfcpu/pdfcpu v0.4.0
)

require (
	github.com/hhrutter/lzw v0.0.0-20190829144645-6f07a24e8650 // indirect
	github.com/hhrutter/tiff v0.0.0-20190829141212-736cae8d0bc7 // indirect
	github.com/mattn/go-runewidth v0.0.14 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.3 // indirect
	golang.org/x/image v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/mftb0/cbxv/internal/util => ../cbxv

replace github.com/mftb0/cbxv/internal/ui => ../cbxv

replace github.com/mftb0/cbxv/internal/model => ../cbxv
