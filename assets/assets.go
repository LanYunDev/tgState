package assets

import "embed"

var (
	//go:embed templates_min
	Templates embed.FS
)
