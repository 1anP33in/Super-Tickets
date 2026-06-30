package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func transferCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "transfer",
			Description: "Transfer this ticket to another department",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "department", Description: "Department name", Required: true},
			},
		},
		RequiresStaff: true,
		SuccessTitle:  "Ticket transferred",
		SuccessBody:   "The ticket was moved to the selected department.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			return svc.Transfer(s, i.ChannelID, optionString(i, "department"))
		},
	}
}
