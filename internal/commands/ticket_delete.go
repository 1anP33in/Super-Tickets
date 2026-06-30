package commands

import (
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func deleteCommand() Command {
	return Command{
		Definition:    &discordgo.ApplicationCommand{Name: "delete", Description: "Permanently delete this ticket channel"},
		RequiresStaff: true,
		SuccessTitle:  "Ticket deleting",
		SuccessBody:   "This channel will be deleted.",
		Execute: func(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			go func() {
				_ = svc.Delete(s, i.ChannelID)
			}()
			return nil
		},
	}
}
