module github.com/mftb0/cbxv

go 1.18

require (
	github.com/gen2brain/go-unarr v0.2.0
	github.com/gotk3/gotk3 v0.6.2
	github.com/pdfcpu/pdfcpu v0.6.0
)

require (
	github.com/hhrutter/lzw v1.0.0 // indirect
	github.com/hhrutter/tiff v1.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.4 // indirect
	golang.org/x/image v0.12.0 // indirect
	golang.org/x/text v0.13.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/mftb0/cbxv/internal/util => ../cbxv

replace github.com/mftb0/cbxv/internal/ui => ../cbxv

replace github.com/mftb0/cbxv/internal/model => ../cbxv
