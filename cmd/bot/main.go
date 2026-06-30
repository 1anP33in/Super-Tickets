package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/1anP33in/super-tickets/internal/commands"
	"github.com/1anP33in/super-tickets/internal/config"
	"github.com/1anP33in/super-tickets/internal/database"
	"github.com/1anP33in/super-tickets/internal/handlers"
	"github.com/1anP33in/super-tickets/internal/models"
	"github.com/1anP33in/super-tickets/internal/utils"
	"github.com/bwmarrin/discordgo"
)

const manageGuildPermission = 0x20

type app struct {
	cfg      config.Config
	store    *database.Store
	discord  *discordgo.Session
	sessions *sessionStore
}

type dashboardSession struct {
	UserID    string       `json:"userId"`
	Username  string       `json:"username"`
	Avatar    string       `json:"avatar"`
	Guilds    []oauthGuild `json:"guilds"`
	ExpiresAt time.Time    `json:"expiresAt"`
}

type oauthGuild struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Owner       bool   `json:"owner"`
	Permissions string `json:"permissions"`
}

type discordUser struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
}

type sessionStore struct {
	mu       sync.RWMutex
	secret   []byte
	sessions map[string]dashboardSession
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		utils.Error("%v", err)
		os.Exit(1)
	}

	store, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		utils.Error("database: %v", err)
		os.Exit(1)
	}
	defer store.Close()

	discord, err := discordgo.New("Bot " + cfg.DiscordToken)
	if err != nil {
		utils.Error("discord: %v", err)
		os.Exit(1)
	}
	discord.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildMembers | discordgo.IntentsMessageContent

	tickets := handlers.NewTicketService(store)
	handlers.AttachEvents(discord, tickets)
	discord.AddHandler(func(session *discordgo.Session, i *discordgo.InteractionCreate) {
		commands.Handle(session, i, tickets)
	})

	if err := discord.Open(); err != nil {
		utils.Error("discord open: %v", err)
		os.Exit(1)
	}
	defer discord.Close()

	if err := handlers.RegisterCommands(discord, cfg.ClientID, cfg.GuildID, commands.Definitions()); err != nil {
		utils.Error("register commands: %v", err)
		os.Exit(1)
	}

	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			tickets.AutoCloseExpired(discord)
		}
	}()

	application := &app{
		cfg:      cfg,
		store:    store,
		discord:  discord,
		sessions: newSessionStore(cfg.SessionSecret),
	}

	server := &http.Server{
		Addr:              cfg.DashboardAddr,
		Handler:           application.routes(),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		utils.Info("dashboard listening on http://localhost%s", cfg.DashboardAddr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			utils.Error("dashboard server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(ctx)
}

func (a *app) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth/discord/login", a.login)
	mux.HandleFunc("/auth/discord/callback", a.callback)
	mux.HandleFunc("/auth/logout", a.logout)
	mux.HandleFunc("/api/me", a.requireSession(a.me))
	mux.HandleFunc("/api/guilds", a.requireSession(a.guilds))
	mux.HandleFunc("/api/config/", a.requireSession(a.configAPI))
	mux.HandleFunc("/api/roles/", a.requireSession(a.rolesAPI))
	mux.HandleFunc("/api/channels/", a.requireSession(a.channelsAPI))
	mux.Handle("/", http.FileServer(http.Dir("Dashboard")))
	return withSecurity(mux)
}

func (a *app) login(w http.ResponseWriter, r *http.Request) {
	state := randomID()
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", Value: a.sessions.sign(state), Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 300})

	params := url.Values{}
	params.Set("client_id", a.cfg.ClientID)
	params.Set("redirect_uri", a.cfg.RedirectURI)
	params.Set("response_type", "code")
	params.Set("scope", "identify guilds")
	params.Set("state", state)
	http.Redirect(w, r, "https://discord.com/api/oauth2/authorize?"+params.Encode(), http.StatusFound)
}

func (a *app) callback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	cookie, err := r.Cookie("oauth_state")
	if err != nil || !a.sessions.verifySigned(cookie.Value, state) {
		http.Error(w, "invalid OAuth state", http.StatusBadRequest)
		return
	}

	token, err := a.exchangeCode(r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "OAuth token exchange failed", http.StatusBadGateway)
		return
	}
	user, guilds, err := fetchDiscordIdentity(token.AccessToken)
	if err != nil {
		http.Error(w, "Discord identity fetch failed", http.StatusBadGateway)
		return
	}

	sessionID := a.sessions.create(dashboardSession{
		UserID:    user.ID,
		Username:  user.Username,
		Avatar:    user.Avatar,
		Guilds:    guilds,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	http.SetCookie(w, &http.Cookie{Name: "st_session", Value: sessionID, Path: "/", HttpOnly: true, SameSite: http.SameSiteLaxMode, MaxAge: 86400})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *app) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie("st_session"); err == nil {
		a.sessions.delete(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: "st_session", Value: "", Path: "/", MaxAge: -1})
	http.Redirect(w, r, "/", http.StatusFound)
}

func (a *app) me(w http.ResponseWriter, r *http.Request, session dashboardSession) {
	writeJSON(w, http.StatusOK, map[string]any{"userId": session.UserID, "username": session.Username, "avatar": session.Avatar})
}

