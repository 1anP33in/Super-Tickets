package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func addCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "add",
			Description: "Add a user or role to this ticket",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionMentionable, Name: "target", Description: "User or role", Required: true},
			},
		},
		RequiresStaff: true,
		SuccessTitle:  "Access added",
		SuccessBody:   "The target can now access this ticket.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			targetID, targetType := mentionable(i)
			return svc.AddAccess(s, i.ChannelID, targetID, targetType)
		},
	}
}
