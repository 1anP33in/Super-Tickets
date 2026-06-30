package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func alertCommand() Command {
	return Command{
		Definition:    &discordgo.ApplicationCommand{Name: "alert", Description: "Warn that this ticket may close due to inactivity"},
		RequiresStaff: true,
		SuccessTitle:  "Alert sent",
		SuccessBody:   "The inactivity warning was posted.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			return svc.Alert(s, i.ChannelID)
		},
	}
}
