package bot

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

// Slash-command names. Exported as constants so tests can
// reference them without string literals that drift.
const (
	CmdInvite = "cc-invite"
	CmdGames  = "cc-games"
)

// commandDefinitions returns the ApplicationCommand payload
// registered with Discord. Guild-scoped registration means the
// commands appear instantly (global commands take up to an hour
// to propagate). Descriptions surface in the client's slash
// picker; keep them terse.
func commandDefinitions() []*discordgo.ApplicationCommand {
	return []*discordgo.ApplicationCommand{
		{
			Name:        CmdInvite,
			Description: "Create a new cmd_and_ctrl game and post the invite link.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "name",
					Description: "Optional display name for the game.",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    false,
					MaxLength:   80,
				},
			},
		},
		{
			Name:        CmdGames,
			Description: "List active and lobby games on cmd_and_ctrl.",
		},
	}
}

// RegisterCommands registers both slash commands on every guild
// in guildIDs. A failure on one guild is logged and skipped
// rather than aborting — a partial registration is better than
// no registration if an operator added the bot to a guild
// without the applications.commands scope.
func RegisterCommands(s *discordgo.Session, appID string, guildIDs []string, log *slog.Logger) {
	defs := commandDefinitions()
	for _, gid := range guildIDs {
		for _, cmd := range defs {
			if _, err := s.ApplicationCommandCreate(appID, gid, cmd); err != nil {
				log.Error("register command", "guild", gid, "command", cmd.Name, "error", err.Error())
				continue
			}
			log.Info("registered command", "guild", gid, "command", cmd.Name)
		}
	}
}

// Handler builds the InteractionCreate callback. Separated from
// NewHandler construction so tests can exercise the per-
// interaction logic without a live Session.
type Handler struct {
	cfg    Config
	client *ServerClient
	log    *slog.Logger
	// now is injected so tests can freeze the default-name
	// timestamp used when /cc-invite is called without an arg.
	now func() time.Time
}

// NewHandler wires the bot's configuration and HTTP client into
// a dispatcher suitable for session.AddHandler.
func NewHandler(cfg Config, client *ServerClient, log *slog.Logger) *Handler {
	return &Handler{cfg: cfg, client: client, log: log, now: time.Now}
}

// Dispatch is the entry point called on every InteractionCreate.
// It gates on the guild allow-list (defense-in-depth; the
// commands should not even be registered on non-allowed
// guilds), then routes to per-command handlers.
func (h *Handler) Dispatch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}
	data := i.ApplicationCommandData()

	if !h.cfg.GuildAllowed(i.GuildID) {
		h.log.Warn("rejected interaction from non-allowed guild", "guild", i.GuildID, "command", data.Name)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("This bot is not authorized for this server."))
		return
	}

	// Discord's interaction-response deadline is 3 s. The
	// loopback admin-login + work call fits well under that on
	// the VPS, but cap the HTTP round-trip at 4 s so a slow
	// path still produces a deterministic error rather than a
	// discord-side "application did not respond".
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	defer cancel()

	switch data.Name {
	case CmdInvite:
		h.handleInvite(ctx, s, i, data)
	case CmdGames:
		h.handleGames(ctx, s, i)
	default:
		h.log.Warn("unknown command", "name", data.Name)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("Unknown command."))
	}
}

// handleInvite runs /cc-invite: read the optional name,
// create a game, respond with a channel-visible invite link.
func (h *Handler) handleInvite(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	name := stringOption(data.Options, "name")
	if strings.TrimSpace(name) == "" {
		name = defaultGameName(h.now())
	}

	meta, err := h.client.CreateGame(ctx, name)
	if err != nil {
		h.log.Error("create game failed", "error", err.Error(), "name", name)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(inviteErrorMessage(err)))
		return
	}

	url := buildInviteURL(h.cfg.ClientBaseURL, meta.ID, meta.InviteToken)
	_ = s.InteractionRespond(i.Interaction, inviteSuccessResponse(meta, url))
}

// handleGames runs /cc-games: fetch the list (invite tokens
// already stripped by the server) and respond ephemerally —
// only the invoker needs to see the board; no need to spam
// the channel.
func (h *Handler) handleGames(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	games, err := h.client.ListGames(ctx)
	if err != nil {
		h.log.Error("list games failed", "error", err.Error())
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(inviteErrorMessage(err)))
		return
	}
	_ = s.InteractionRespond(i.Interaction, gamesListResponse(games))
}

// buildInviteURL composes the join URL the web client expects.
// Shape: {clientBase}/#/games/{uuid}/join?t={invite_token}
// clientBase is normalised by trimming a trailing slash so the
// result doesn't double up separators.
func buildInviteURL(clientBase string, gameID uuid.UUID, inviteToken string) string {
	return fmt.Sprintf("%s/#/games/%s/join?t=%s",
		strings.TrimRight(clientBase, "/"), gameID, inviteToken)
}

// defaultGameName is used when /cc-invite is called without a
// name arg. Timestamp lets the playgroup distinguish back-to-
// back games in the lobby list.
func defaultGameName(t time.Time) string {
	return "Discord game · " + t.UTC().Format("2006-01-02 15:04 UTC")
}

// stringOption reads a named string option from the interaction
// options slice. Returns "" if absent — both for "missing" and
// for "present but empty", which is what we want for the
// /cc-invite [name] optional-with-fallback shape.
func stringOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name && o.Type == discordgo.ApplicationCommandOptionString {
			return o.StringValue()
		}
	}
	return ""
}

// inviteSuccessResponse is the channel-visible embed posted on
// /cc-invite success. The whole playgroup needs to see the URL,
// so this response is NOT ephemeral.
func inviteSuccessResponse(meta lobby.GameMeta, url string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{{
				Title:       "Game created: " + meta.Name,
				Description: "Click to join:\n" + url,
				Color:       0x5865F2, // Discord blurple — on-brand accent
			}},
		},
	}
}

// gamesListResponse is the ephemeral /cc-games output. Empty
// list gets a friendly placeholder rather than an empty embed.
func gamesListResponse(games []lobby.GameMeta) *discordgo.InteractionResponse {
	if len(games) == 0 {
		return ephemeralResponse("No games right now. Run `/cc-invite` to start one.")
	}
	var b strings.Builder
	b.WriteString("**Current games:**\n")
	for _, g := range games {
		seats := len(g.Players)
		fmt.Fprintf(&b, "• `%s` — %s (%d seat", g.Name, g.State, seats)
		if seats != 1 {
			b.WriteString("s")
		}
		b.WriteString(")\n")
	}
	return ephemeralResponse(b.String())
}

// ephemeralResponse is shorthand for a flags-Ephemeral message.
// Used for errors and the /cc-games output — both are
// low-signal for the channel at large.
func ephemeralResponse(text string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: text,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}
}

// inviteErrorMessage maps a ServerClient error to a user-
// visible string. Keeps the interaction handlers free of
// switch-on-error clutter.
func inviteErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrServerUnreachable):
		return "Game server is not reachable right now."
	case errors.Is(err, ErrUnauthorized):
		return "Bot is not authorized against the game server — check CMDCTRL_ADMIN_TOKEN."
	default:
		return "Something went wrong talking to the game server."
	}
}
