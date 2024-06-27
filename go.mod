module github.com/mftb0/cbxv

go 1.18

require (
	github.com/gen2brain/go-unarr v0.2.3
	github.com/gotk3/gotk3 v0.6.5-0.20240618185848-ff349ae13f56
	github.com/pdfcpu/pdfcpu v0.8.0
)

require (
	github.com/hhrutter/lzw v1.0.0 // indirect
	github.com/hhrutter/tiff v1.0.1 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/image v0.15.0 // indirect
	golang.org/x/text v0.14.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/mftb0/cbxv/internal/util => ../cbxv

replace github.com/mftb0/cbxv/internal/ui => ../cbxv

replace github.com/mftb0/cbxv/internal/model => ../cbxv
