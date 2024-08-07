package main

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cespare/xxhash/v2"
)

func archiveVideo(storage Storage, url string) (string, error) {
	ytdlpCmd := exec.Command(
		"yt-dlp",
		"--format", "bestvideo*+bestaudio/best",
		"--print", "after_move:filepath",
		"--path", downloadDir,
		url,
	)

	ytdlpCmd.Stderr = os.Stderr

	ytdlpPipe, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	if err := ytdlpCmd.Start(); err != nil {
		return nil, err
	}

	path, err := io.ReadAll(ytdlpPipe)
	if err != nil {
		return nil, err
	}

	slog.Info("Saved video", "path", path)

	videoPath := strings.TrimSpace(string(path))
	return os.Open(videoPath)
}

func uploadVideo(storage Storage, videoFile *os.File) (string, error) {
	fileContentBuf := new(bytes.Buffer)
	fileReader := io.TeeReader(file, fileContentBuf)
	hash, err := hash(fileReader)
	if err != nil {
		return "", err
	}

	slog.Info("File hash: %d", hash)

	videoUID := uuid.New().String()
	videoSize := int64(fileContentBuf.Len())

	log.Printf("File hash: %#v", videoHash)

	ctx := context.Background()
	videoID, err := db.InsertFileInfoRequest(ctx, username, url, videoSize, videoHash)
	if err != nil {
		switch {
		case errors.Is(err, types.ErrFileHashAlreadyExists):
			return db.FileURLByHash(ctx, videoHash)
		default:
			return "", err
		}
	}

	videoURL, err := objStorage.PutVideo(ctx, videoID.String(), fileContentBuf, videoSize)
	if err != nil {
		return "", err
	}

	if err := db.InsertFileInfoURL(ctx, videoID, videoURL); err != nil {
		return "", err
	}

	return videoURL, nil
}

func hash(r io.Reader) ([]byte, error) {
	digest := xxhash.New()
	if _, err := io.Copy(digest, r); err != nil {
		return nil, err
	}

	hash := make([]byte, 0, 8)
	return digest.Sum(hash), nil
}
