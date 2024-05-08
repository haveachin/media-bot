package object

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client     *minio.Client
	endpoint   string
	bucketName string
	useSSL     bool
}

func New(endpoint, accessKey, secretKey, buckitName string, useSSL bool) (*Storage, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	return &Storage{
		client:     cli,
		endpoint:   endpoint,
		bucketName: buckitName,
		useSSL:     useSSL,
	}, nil
}

func (s *Storage) PutVideo(ctx context.Context, name string, r io.Reader, size int64) (string, error) {
	_, err := s.client.PutObject(ctx, s.bucketName, name, r, size, minio.PutObjectOptions{
		ContentType: "video/mp4",
	})
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucketName, name)
	if s.useSSL {
		url = "https://" + url
	} else {
		url = "http://" + url
	}

	return url, nil
}
