package commands

import (
	"fmt"
	"strings"

	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/1anP33in/super-tickets/internal/utils"
	"github.com/bwmarrin/discordgo"
)

type Command struct {
	Definition    *discordgo.ApplicationCommand
	RequiresStaff bool
	Execute       func(*discordgo.Session, *discordgo.InteractionCreate, *handlers.TicketService) error
	SuccessTitle  string
	SuccessBody   string
}

func registry() []Command {
	return []Command{
		ticketCommand(),
		blacklistCommand(),
		alertCommand(),
		claimCommand(),
		unclaimCommand(),
		transferCommand(),
		renameCommand(),
		transcriptCommand(),
		deleteCommand(),
		addCommand(),
		removeCommand(),
		closeCommand(),
	}
}

func Definitions() []*discordgo.ApplicationCommand {
	commands := registry()
	definitions := make([]*discordgo.ApplicationCommand, 0, len(commands))
	for _, command := range commands {
		definitions = append(definitions, command.Definition)
	}
	return definitions
}

func Handle(s *discordgo.Session, i *discordgo.InteractionCreate, svc *handlers.TicketService) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	name := i.ApplicationCommandData().Name
	for _, command := range registry() {
		if command.Definition.Name != name {
			continue
		}
		if command.RequiresStaff {
			if err := svc.EnsureStaff(s, i.GuildID, i.Member); err != nil {
				respond(s, i, err, "", "")
				return
			}
		}
		err := command.Execute(s, i, svc)
		respond(s, i, err, command.SuccessTitle, command.SuccessBody)
		return
	}
}

func respond(s *discordgo.Session, i *discordgo.InteractionCreate, err error, title, description string) {
	embed := utils.Success(title, description)
	flags := discordgo.MessageFlagsEphemeral
	if err != nil {
		embed = utils.Danger("Unable to complete action", err.Error())
	}
	if title == "" {
		embed.Title = "Done"
	}
	_ = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}, Flags: flags},
	})
}

func optionString(i *discordgo.InteractionCreate, name string) string {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name {
			return strings.TrimSpace(opt.StringValue())
		}
	}
	return ""
}

func optionUserID(i *discordgo.InteractionCreate, name string) string {
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name == name {
			return fmt.Sprint(opt.Value)
		}
	}
	return ""
}

func mentionable(i *discordgo.InteractionCreate) (string, discordgo.PermissionOverwriteType) {
	data := i.ApplicationCommandData()
	for _, opt := range i.ApplicationCommandData().Options {
		if opt.Name != "target" {
			continue
		}
		id := fmt.Sprint(opt.Value)
		if data.Resolved != nil {
			if _, ok := data.Resolved.Roles[id]; ok {
				return id, discordgo.PermissionOverwriteTypeRole
			}
		}
		return id, discordgo.PermissionOverwriteTypeMember
	}
	return "", discordgo.PermissionOverwriteTypeMember
}

func channelID(channel *discordgo.Channel) string {
	if channel == nil {
		return ""
	}
	return channel.ID
}
