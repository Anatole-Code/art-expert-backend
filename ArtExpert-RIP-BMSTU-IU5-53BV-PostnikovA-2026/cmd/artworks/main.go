package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/controller_artworkscreens"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
)

func main() {
	artworkRoot, err := artworkComposeRoot()
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := model_artworks.UploadArtworkFiles(ctx, "127.0.0.1:9000", "root", "rootpassword", artworkRoot); err != nil {
		log.Printf("хранилище картин недоступно, экраны покажут файлы коллекции: %v", err)
	}
	artworkScreens := controller_artworkscreens.OpenArtworkScreens(model_artworks.NewMemoryArtworkLookup(), "http://127.0.0.1:9000", filepath.Join(artworkRoot, "artworks-files"), time.Now())
	log.Print("экраны картин слушают http://127.0.0.1:8088")
	log.Fatal(http.ListenAndServe("127.0.0.1:8088", artworkScreens))
}

func artworkComposeRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "docker-compose.yml")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
