package main

import (
	"bytes"
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/cespare/xxhash/v2"
	"github.com/google/uuid"
)

func archiveVideo(storage Storage, url string) (string, error) {
	ytdlpCmd := exec.Command(
		"yt-dlp",
		"--format", "bestvideo*+bestaudio/best",
		"--print", "after_move:filepath",
		"--path", downloadDir,
		url,
	)

	ytdlpPipe, err := ytdlpCmd.StdoutPipe()
	if err != nil {
		return "", err
	}

	if err := ytdlpCmd.Start(); err != nil {
		return "", err
	}

	path, err := io.ReadAll(ytdlpPipe)
	if err != nil {
		return "", err
	}

	videoPath := strings.TrimSpace(string(path))
	file, err := os.Open(videoPath)
	if err != nil {
		return "", err
	}
	defer func() {
		file.Close()
		os.Remove(videoPath)
	}()

	fileContentBuf := new(bytes.Buffer)
	fileReader := io.TeeReader(file, fileContentBuf)
	hash, err := hash(fileReader)
	if err != nil {
		return "", err
	}

	log.Printf("File hash: %d", hash)

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
