package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func removeCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "remove",
			Description: "Remove a user or role from this ticket",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionMentionable, Name: "target", Description: "User or role", Required: true},
			},
		},
		RequiresStaff: true,
		SuccessTitle:  "Access removed",
		SuccessBody:   "The target can no longer access this ticket.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			targetID, _ := mentionable(i)
			return svc.RemoveAccess(s, i.ChannelID, targetID)
		},
	}
}
