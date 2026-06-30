package handlers

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/1anP33in/super-tickets/internal/database"
	"github.com/1anP33in/super-tickets/internal/models"
	"github.com/1anP33in/super-tickets/internal/utils"
	"github.com/bwmarrin/discordgo"
)

type TicketService struct {
	Store *database.Store
}

func NewTicketService(store *database.Store) *TicketService {
	return &TicketService{Store: store}
}

func (svc *TicketService) OpenTicket(s *discordgo.Session, guildID, ownerID, departmentName string) (*discordgo.Channel, error) {
	cfg, err := svc.Store.GetGuildConfig(guildID)
	if err != nil {
		return nil, err
	}
	if cfg.TicketCategoryID == "" {
		return nil, errors.New("ticket category is not configured")
	}
	if utils.IsBlacklisted(ownerID, cfg) {
		return nil, errors.New("this user is blacklisted from creating tickets")
	}

	owner, err := s.User(ownerID)
	if err != nil {
		return nil, err
	}

	channelName := fmt.Sprintf("ticket-%s", sanitizeChannelName(owner.Username))
	overwrites := []*discordgo.PermissionOverwrite{
		{
			ID:   guildID,
			Type: discordgo.PermissionOverwriteTypeRole,
			Deny: discordgo.PermissionViewChannel,
		},
		{
			ID:    ownerID,
			Type:  discordgo.PermissionOverwriteTypeMember,
			Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
		},
	}

	for _, roleID := range cfg.SupportRoles {
		overwrites = append(overwrites, staffOverwrite(roleID))
	}
	dept := findDepartment(cfg, departmentName)
	if dept.RoleID != "" {
		overwrites = append(overwrites, staffOverwrite(dept.RoleID))
	}

	channel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
		Name:                 channelName,
		Type:                 discordgo.ChannelTypeGuildText,
		ParentID:             cfg.TicketCategoryID,
		PermissionOverwrites: overwrites,
		Topic:                fmt.Sprintf("Ticket owner: %s", ownerID),
	})
	if err != nil {
		return nil, err
	}

	ticket := models.Ticket{
		ChannelID:    channel.ID,
		GuildID:      guildID,
		OwnerID:      ownerID,
		Department:   dept.Name,
		CreatedAt:    time.Now().UTC(),
		LastActivity: time.Now().UTC(),
	}
	if err := svc.Store.CreateTicket(ticket); err != nil {
		return nil, err
	}

	_, _ = s.ChannelMessageSendEmbed(channel.ID, utils.Success("Ticket created", fmt.Sprintf("<@%s>, a support member will be with you soon.", ownerID)))
	return channel, nil
}

func (svc *TicketService) Claim(s *discordgo.Session, channelID, staffID string) error {
	ticket, err := svc.Store.TicketByChannel(channelID)
	if err != nil {
		return err
	}
	ticket.ClaimedBy = staffID
	if err := svc.Store.UpdateTicket(ticket); err != nil {
		return err
	}
	_, err = s.ChannelMessageSendEmbed(channelID, utils.Success("Ticket claimed", fmt.Sprintf("<@%s> is now helping <@%s>.", staffID, ticket.OwnerID)))
	return err
}

func (svc *TicketService) Unclaim(s *discordgo.Session, channelID string) error {
	ticket, err := svc.Store.TicketByChannel(channelID)
	if err != nil {
		return err
	}
	ticket.ClaimedBy = ""
	if err := svc.Store.UpdateTicket(ticket); err != nil {
		return err
	}
	_, err = s.ChannelMessageSendEmbed(channelID, utils.Warning("Ticket unclaimed", "This ticket is back in the open support queue."))
	return err
}

func (svc *TicketService) Transfer(s *discordgo.Session, channelID, departmentName string) error {
	ticket, err := svc.Store.TicketByChannel(channelID)
	if err != nil {
		return err
	}
	cfg, err := svc.Store.GetGuildConfig(ticket.GuildID)
	if err != nil {
		return err
	}
	dept := findDepartment(cfg, departmentName)
	if dept.Name == "" {
		return errors.New("department was not found")
	}

	ticket.Department = dept.Name
	if err := svc.Store.UpdateTicket(ticket); err != nil {
		return err
	}
	if dept.RoleID != "" {
		if err := s.ChannelPermissionSet(channelID, dept.RoleID, discordgo.PermissionOverwriteTypeRole, discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionReadMessageHistory, 0); err != nil {
			return err
		}
	}
	_, err = s.ChannelMessageSendEmbed(channelID, utils.Success("Ticket transferred", fmt.Sprintf("Moved to %s department.", dept.Name)))
	return err
}

func (svc *TicketService) Rename(s *discordgo.Session, channelID, name string) error {
	_, err := s.ChannelEdit(channelID, &discordgo.ChannelEdit{Name: sanitizeChannelName(name)})
	return err
}

func (svc *TicketService) AddAccess(s *discordgo.Session, channelID, targetID string, overwriteType discordgo.PermissionOverwriteType) error {
	return s.ChannelPermissionSet(channelID, targetID, overwriteType, discordgo.PermissionViewChannel|discordgo.PermissionSendMessages|discordgo.PermissionReadMessageHistory, 0)
}

func (svc *TicketService) RemoveAccess(s *discordgo.Session, channelID, targetID string) error {
	return s.ChannelPermissionDelete(channelID, targetID)
}

func (svc *TicketService) Alert(s *discordgo.Session, channelID string) error {
	_, err := s.ChannelMessageSendEmbed(channelID, utils.Warning("Inactivity warning", "This ticket may be closed soon due to inactivity. Reply here if you still need help."))
	return err
}

