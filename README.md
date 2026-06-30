# Super Tickets — Bot Scaffold (Go)

Folder structure for the Discord bot side of Super Tickets, written in Go
(idiomatic `cmd/` + `internal/` layout, built around `discordgo`). Every file
is an empty stub — no bot logic is implemented here, ready for you to fill in.

## Layout

```
bot/
├── go.mod / go.sum            # Module def + deps (discordgo, db driver, etc.)
├── Makefile                   # make run / make build / make tidy
├── .env.example                 # DISCORD_TOKEN, CLIENT_ID, GUILD_ID, DATABASE_URL
│
├── cmd/bot/
│   └── main.go                  # Entry point — loads config, opens session, starts handlers
│
└── internal/
    ├── commands/                # Slash command definitions + handlers
    │   ├── register.go            # Builds the discordgo.ApplicationCommand list
    │   ├── ticket_setup.go        # /ticket-setup — roles & categories
    │   ├── ticket_close.go        # /ticket-close — closes + archives a ticket
    │   └── ticket_roles.go        # /ticket-roles — manages support team roles
    │
    ├── events/                  # discordgo event listener callbacks
    │   ├── ready.go                # Bot startup / presence
    │   ├── interaction_create.go   # Routes slash commands + component interactions
    │   └── ticket_create.go        # Fires when a new ticket channel is opened
    │
    ├── handlers/                # Wiring / orchestration logic
    │   ├── command_handler.go     # Registers commands with the Discord API
    │   ├── event_handler.go       # Attaches event callbacks to the session
    │   └── ticket_handler.go      # Core open/close/archive/transcript logic
    │
    ├── config/                  # Static + per-guild configuration
    │   ├── config.go               # Loads .env / bot-wide defaults
    │   └── categories.go           # Live/archive category ID mappings
    │
    ├── database/                 # Persistence layer
    │   ├── schema.sql              # Tables: guilds, tickets, roles, transcripts
    │   ├── migrations.go           # Migration runner
    │   └── db.go                   # DB connection setup (pgx/database-sql)
    │
    ├── models/                   # Shared structs
    │   ├── guild.go                # Guild config struct (roles, categories, naming format)
    │   └── ticket.go               # Ticket struct (status, owner, channel ID, timestamps)
    │
    └── utils/                    # Shared helpers
        ├── embeds.go                # Reusable discordgo.MessageEmbed builders
        ├── permissions.go           # Role/permission checks
        └── logger.go                 # Structured logging helper
```

## How this maps to the dashboard

The web dashboard (`index.html` / `app.js` / `style.css`) edits a `systemState`
object in the browser: roles, live/archive category IDs, and the ticket
naming format. On a real backend, "Apply Changes" would call an API endpoint
that writes into the `guilds` table (see `internal/database/schema.sql`),
which `internal/handlers/ticket_handler.go` reads from any time a ticket is
opened or closed.

## Next steps (when you're ready to write the bot)

1. `go mod tidy` inside `bot/` to pull down `discordgo` and a DB driver (e.g. `pgx` or `mattn/go-sqlite3`).
2. Fill in `cmd/bot/main.go` to load `.env`, open a `discordgo.Session`, and call into `handlers`.
3. Implement `internal/handlers/ticket_handler.go` as the single source of truth for ticket lifecycle.
4. Stand up a small HTTP API (e.g. with `net/http` or `chi`) so the dashboard's "Apply Changes" button can persist real data instead of the in-memory `systemState`.
