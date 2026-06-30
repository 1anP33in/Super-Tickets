package commands

import (
	"fmt"
	"slices"

	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/bwmarrin/discordgo"
)

func blacklistCommand() Command {
	return Command{
		Definition: &discordgo.ApplicationCommand{
			Name:        "blacklist",
			Description: "Ban or unban a user from creating tickets",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionString, Name: "action", Description: "Action", Required: true, Choices: []*discordgo.ApplicationCommandOptionChoice{{Name: "ban", Value: "ban"}, {Name: "unban", Value: "unban"}}},
				{Type: discordgo.ApplicationCommandOptionUser, Name: "user", Description: "User", Required: true},
			},
		},
		RequiresStaff: true,
		SuccessTitle:  "Blacklist updated",
		SuccessBody:   "The blacklist has been saved.",
		Execute: func(_ *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) error {
			cfg, err := svc.Store.GetGuildConfig(i.GuildID)
			if err != nil {
				return err
			}
			userID := optionUserID(i, "user")
			switch action := optionString(i, "action"); action {
			case "ban":
				if !slices.Contains(cfg.Blacklist, userID) {
					cfg.Blacklist = append(cfg.Blacklist, userID)
				}
			case "unban":
				cfg.Blacklist = slices.DeleteFunc(cfg.Blacklist, func(id string) bool { return id == userID })
			default:
				return fmt.Errorf("unsupported action: %s", action)
			}
			return svc.Store.SaveGuildConfig(cfg)
		},
	}
}
