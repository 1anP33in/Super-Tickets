package models

type Department struct {
	Name   string `json:"name"`
	RoleID string `json:"roleId"`
	Emoji  string `json:"emoji,omitempty"`
}

type GuildConfig struct {
	GuildID             string       `json:"guildId"`
	TicketCategoryID    string       `json:"ticketCategoryId"`
	LogChannelID        string       `json:"logChannelId"`
	TranscriptChannelID string       `json:"transcriptChannelId"`
	AutoCloseMinutes    int          `json:"autoCloseMinutes"`
	SupportRoles        []string     `json:"supportRoles"`
	Departments         []Department `json:"departments"`
	Blacklist           []string     `json:"blacklist"`
}

func DefaultGuildConfig(guildID string) GuildConfig {
	return GuildConfig{
		GuildID:          guildID,
		AutoCloseMinutes: 1440,
		SupportRoles:     []string{},
		Departments:      []Department{},
		Blacklist:        []string{},
	}
}
