package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func unclaimCommand() Command {
	return Command{
		Definition:    &discordgo.ApplicationCommand{Name: "unclaim", Description: "Release the current ticket"},
		RequiresStaff: true,
		SuccessTitle:  "Ticket unclaimed",
		SuccessBody:   "This ticket is back in the open queue.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			return svc.Unclaim(s, i.ChannelID)
		},
	}
}
