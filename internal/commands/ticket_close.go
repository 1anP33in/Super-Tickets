package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func closeCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "close",
			Description: "Close this ticket",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "reason", Description: "Reason", Required: false},
			},
		},
		RequiresStaff: true,
		SuccessTitle:  "Ticket closed",
		SuccessBody:   "The ticket was locked for the requester.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			return svc.Close(s, i.ChannelID, optionString(i, "reason"))
		},
	}
}
