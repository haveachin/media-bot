package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
)

func cmds() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        "video",
			Type:        discordgo.ChatApplicationCommand,
			Description: "Downloads a video and previews it in the channel",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "Video URL",
					Required:    true,
				},
			},
		},
	}
}

func syncCommands(s *discordgo.Session, guildID string, desiredCommands []*discordgo.ApplicationCommand) error {
	appID := s.State.Application.ID
	existingCommands, err := s.ApplicationCommands(appID, guildID)
	if err != nil {
		return err
	}

	existingMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range existingCommands {
		existingMap[cmd.Name] = cmd
	}

	desiredMap := make(map[string]*discordgo.ApplicationCommand)
	for _, cmd := range desiredCommands {
		desiredMap[cmd.Name] = cmd
	}

	for _, cmd := range existingCommands {
		if _, found := desiredMap[cmd.Name]; found {
			continue
		}

		err := s.ApplicationCommandDelete(appID, guildID, cmd.ID)
		if err != nil {
			slog.Error("Failed to delete command",
				slog.String("err", err.Error()),
			)
		}
	}

	for _, cmd := range desiredCommands {
		if existingCmd, found := existingMap[cmd.Name]; found {
			if _, err := s.ApplicationCommandEdit(appID, guildID, existingCmd.ID, cmd); err != nil {
				slog.Error("Failed to edit command",
					slog.String("err", err.Error()),
				)
			}
			continue
		}

		if _, err := s.ApplicationCommandCreate(appID, guildID, cmd); err != nil {
			slog.Error("Failed to create command",
				slog.String("err", err.Error()),
			)
		}
	}

	return nil
}

type ObjectStorage interface {
	PutVideo(ctx context.Context, name string, r io.Reader, size int64) (string, error)
}

type Database interface {
	InsertFileInfoRequest(
		ctx context.Context,
		username string,
		requestedURL string,
		fileSize int64,
		fileHash []byte,
	) (uuid.UUID, error)
	InsertFileInfoURL(
		ctx context.Context,
		fileInfoID uuid.UUID,
		url string,
	) error
	FileURLByHash(
		ctx context.Context,
		hash []byte,
	) (string, error)
}

func cmdHandler(db Database, objStorage ObjectStorage) func(s *discordgo.Session, i *discordgo.InteractionCreate) {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		cmdHandlers := map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate) error{
			"video": cmdVideoHandler(db, objStorage),
		}

		if h, ok := cmdHandlers[i.ApplicationCommandData().Name]; ok {
			if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
			}); err != nil {
				slog.Error("Failed to reponde to interaction",
					slog.String("err", err.Error()),
				)
				return
			}

			if err := h(s, i); err != nil {
				errMsg := err.Error()
				if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
					Content: &errMsg,
				}); err != nil {
					slog.Error("Failed to edit to interaction",
						slog.String("err", err.Error()),
					)
					return
				}
			}
		}
	}
}

func cmdVideoHandler(db Database, storage ObjectStorage) func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	return func(s *discordgo.Session, i *discordgo.InteractionCreate) error {
		args := i.ApplicationCommandData().Options
		if len(args) < 1 {
			return fmt.Errorf("invalid args count")
		}
		
		url := args[0].StringValue()
		videoFile, err := downloadVideo(url)
		if err != nil {
			return err
		}
		defer func() {
			videoFile.Close()
			os.Remove(videoFile.Name())
		}()

		info, err := videoFile.Stat()
		if err != nil {
			return err
		}

		fileIsBiggerThan25MB := info.Size() >= 26214400
		if fileIsBiggerThan25MB {
			videoURL, err := uploadVideo(storage, videoFile)
			if err != nil {
				return err
			}

			content := fmt.Sprintf("[Video](%s)", videoURL)
			if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
				Content: &content,
			}); err != nil {
				return err
			}
			return nil
		}

		if _, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
			Files: []*discordgo.File{
				{
					Name:   "video.mp4",
					Reader: videoFile,
				},
			},
		}); err != nil {
			return err
		}
		return nil
	}
}
