package models

import "time"

type Ticket struct {
	ChannelID    string    `json:"channelId"`
	GuildID      string    `json:"guildId"`
	OwnerID      string    `json:"ownerId"`
	ClaimedBy    string    `json:"claimedBy,omitempty"`
	Department   string    `json:"department,omitempty"`
	CreatedAt    time.Time `json:"createdAt"`
	LastActivity time.Time `json:"lastActivity"`
	Closed       bool      `json:"closed"`
	CloseReason  string    `json:"closeReason,omitempty"`
}
