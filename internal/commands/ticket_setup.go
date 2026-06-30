package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func ticketCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "ticket",
			Description: "Create a support ticket",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "department", Description: "Department name", Required: false},
			},
		},
		SuccessTitle: "Ticket created",
		SuccessBody:  "Your ticket channel is ready.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			_, err := svc.OpenTicket(s, i.GuildID, i.Member.User.ID, optionString(i, "department"))
			return err
		},
	}
}
