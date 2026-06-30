package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func renameCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "rename",
			Description: "Rename the current ticket channel",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "name", Description: "New channel name", Required: true},
			},
		},
		RequiresStaff: true,
		SuccessTitle:  "Ticket renamed",
		SuccessBody:   "The channel name was updated.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			return svc.Rename(s, i.ChannelID, optionString(i, "name"))
		},
	}
}
