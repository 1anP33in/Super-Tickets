package utils

import (
	"slices"

	"github.com/1anP33in/super-tickets/internal/models"
	"github.com/bwmarrin/discordgo"
)

func HasSupportAccess(member *discordgo.Member, cfg models.GuildConfig) bool {
	if member == nil {
		return false
	}
	if member.Permissions&discordgo.PermissionManageServer != 0 || member.Permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}
	for _, roleID := range member.Roles {
		if slices.Contains(cfg.SupportRoles, roleID) {
			return true
		}
		for _, dept := range cfg.Departments {
			if dept.RoleID == roleID {
				return true
			}
		}
	}
	return false
}

func IsBlacklisted(userID string, cfg models.GuildConfig) bool {
	return slices.Contains(cfg.Blacklist, userID)
}
