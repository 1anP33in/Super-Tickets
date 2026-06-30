package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func claimCommand() Command {
	return Command{
		Definition:    &discordgo.ApplicationCommand{Name: "claim", Description: "Claim the current ticket"},
		RequiresStaff: true,
		SuccessTitle:  "Ticket claimed",
		SuccessBody:   "You are now assigned to this ticket.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			return svc.Claim(s, i.ChannelID, i.Member.User.ID)
		},
	}
}
