package main

import (
	"log"
	"log/slog"
	"media-bot/storage/object"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

const envVarPrefix = "MEDIABOT_"

var (
	discordToken    string
	minioAccessKey  string
	minioSecretKey  string
	minioBucketName string
	minioEndpoint   string
	downloadDir     = "downloads/"
)

func envVarStr(key string, val *string) {
	if val == nil {
		panic("env var val is a nil ptr")
	}

	key = envVarPrefix + key
	envVal, ok := os.LookupEnv(key)
	if !ok {
		return
	}
	*val = envVal
}

func initEnvVars() {
	envVarStr("DISCORD_TOKEN", &discordToken)
	envVarStr("MINIO_ACCESS_KEY", &minioAccessKey)
	envVarStr("MINIO_SECRET_KEY", &minioSecretKey)
	envVarStr("MINIO_BUCKET_NAME", &minioBucketName)
	envVarStr("MINIO_ENDPOINT", &minioEndpoint)
}

func main() {
	initEnvVars()

	if err := run(); err != nil {
		slog.Error("Failed to run",
			slog.String("err", err.Error()),
		)
	}
}

func run() error {
	objStorage, err := object.New(minioEndpoint, minioAccessKey, minioSecretKey, minioBucketName, true)
	if err != nil {
		return err
	}

	videoURL, err := archiveVideo(objStorage, "https://fxtwitter.com/CandyRibons/status/1787557585765507434")
	if err != nil {
		return err
	}

	log.Println(videoURL)

	return nil

	sess, err := discordgo.New("Bot " + discordToken)
	if err != nil {
		return err
	}

	if err := sess.Open(); err != nil {
		return err
	}
	defer sess.Close()

	if err := syncCommands(sess, "", cmds()); err != nil {
		return err
	}

	sess.AddHandler(cmdHandler(objStorage))

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		slog.Info("Received exit signal",
			slog.String("signal", sig.String()),
		)
	}

	return nil
}
