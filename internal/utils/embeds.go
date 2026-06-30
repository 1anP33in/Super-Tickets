package utils

import "github.com/bwmarrin/discordgo"

const (
	ColorPrimary = 0x5865F2
	ColorSuccess = 0x3BA55D
	ColorWarning = 0xFAA61A
	ColorDanger  = 0xED4245
)

func Embed(title, description string, color int) *discordgo.MessageEmbed {
	return &discordgo.MessageEmbed{
		Title:       title,
		Description: description,
		Color:       color,
	}
}

func Success(title, description string) *discordgo.MessageEmbed {
	return Embed(title, description, ColorSuccess)
}

func Warning(title, description string) *discordgo.MessageEmbed {
	return Embed(title, description, ColorWarning)
}

func Danger(title, description string) *discordgo.MessageEmbed {
	return Embed(title, description, ColorDanger)
}
