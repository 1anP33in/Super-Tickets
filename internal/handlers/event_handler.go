package handlers

import (
	"github.com/1anP33in/super-tickets/internal/events"
	"github.com/bwmarrin/discordgo"
)

func AttachEvents(s *discordgo.Session, svc *TicketService) {
	s.AddHandler(events.Ready)
	s.AddHandler(func(session *discordgo.Session, m *discordgo.MessageCreate) {
		events.MessageCreate(session, m, svc.Store)
	})
}
