package config

import (
	"bufio"
	"errors"
	"os"
	"strings"
)

type Config struct {
	DiscordToken  string
	ClientID      string
	ClientSecret  string
	RedirectURI   string
	SessionSecret string
	DatabaseURL   string
	DashboardAddr string
	GuildID       string
}

func Load() (Config, error) {
	loadDotEnv(".env")

	cfg := Config{
		DiscordToken:  strings.TrimSpace(os.Getenv("DISCORD_TOKEN")),
		ClientID:      firstEnv("CLIENT_ID", "DISCORD_CLIENT_ID"),
		ClientSecret:  strings.TrimSpace(os.Getenv("CLIENT_SECRET")),
		RedirectURI:   strings.TrimSpace(os.Getenv("REDIRECT_URI")),
		SessionSecret: strings.TrimSpace(os.Getenv("SESSION_SECRET")),
		DatabaseURL:   strings.TrimSpace(os.Getenv("DATABASE_URL")),
		DashboardAddr: strings.TrimSpace(os.Getenv("DASHBOARD_ADDR")),
		GuildID:       firstEnv("GUILD_ID", "DISCORD_GUILD_ID"),
	}

	if cfg.DatabaseURL == "" {
		cfg.DatabaseURL = "super_tickets.db"
	}
	if cfg.DashboardAddr == "" {
		cfg.DashboardAddr = ":8080"
	}
	if cfg.RedirectURI == "" {
		cfg.RedirectURI = "http://localhost:8080/auth/discord/callback"
	}
	if cfg.DiscordToken == "" || cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.SessionSecret == "" {
		return cfg, errors.New("DISCORD_TOKEN, CLIENT_ID, CLIENT_SECRET, and SESSION_SECRET are required")
	}
	return cfg, nil
}

func firstEnv(names ...string) string {
	for _, name := range names {
		if value := strings.TrimSpace(os.Getenv(name)); value != "" {
			return value
		}
	}
	return ""
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || !strings.Contains(line, "=") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
}
