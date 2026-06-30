package events

import (
	"github.com/1anP33in/super-tickets/internal/utils"
	"github.com/bwmarrin/discordgo"
)

func Ready(_ *discordgo.Session, ready *discordgo.Ready) {
	utils.Info("logged in as %s", ready.User.String())
}