func (a *app) guilds(w http.ResponseWriter, r *http.Request, session dashboardSession) {
	botGuilds := a.botGuildIDs()
	type responseGuild struct {
		ID         string `json:"id"`
		Name       string `json:"name"`
		Icon       string `json:"icon"`
		CanManage  bool   `json:"canManage"`
		BotPresent bool   `json:"botPresent"`
	}
	var response []responseGuild
	for _, guild := range session.Guilds {
		canManage := guild.Owner || hasManageGuild(guild.Permissions)
		if !canManage {
			continue
		}
		response = append(response, responseGuild{
			ID:         guild.ID,
			Name:       guild.Name,
			Icon:       guild.Icon,
			CanManage:  true,
			BotPresent: botGuilds[guild.ID],
		})
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *app) configAPI(w http.ResponseWriter, r *http.Request, session dashboardSession) {
	guildID := strings.TrimPrefix(r.URL.Path, "/api/config/")
	if !a.canManage(session, guildID) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	switch r.Method {
	case http.MethodGet:
		cfg, err := a.store.GetGuildConfig(guildID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	case http.MethodPost:
		var cfg models.GuildConfig
		if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}
		cfg.GuildID = guildID
		if cfg.AutoCloseMinutes <= 0 {
			http.Error(w, "autoCloseMinutes must be positive", http.StatusBadRequest)
			return
		}
		if err := a.store.SaveGuildConfig(cfg); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		writeJSON(w, http.StatusOK, cfg)
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (a *app) rolesAPI(w http.ResponseWriter, r *http.Request, session dashboardSession) {
	guildID := strings.TrimPrefix(r.URL.Path, "/api/roles/")
	if !a.canManage(session, guildID) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	roles, err := a.discord.GuildRoles(guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	writeJSON(w, http.StatusOK, roles)
}

func (a *app) channelsAPI(w http.ResponseWriter, r *http.Request, session dashboardSession) {
	guildID := strings.TrimPrefix(r.URL.Path, "/api/channels/")
	if !a.canManage(session, guildID) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}
	channels, err := a.discord.GuildChannels(guildID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	type channel struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type int    `json:"type"`
	}
	var response []channel
	for _, c := range channels {
		if c.Type == discordgo.ChannelTypeGuildText || c.Type == discordgo.ChannelTypeGuildCategory {
			response = append(response, channel{ID: c.ID, Name: c.Name, Type: int(c.Type)})
		}
	}
	writeJSON(w, http.StatusOK, response)
}

func (a *app) requireSession(next func(http.ResponseWriter, *http.Request, dashboardSession)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("st_session")
		if err != nil {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		session, ok := a.sessions.get(cookie.Value)
		if !ok {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r, session)
	}
}

func (a *app) exchangeCode(code string) (tokenResponse, error) {
	values := url.Values{}
	values.Set("client_id", a.cfg.ClientID)
	values.Set("client_secret", a.cfg.ClientSecret)
	values.Set("grant_type", "authorization_code")
	values.Set("code", code)
	values.Set("redirect_uri", a.cfg.RedirectURI)

	req, err := http.NewRequest(http.MethodPost, "https://discord.com/api/oauth2/token", strings.NewReader(values.Encode()))
	if err != nil {
		return tokenResponse{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return tokenResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return tokenResponse{}, fmt.Errorf("discord returned %s", resp.Status)
	}
	var token tokenResponse
	return token, json.NewDecoder(resp.Body).Decode(&token)
}

func fetchDiscordIdentity(accessToken string) (discordUser, []oauthGuild, error) {
	var user discordUser
	if err := discordGet(accessToken, "https://discord.com/api/users/@me", &user); err != nil {
		return user, nil, err
	}
	var guilds []oauthGuild
	if err := discordGet(accessToken, "https://discord.com/api/users/@me/guilds", &guilds); err != nil {
		return user, nil, err
	}
	return user, guilds, nil
}

func discordGet(accessToken, endpoint string, target any) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("discord returned %s: %s", resp.Status, body)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func (a *app) botGuildIDs() map[string]bool {
	ids := map[string]bool{}
	for _, guild := range a.discord.State.Guilds {
		ids[guild.ID] = true
	}
	return ids
}

func (a *app) canManage(session dashboardSession, guildID string) bool {
	if !a.botGuildIDs()[guildID] {
		return false
	}
	for _, guild := range session.Guilds {
		if guild.ID == guildID {
			return guild.Owner || hasManageGuild(guild.Permissions)
		}
	}
	return false
}

func hasManageGuild(permissionValue string) bool {
	value, err := strconv.ParseUint(permissionValue, 10, 64)
	if err != nil {
		return false
	}
	return value&manageGuildPermission != 0
}

func newSessionStore(secret string) *sessionStore {
	return &sessionStore{secret: []byte(secret), sessions: map[string]dashboardSession{}}
}

func (s *sessionStore) create(session dashboardSession) string {
	id := randomID()
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[id] = session
	return id
}

func (s *sessionStore) get(id string) (dashboardSession, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	if !ok || time.Now().After(session.ExpiresAt) {
		return dashboardSession{}, false
	}
	return session, true
}

func (s *sessionStore) delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, id)
}

func (s *sessionStore) sign(value string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString([]byte(value + "." + hex.EncodeToString(mac.Sum(nil))))
}

func (s *sessionStore) verifySigned(signed, expected string) bool {
	raw, err := base64.RawURLEncoding.DecodeString(signed)
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(raw), ".", 2)
	if len(parts) != 2 || parts[0] != expected {
		return false
	}
	return hmac.Equal([]byte(s.sign(parts[0])), []byte(signed))
}

func randomID() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b[:])
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func withSecurity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		next.ServeHTTP(w, r)
	})
}

func dashboardPath(name string) string {
	return filepath.Join("Dashboard", name)
}
