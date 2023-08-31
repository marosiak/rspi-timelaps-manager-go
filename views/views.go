package views

import "embed"

//go:embed *
var ViewsFileSystem embed.FS

func GetViewsFileSystem() embed.FS {
	return ViewsFileSystem
}
