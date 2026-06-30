package events

import (
	"github.com/1anP33in/super-tickets/internal/database"
	"github.com/bwmarrin/discordgo"
)

func MessageCreate(_ *discordgo.Session, m *discordgo.MessageCreate, store *database.Store) {
	if m.Author == nil || m.Author.Bot || m.GuildID == "" {
		return
	}
	ticket, err := store.TicketByChannel(m.ChannelID)
	if err != nil || ticket.Closed {
		return
	}
	_ = store.UpdateTicket(ticket)
}
