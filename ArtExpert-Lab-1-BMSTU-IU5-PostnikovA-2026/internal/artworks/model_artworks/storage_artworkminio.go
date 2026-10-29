package model_artworks

import (
	"context"
	"log"
	"mime"
	"os"
	"path/filepath"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const ArtworkBucket = "artworks"

type ArtworkMinioSource struct {
	ArtworkKey  string
	ArtworkPath string
}

func UploadArtworkFiles(ctx context.Context, endpoint, accessKey, secretKey, artworkRoot string) error {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		return err
	}
	exists, err := client.BucketExists(ctx, ArtworkBucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := client.MakeBucket(ctx, ArtworkBucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}
	policy := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::artworks/*"]}]}`
	if err := client.SetBucketPolicy(ctx, ArtworkBucket, policy); err != nil {
		return err
	}
	for _, source := range artworkImageSources(artworkRoot) {
		putArtworkFile(ctx, client, source.ArtworkKey, source.ArtworkPath)
	}
	for _, source := range artworkVideoSources(artworkRoot) {
		putArtworkFile(ctx, client, source.ArtworkKey, source.ArtworkPath)
	}
	return nil
}

func artworkImageSources(artworkRoot string) []ArtworkMinioSource {
	folder := filepath.Join(artworkRoot, "artworks-files")
	return []ArtworkMinioSource{
		{"nobleman-hand-chest.jpg", filepath.Join(folder, "nobleman-hand-chest.jpg")},
		{"virgin-mary.jpg", filepath.Join(folder, "virgin-mary.jpg")},
		{"christ-carrying-cross.jpg", filepath.Join(folder, "christ-carrying-cross.jpg")},
		{"david-goliath.jpg", filepath.Join(folder, "david-goliath.jpg")},
		{"self-portrait-titian.jpg", filepath.Join(folder, "self-portrait-titian.jpg")},
		{"the-cardinal.jpg", filepath.Join(folder, "the-cardinal.jpg")},
		{"virgin-of-the-chair.jpg", filepath.Join(folder, "virgin-of-the-chair.jpg")},
		{"three-graces.jpg", filepath.Join(folder, "three-graces.jpg")},
	}
}

func artworkVideoSources(artworkRoot string) []ArtworkMinioSource {
	folder := filepath.Join(artworkRoot, "artworks-files")
	return []ArtworkMinioSource{
		{"nobleman-hand-chest.mp4", filepath.Join(folder, "nobleman-hand-chest.mp4")},
		{"virgin-mary.mp4", filepath.Join(folder, "virgin-mary.mp4")},
		{"christ-carrying-cross.mp4", filepath.Join(folder, "christ-carrying-cross.mp4")},
		{"david-goliath.mp4", filepath.Join(folder, "david-goliath.mp4")},
		{"self-portrait-titian.mp4", filepath.Join(folder, "self-portrait-titian.mp4")},
		{"the-cardinal.mp4", filepath.Join(folder, "the-cardinal.mp4")},
		{"virgin-of-the-chair.mp4", filepath.Join(folder, "virgin-of-the-chair.mp4")},
		{"three-graces.mp4", filepath.Join(folder, "three-graces.mp4")},
	}
}

func putArtworkFile(ctx context.Context, client *minio.Client, artworkKey, artworkPath string) {
	info, err := os.Stat(artworkPath)
	if err != nil {
		log.Printf("картина: нет файла %s: %v", artworkPath, err)
		return
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	contentType := mime.TypeByExtension(filepath.Ext(artworkPath))
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err = client.FPutObject(ctx, ArtworkBucket, artworkKey, artworkPath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Printf("картина: ключ %s не загружен: %v", artworkKey, err)
		return
	}
	log.Printf("картина: %s (%d байт)", artworkKey, info.Size())
}
