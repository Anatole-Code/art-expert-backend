package main

import (
	"log"
	"path/filepath"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/config_artworkserver"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/dsn_artworkdb"
	"github.com/Anatole-Code/art-expert-backend/internal/artworks/store_artworkdb"
	"github.com/joho/godotenv"
)

func main() {
	root, err := config_artworkserver.FindProjectRoot()
	if err != nil {
		log.Fatal(err)
	}
	if err := godotenv.Load(filepath.Join(root, ".env")); err != nil {
		log.Fatal(err)
	}
	db, err := store_artworkdb.Open(dsn_artworkdb.FromEnv())
	if err != nil {
		log.Fatal(err)
	}
	if err := store_artworkdb.Migrate(db); err != nil {
		log.Fatal(err)
	}
	log.Print("таблицы картин, знатоков и отметок готовы")
}
