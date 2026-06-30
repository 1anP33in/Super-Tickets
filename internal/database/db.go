package database

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/1anP33in/super-tickets/internal/models"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func Open(databaseURL string) (*Store, error) {
	db, err := sql.Open("sqlite", databaseURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)

	store := &Store{db: db}
	if err := store.Migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) GetGuildConfig(guildID string) (models.GuildConfig, error) {
	var raw string
	err := s.db.QueryRow(`SELECT config_json FROM guild_configs WHERE guild_id = ?`, guildID).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		cfg := models.DefaultGuildConfig(guildID)
		return cfg, s.SaveGuildConfig(cfg)
	}
	if err != nil {
		return models.GuildConfig{}, err
	}

	var cfg models.GuildConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return models.GuildConfig{}, err
	}
	if cfg.GuildID == "" {
		cfg.GuildID = guildID
	}
	if cfg.SupportRoles == nil {
		cfg.SupportRoles = []string{}
	}
	if cfg.Departments == nil {
		cfg.Departments = []models.Department{}
	}
	if cfg.Blacklist == nil {
		cfg.Blacklist = []string{}
	}
	return cfg, nil
}

func (s *Store) SaveGuildConfig(cfg models.GuildConfig) error {
	if cfg.AutoCloseMinutes <= 0 {
		cfg.AutoCloseMinutes = 1440
	}
	if cfg.SupportRoles == nil {
		cfg.SupportRoles = []string{}
	}
	if cfg.Departments == nil {
		cfg.Departments = []models.Department{}
	}
	if cfg.Blacklist == nil {
		cfg.Blacklist = []string{}
	}

	body, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	_, err = s.db.Exec(`
		INSERT INTO guild_configs (guild_id, config_json, updated_at)
		VALUES (?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(guild_id) DO UPDATE SET config_json = excluded.config_json, updated_at = CURRENT_TIMESTAMP
	`, cfg.GuildID, string(body))
	return err
}

func (s *Store) CreateTicket(ticket models.Ticket) error {
	if ticket.CreatedAt.IsZero() {
		ticket.CreatedAt = time.Now().UTC()
	}
	if ticket.LastActivity.IsZero() {
		ticket.LastActivity = ticket.CreatedAt
	}
	_, err := s.db.Exec(`
		INSERT INTO tickets (channel_id, guild_id, owner_id, claimed_by, department, created_at, last_activity, closed, close_reason)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, ticket.ChannelID, ticket.GuildID, ticket.OwnerID, ticket.ClaimedBy, ticket.Department, ticket.CreatedAt, ticket.LastActivity, ticket.Closed, ticket.CloseReason)
	return err
}

func (s *Store) TicketByChannel(channelID string) (models.Ticket, error) {
	row := s.db.QueryRow(`
		SELECT channel_id, guild_id, owner_id, claimed_by, department, created_at, last_activity, closed, close_reason
		FROM tickets WHERE channel_id = ?
	`, channelID)
	return scanTicket(row)
}

func (s *Store) UpdateTicket(ticket models.Ticket) error {
	_, err := s.db.Exec(`
		UPDATE tickets
		SET claimed_by = ?, department = ?, last_activity = ?, closed = ?, close_reason = ?
		WHERE channel_id = ?
	`, ticket.ClaimedBy, ticket.Department, time.Now().UTC(), ticket.Closed, ticket.CloseReason, ticket.ChannelID)
	return err
}

func (s *Store) OpenTicketsOlderThan(before time.Time) ([]models.Ticket, error) {
	rows, err := s.db.Query(`
		SELECT channel_id, guild_id, owner_id, claimed_by, department, created_at, last_activity, closed, close_reason
		FROM tickets WHERE closed = 0 AND last_activity < ?
	`, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []models.Ticket
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

type ticketScanner interface {
	Scan(dest ...any) error
}

func scanTicket(scanner ticketScanner) (models.Ticket, error) {
	var t models.Ticket
	var closed bool
	err := scanner.Scan(&t.ChannelID, &t.GuildID, &t.OwnerID, &t.ClaimedBy, &t.Department, &t.CreatedAt, &t.LastActivity, &closed, &t.CloseReason)
	t.Closed = closed
	return t, err
}
