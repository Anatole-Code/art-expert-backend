package model_artworkdb

import "time"

type Connoisseur struct {
	ConnoisseurID    int32  `gorm:"column:connoisseur_id;primaryKey;autoIncrement;type:integer"`
	ConnoisseurLogin string `gorm:"column:connoisseur_login;type:varchar(50);not null;uniqueIndex"`
}

func (Connoisseur) TableName() string { return "connoisseurs" }

type Artwork struct {
	ArtworkID       int32      `gorm:"column:artwork_id;primaryKey;autoIncrement;type:integer"`
	ArtworkTitle    string     `gorm:"column:artwork_title;type:varchar(120);not null"`
	AttributionNote *string    `gorm:"column:attribution_note;type:varchar(500)"`
	ArtworkStatus   string     `gorm:"column:artwork_status;type:varchar(32);not null"`
	ArtworkImageURL *string    `gorm:"column:artwork_image_url;type:varchar(300)"`
	ArtworkVideoURL *string    `gorm:"column:artwork_video_url;type:varchar(300)"`
	CreatedOn       *time.Time `gorm:"column:created_on;type:date"`
	DiagonalCm      *int32     `gorm:"column:diagonal_cm;type:integer"`
	PainterName     *string    `gorm:"column:painter_name;type:varchar(80)"`
	RecordedAt      time.Time  `gorm:"column:recorded_at;type:timestamp;not null"`
	ConnoisseurID   int32      `gorm:"column:connoisseur_id;not null;type:integer"`
	NoteFormedAt    *time.Time `gorm:"column:note_formed_at;type:timestamp"`
}

func (Artwork) TableName() string { return "artworks" }

type ArtworkMark struct {
	ArtworkMarkID int32 `gorm:"column:artwork_mark_id;primaryKey;autoIncrement;type:integer"`
	ArtworkID     int32 `gorm:"column:artwork_id;not null;type:integer;uniqueIndex:artwork_marks_pair_unique"`
	ConnoisseurID int32 `gorm:"column:connoisseur_id;not null;type:integer;uniqueIndex:artwork_marks_pair_unique"`
}

func (ArtworkMark) TableName() string { return "artwork_marks" }
