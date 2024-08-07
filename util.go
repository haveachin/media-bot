package main

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/google/uuid"
)

func downloadVideo(url string) (*os.File, error) {
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
	fileReader := io.TeeReader(videoFile, fileContentBuf)
	hash, err := hash(fileReader)
	if err != nil {
		return "", err
	}

	slog.Debug("Created file hash",
		"hash", hash,
	)

	videoUID := uuid.New().String()
	videoSize := int64(fileContentBuf.Len())
	return storage.PutVideo(context.Background(), videoUID, fileContentBuf, videoSize)
}

func hash(r io.Reader) (uint64, error) {
	digest := xxhash.New()
	if _, err := io.Copy(digest, r); err != nil {
		return 0, err
	}
	return digest.Sum64(), nil
}
