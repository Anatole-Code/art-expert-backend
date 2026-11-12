package store_artworkdb

import (
	"fmt"

	"github.com/Anatole-Code/art-expert-backend/internal/artworks/model_artworkdb"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("пустая строка подключения к базе картин")
	}
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func Migrate(db *gorm.DB) error {
	drops := []string{
		`ALTER TABLE artwork_marks DROP CONSTRAINT IF EXISTS artwork_marks_artwork_fk`,
		`ALTER TABLE artwork_marks DROP CONSTRAINT IF EXISTS artwork_marks_connoisseur_fk`,
		`ALTER TABLE artworks DROP CONSTRAINT IF EXISTS artworks_connoisseur_fk`,
	}
	for _, statement := range drops {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	if err := db.AutoMigrate(
		&model_artworkdb.Connoisseur{},
		&model_artworkdb.Artwork{},
		&model_artworkdb.ArtworkMark{},
	); err != nil {
		return err
	}
	statements := []string{
		`ALTER TABLE artworks ALTER COLUMN diagonal_cm TYPE integer`,
		`ALTER TABLE artworks ALTER COLUMN connoisseur_id TYPE integer`,
		`ALTER TABLE artwork_marks ALTER COLUMN artwork_id TYPE integer`,
		`ALTER TABLE artwork_marks ALTER COLUMN connoisseur_id TYPE integer`,
		`ALTER TABLE artworks ADD CONSTRAINT artworks_connoisseur_fk FOREIGN KEY (connoisseur_id) REFERENCES connoisseurs (connoisseur_id) ON DELETE RESTRICT ON UPDATE RESTRICT`,
		`ALTER TABLE artwork_marks ADD CONSTRAINT artwork_marks_artwork_fk FOREIGN KEY (artwork_id) REFERENCES artworks (artwork_id) ON DELETE RESTRICT ON UPDATE RESTRICT`,
		`ALTER TABLE artwork_marks ADD CONSTRAINT artwork_marks_connoisseur_fk FOREIGN KEY (connoisseur_id) REFERENCES connoisseurs (connoisseur_id) ON DELETE RESTRICT ON UPDATE RESTRICT`,
		`ALTER TABLE artworks DROP CONSTRAINT IF EXISTS artworks_status_known`,
		`ALTER TABLE artworks ADD CONSTRAINT artworks_status_known CHECK (artwork_status IN ('artwork_draft', 'artwork_published', 'artwork_removed'))`,
		`CREATE UNIQUE INDEX IF NOT EXISTS artworks_one_draft_per_connoisseur ON artworks (connoisseur_id) WHERE artwork_status = 'artwork_draft'`,
		`CREATE INDEX IF NOT EXISTS artworks_created_on_idx ON artworks (created_on)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	return nil
}
