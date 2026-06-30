# Super Tickets

Super Tickets is a Go-based Discord ticket bot with a built-in web dashboard.
It uses `discordgo`, SQLite persistence, Discord OAuth2, and vanilla HTML/CSS/JS.

## What It Does

- Registers slash commands for ticket creation and staff actions.
- Creates private ticket channels under a configured category.
- Uses configured support roles and department roles for ticket permissions.
- Supports blacklist, alert, claim, unclaim, transfer, rename, transcript, delete, add, remove, and close commands.
- Stores guild settings and ticket records in a real SQLite database.
- Provides a Discord OAuth dashboard for server managers.
- Loads roles and channels dynamically from the actual Discord server.
- Saves General Settings, Support Roles, Departments, and Blacklist through the API.

## Setup

1. Copy `.env.example` to `.env`.
2. Fill in your Discord application and bot values.
3. Set your Discord OAuth redirect URL to:

```text
http://localhost:8080/auth/discord/callback
```

4. Install dependencies:

```bash
go mod tidy
```

5. Run the app:

```bash
make run
```

The dashboard will be available at:

```text
http://localhost:8080
```

## Environment

```env
DISCORD_TOKEN=
CLIENT_ID=
CLIENT_SECRET=
REDIRECT_URI=http://localhost:8080/auth/discord/callback
SESSION_SECRET=change-me
DATABASE_URL=super_tickets.db
DASHBOARD_ADDR=:8080
GUILD_ID=
```

`GUILD_ID` is optional. If it is set, slash commands are registered only to that guild. If it is empty, commands are registered globally.

## Commands

- `/ticket`
- `/blacklist`
- `/alert`
- `/claim`
- `/unclaim`
- `/transfer`
- `/rename`
- `/transcript`
- `/delete`
- `/add`
- `/remove`
- `/close`

## Bot Permissions

The bot needs permission to manage channels, view channels, send messages, embed links, attach files, read message history, and manage permissions for ticket channels. Its role must be high enough to manage the channels and permission overwrites it creates.

## Project Layout

```text
cmd/bot/main.go                 App entrypoint, Discord client, dashboard API
Dashboard/                      Vanilla dashboard assets
internal/commands/              Slash command definitions and handlers
internal/config/                Environment loading
internal/database/              SQLite store and migrations
internal/events/                Discord event handlers
internal/handlers/              Command registration and ticket service
internal/models/                Shared data models
internal/utils/                 Embeds, permissions, logging
```
