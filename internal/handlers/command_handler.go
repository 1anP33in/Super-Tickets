package handlers

import (
	"github.com/1anP33in/super-tickets/internal/utils"
	"github.com/bwmarrin/discordgo"
)

func RegisterCommands(s *discordgo.Session, appID, guildID string, definitions []*discordgo.ApplicationCommand) error {
	for _, command := range definitions {
		if _, err := s.ApplicationCommandCreate(appID, guildID, command); err != nil {
			return err
		}
		utils.Info("registered /%s", command.Name)
	}
	return nil
}