func (svc *TicketService) Close(s *discordgo.Session, channelID, reason string) error {
	ticket, err := svc.Store.TicketByChannel(channelID)
	if err != nil {
		return err
	}
	ticket.Closed = true
	ticket.CloseReason = reason
	if err := svc.Store.UpdateTicket(ticket); err != nil {
		return err
	}
	_ = s.ChannelPermissionSet(channelID, ticket.OwnerID, discordgo.PermissionOverwriteTypeMember, discordgo.PermissionViewChannel|discordgo.PermissionReadMessageHistory, discordgo.PermissionSendMessages)
	_, err = s.ChannelMessageSendEmbed(channelID, utils.Warning("Ticket closed", fallback(reason, "No reason provided.")))
	return err
}

func (svc *TicketService) Delete(s *discordgo.Session, channelID string) error {
	_, err := s.ChannelDelete(channelID)
	return err
}

func (svc *TicketService) Transcript(s *discordgo.Session, channelID string) (string, error) {
	messages, err := fetchMessages(s, channelID)
	if err != nil {
		return "", err
	}
	var body bytes.Buffer
	body.WriteString("<!doctype html><html><head><meta charset=\"utf-8\"><title>Ticket Transcript</title>")
	body.WriteString("<style>body{font-family:Arial,sans-serif;background:#101216;color:#e7eaf0;padding:24px}.msg{border-bottom:1px solid #2b2f38;padding:12px 0}.meta{color:#9aa3b2;font-size:12px}.content{white-space:pre-wrap}</style></head><body>")
	body.WriteString("<h1>Ticket Transcript</h1>")
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		author := "Unknown"
		if msg.Author != nil {
			author = html.EscapeString(msg.Author.String())
		}
		body.WriteString("<div class=\"msg\"><div class=\"meta\">")
		body.WriteString(author)
		body.WriteString(" - ")
		body.WriteString(msg.Timestamp.Format(time.RFC3339))
		body.WriteString("</div><div class=\"content\">")
		body.WriteString(html.EscapeString(msg.Content))
		body.WriteString("</div></div>")
	}
	body.WriteString("</body></html>")

	if err := os.MkdirAll("transcripts", 0755); err != nil {
		return "", err
	}
	path := fmt.Sprintf("transcripts/%s.html", channelID)
	return path, os.WriteFile(path, body.Bytes(), 0644)
}

func (svc *TicketService) EnsureStaff(s *discordgo.Session, guildID string, member *discordgo.Member) error {
	cfg, err := svc.Store.GetGuildConfig(guildID)
	if err != nil {
		return err
	}
	if !utils.HasSupportAccess(member, cfg) {
		return errors.New("you do not have permission to manage tickets")
	}
	return nil
}

func (svc *TicketService) AutoCloseExpired(s *discordgo.Session) {
	candidates, err := svc.Store.OpenTicketsOlderThan(time.Now().UTC().Add(-1 * time.Minute))
	if err != nil {
		utils.Error("auto-close scan: %v", err)
		return
	}
	for _, ticket := range candidates {
		cfg, err := svc.Store.GetGuildConfig(ticket.GuildID)
		if err != nil || cfg.AutoCloseMinutes <= 0 {
			continue
		}
		if time.Since(ticket.LastActivity) < time.Duration(cfg.AutoCloseMinutes)*time.Minute {
			continue
		}
		ticket.Closed = true
		ticket.CloseReason = "Closed automatically due to inactivity."
		if err := svc.Store.UpdateTicket(ticket); err != nil {
			utils.Error("auto-close ticket %s: %v", ticket.ChannelID, err)
			continue
		}
		_ = s.ChannelPermissionSet(ticket.ChannelID, ticket.OwnerID, discordgo.PermissionOverwriteTypeMember, discordgo.PermissionViewChannel|discordgo.PermissionReadMessageHistory, discordgo.PermissionSendMessages)
		_, _ = s.ChannelMessageSendEmbed(ticket.ChannelID, utils.Warning("Ticket closed", ticket.CloseReason))
	}
}

func staffOverwrite(roleID string) *discordgo.PermissionOverwrite {
	return &discordgo.PermissionOverwrite{
		ID:    roleID,
		Type:  discordgo.PermissionOverwriteTypeRole,
		Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
	}
}

func findDepartment(cfg models.GuildConfig, name string) models.Department {
	for _, dept := range cfg.Departments {
		if strings.EqualFold(dept.Name, name) {
			return dept
		}
	}
	if len(cfg.Departments) > 0 && name == "" {
		return cfg.Departments[0]
	}
	return models.Department{Name: strings.TrimSpace(name)}
}

func fetchMessages(s *discordgo.Session, channelID string) ([]*discordgo.Message, error) {
	var all []*discordgo.Message
	before := ""
	for {
		batch, err := s.ChannelMessages(channelID, 100, before, "", "")
		if err != nil {
			return nil, err
		}
		if len(batch) == 0 {
			return all, nil
		}
		all = append(all, batch...)
		before = batch[len(batch)-1].ID
		if len(batch) < 100 {
			return all, nil
		}
	}
}

var channelNamePattern = regexp.MustCompile(`[^a-z0-9-]+`)

func sanitizeChannelName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	name = strings.ReplaceAll(name, " ", "-")
	name = channelNamePattern.ReplaceAllString(name, "")
	name = strings.Trim(name, "-")
	if name == "" {
		return "ticket"
	}
	if len(name) > 90 {
		return name[:90]
	}
	return name
}

func fallback(value, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return value
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}
