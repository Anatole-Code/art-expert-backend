package controller_artworkscreens

import (
	"html/template"
	"net/http"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/view_artworkscreens"
)

type ArtworkScreens struct {
	artworkLookup   *model_artworks.MemoryArtworkLookup
	artworkFileBase string
	artworkTmpl     *template.Template
	openedOn        time.Time
}

func OpenArtworkScreens(artworkLookup *model_artworks.MemoryArtworkLookup, artworkFileBase, artworkDefaultDir string, openedOn time.Time) http.Handler {
	screens := &ArtworkScreens{
		artworkLookup:   artworkLookup,
		artworkFileBase: artworkFileBase,
		artworkTmpl:     template.Must(template.ParseFS(view_artworkscreens.Files, "templates/*.html")),
		openedOn:        time.Date(openedOn.Year(), openedOn.Month(), openedOn.Day(), 0, 0, 0, 0, time.UTC),
	}
	artworkMux := http.NewServeMux()
	artworkMux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/artworks", http.StatusFound)
	})
	artworkMux.HandleFunc("GET /artworks", screens.showArtworkFeed)
	artworkMux.HandleFunc("GET /artworks/{id}", screens.showArtworkFeed)
	artworkMux.HandleFunc("GET /artworks-draft-form", screens.showArtworkDraftForm)
	artworkMux.HandleFunc("GET /artworks-created-on-tile", screens.showCreatedOnTile)
	style, err := artworkStyleFS()
	if err != nil {
		panic(err)
	}
	artworkMux.Handle("GET /artworks-screens/", http.StripPrefix("/artworks-screens/", http.FileServer(http.FS(style))))
	if artworkDefaultDir != "" {
		artworkMux.Handle("GET /artworks-default/", http.StripPrefix("/artworks-default/", http.FileServer(http.Dir(artworkDefaultDir))))
	}
	return artworkMux
}
