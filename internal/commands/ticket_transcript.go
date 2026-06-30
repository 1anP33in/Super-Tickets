package commands

import (
	"os"

	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func transcriptCommand() Command {
	return Command{
		Definition:    &discordgo.ApplicationCommand{Name: "transcript", Description: "Generate an HTML transcript for this ticket"},
		RequiresStaff: true,
		SuccessTitle:  "Transcript generated",
		SuccessBody:   "Saved transcript and sent it to the transcript channel when configured.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			path, err := svc.Transcript(s, i.ChannelID)
			if err != nil {
				return err
			}
			cfg, cfgErr := svc.Store.GetGuildConfig(i.GuildID)
			if cfgErr == nil && cfg.TranscriptChannelID != "" {
				if file, openErr := os.Open(path); openErr == nil {
					_, _ = s.ChannelFileSend(cfg.TranscriptChannelID, path, file)
					_ = file.Close()
				}
			}
			return nil
		},
	}
}
