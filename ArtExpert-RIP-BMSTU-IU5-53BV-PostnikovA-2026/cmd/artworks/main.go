package main

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"time"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/config_artworkserver"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/controller_artworkscreens"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/dsn_artworkdb"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworks"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/store_artworkdb"
	"github.com/joho/godotenv"
)

func main() {
	artworkRoot, err := config_artworkserver.FindProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	serverConfig, err := config_artworkserver.Load(artworkRoot)
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := godotenv.Load(filepath.Join(artworkRoot, ".env")); err != nil {
		log.Fatal(err)
	}
	artworkDB, err := store_artworkdb.Open(dsn_artworkdb.FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	if err := model_artworks.UploadArtworkFiles(ctx, "127.0.0.1:9000", "root", "rootpassword", artworkRoot); err != nil {
		log.Printf("хранилище картин недоступно, экраны покажут файлы коллекции: %v", err)
	}
	artworkScreens := controller_artworkscreens.OpenArtworkScreens(store_artworkdb.NewArtworkLookup(artworkDB), "http://127.0.0.1:9000", filepath.Join(artworkRoot, "artworks-files"), time.Now())
	log.Printf("экраны картин слушают http://%s", serverConfig.ListenAddress())
	log.Fatal(http.ListenAndServe(serverConfig.ListenAddress(), artworkScreens))
}
