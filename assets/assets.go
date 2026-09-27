// Package assets embeds the new catalogue illustrations in the executable.
package assets

import (
	"embed"
	"fyne.io/fyne/v2"
	"path"
)

//go:embed photos/*.png brand.svg
var files embed.FS

var photos = map[string]fyne.Resource{}

func init() {
	for _, name := range []string{"city.png", "trail.png", "road.png", "junior.png"} {
		data, _ := files.ReadFile("photos/" + name)
		photos[name] = fyne.NewStaticResource(name, data)
	}
}

func Photo(name string) fyne.Resource {
	if resource, ok := photos[path.Base(name)]; ok {
		return resource
	}
	return photos["city.png"]
}

func Logo() fyne.Resource {
	data, _ := files.ReadFile("brand.svg")
	return fyne.NewStaticResource("brand.svg", data)
}
