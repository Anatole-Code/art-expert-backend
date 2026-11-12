package model_artworks

import "errors"

func CreatorConnoisseurID() int {
	return 1
}

var (
	ErrArtworkTitleMissing          = errors.New("нужно название картины")
	ErrArtworkDraftExists           = errors.New("черновик уже есть")
	ErrArtworkDraftMissing          = errors.New("черновика нет")
	ErrArtworkPublicationIncomplete = errors.New("нужны описание, дата и диагональ")
	ErrArtworkNotPublished          = errors.New("опубликованной картины нет")
)
