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
//
// Every command lives under the c2- prefix (ADR 0095, #1631). The old
// cc- names are gone entirely — RegisterCommands bulk-overwrites each
// guild's command set, so a stale cc-* registration disappears on the
// first boot of this binary with no separate cleanup step.
const (
	CmdInvite    = "c2-invite"
	CmdGames     = "c2-games"
	CmdEnd       = "c2-end"
	CmdDeckCheck = "c2-deck-check"
	CmdDeckReq   = "c2-deck-req"
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
		{
			Name:        CmdEnd,
			Description: "Archive a cmd_and_ctrl game, after confirming (admin only).",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:         "game",
					Description:  "Game to archive — id or name; autocompletes over active games.",
					Type:         discordgo.ApplicationCommandOptionString,
					Required:     true,
					Autocomplete: true,
					MaxLength:    100,
				},
			},
		},
		{
			Name:        CmdDeckCheck,
			Description: "Check how much of a decklist the engine automates.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "link",
					Description: "Moxfield or Archidekt deck link.",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					MaxLength:   300,
				},
			},
		},
		{
			Name:        CmdDeckReq,
			Description: "Ask for a deck's missing cards to be added to the engine.",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Name:        "link",
					Description: "Moxfield or Archidekt deck link.",
					Type:        discordgo.ApplicationCommandOptionString,
					Required:    true,
					MaxLength:   300,
				},
			},
		},
	}
}

// commandRegistrar is the subset of *discordgo.Session that
// RegisterCommands needs. Defined as an interface — rather than
// taking *discordgo.Session directly — so tests can fake the bulk
// overwrite call without a live Discord connection or reaching into
// discordgo's package-level endpoint variables. *discordgo.Session
// satisfies this automatically.
type commandRegistrar interface {
	ApplicationCommandBulkOverwrite(appID string, guildID string, commands []*discordgo.ApplicationCommand, options ...discordgo.RequestOption) ([]*discordgo.ApplicationCommand, error)
}

// RegisterCommands overwrites the whole slash-command set on every
// guild in guildIDs with commandDefinitions(). Bulk overwrite
// (ApplicationCommandBulkOverwrite), not one ApplicationCommandCreate
// per command: Discord replaces the guild's entire command list in
// one call, so a command that existed under the old cc- names (or
// any command dropped from commandDefinitions since the last deploy)
// disappears the moment this runs — no separate delete step, no
// window where both the old and new names are registered at once.
// A failure on one guild is logged and skipped rather than aborting —
// a partial registration is better than no registration if an
// operator added the bot to a guild without the applications.commands
// scope.
func RegisterCommands(s commandRegistrar, appID string, guildIDs []string, log *slog.Logger) {
	defs := commandDefinitions()
	for _, gid := range guildIDs {
		if _, err := s.ApplicationCommandBulkOverwrite(appID, gid, defs); err != nil {
			log.Error("register commands", "guild", gid, "error", err.Error())
			continue
		}
		log.Info("registered commands", "guild", gid, "count", len(defs))
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
	// timestamp used when /c2-invite is called without an arg, and
	// the /c2-end confirmation-expiry clock.
	now func() time.Time
	// confirmations holds outstanding /c2-end confirm/cancel
	// buttons. In-memory only — same "no SIGHUP reload" trade-off
	// ADR 0004 already accepts for the guild allow-list; a bot
	// restart drops any confirmation mid-flight and the operator
	// just runs /c2-end again.
	confirmations *endConfirmations
	// deckLinks holds the full deck link behind a /c2-deck-check
	// "Request these cards" button whose URL doesn't fit in Discord's
	// 100-character custom-ID cap (ADR 0095 §4, #1631). Same in-memory,
	// no-SIGHUP-reload trade-off as confirmations above.
	deckLinks *deckLinkStore
}

// NewHandler wires the bot's configuration and HTTP client into
// a dispatcher suitable for session.AddHandler.
func NewHandler(cfg Config, client *ServerClient, log *slog.Logger) *Handler {
	return &Handler{
		cfg: cfg, client: client, log: log, now: time.Now,
		confirmations: newEndConfirmations(),
		deckLinks:     newDeckLinkStore(),
	}
}

// defaultInteractionTimeout caps the HTTP round-trip for every
// command except the two deck ones (see deckInteractionTimeout
// below). Discord's interaction-response deadline is 3s; the
// loopback admin-login + work call fits well under that on the VPS,
// but 4s gives a slow path a deterministic error rather than a
// discord-side "application did not respond".
const defaultInteractionTimeout = 4 * time.Second

// deckInteractionTimeout is the budget for /c2-deck-check,
// /c2-deck-req and the "Request these cards" button (ADR 0095 §4,
// #1631) — the first commands this bot defers. A deck fetch plus a
// GitHub file-or-comment call can comfortably exceed the 4s default,
// and deferring buys roughly 15 minutes; 20s is generous headroom
// without leaving a slash command "thinking" for an awkwardly long
// time on a slow deck host.
const deckInteractionTimeout = 20 * time.Second

// commandTimeout picks the interaction budget for a slash-command
// name. Only the two deck commands get the longer, deferred-reply
// budget; every other command keeps the tight, non-deferred one.
func commandTimeout(name string) time.Duration {
	switch name {
	case CmdDeckCheck, CmdDeckReq:
		return deckInteractionTimeout
	default:
		return defaultInteractionTimeout
	}
}

// componentTimeout picks the interaction budget for a message
// component click by its custom-ID namespace. Only the deck-request
// button — which runs the same server call /c2-deck-req does — gets
// the longer budget.
func componentTimeout(customID string) time.Duration {
	if strings.HasPrefix(customID, deckRequestCustomIDPrefix) {
		return deckInteractionTimeout
	}
	return defaultInteractionTimeout
}

// Dispatch is the entry point called on every InteractionCreate. It
// gates on the guild allow-list (defense-in-depth; the commands
// should not even be registered on non-allowed guilds) for every
// interaction type this bot handles, then routes to per-command,
// per-autocomplete or per-component handlers.
func (h *Handler) Dispatch(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand, discordgo.InteractionApplicationCommandAutocomplete, discordgo.InteractionMessageComponent:
		// handled below
	default:
		return
	}

	if !h.cfg.GuildAllowed(i.GuildID) {
		h.log.Warn("rejected interaction from non-allowed guild", "guild", i.GuildID, "type", i.Type.String())
		if i.Type == discordgo.InteractionApplicationCommandAutocomplete {
			_ = s.InteractionRespond(i.Interaction, autocompleteResponse(nil))
		} else {
			_ = s.InteractionRespond(i.Interaction, ephemeralResponse("This bot is not authorized for this server."))
		}
		return
	}

	switch i.Type {
	case discordgo.InteractionApplicationCommandAutocomplete:
		ctx, cancel := context.WithTimeout(context.Background(), defaultInteractionTimeout)
		defer cancel()
		h.dispatchAutocomplete(ctx, s, i)
		return
	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID
		ctx, cancel := context.WithTimeout(context.Background(), componentTimeout(customID))
		defer cancel()
		h.dispatchComponent(ctx, s, i, customID)
		return
	}

	data := i.ApplicationCommandData()
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout(data.Name))
	defer cancel()

	switch data.Name {
	case CmdInvite:
		h.handleInvite(ctx, s, i, data)
	case CmdGames:
		h.handleGames(ctx, s, i)
	case CmdEnd:
		h.handleEnd(ctx, s, i, data)
	case CmdDeckCheck:
		h.handleDeckCheck(ctx, s, i, data)
	case CmdDeckReq:
		h.handleDeckReq(ctx, s, i, data)
	default:
		h.log.Warn("unknown command", "name", data.Name)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("Unknown command."))
	}
}

// handleInvite runs /c2-invite: read the optional name,
// create a game, respond with a channel-visible invite link.
func (h *Handler) handleInvite(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	name, meta, err := h.createInviteGame(ctx, i, data)
	if err != nil {
		h.log.Error("create game failed", "error", err.Error(), "name", name)
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(inviteErrorMessage(err)))
		return
	}

	url := buildInviteURL(h.cfg.ClientBaseURL, meta.ID, meta.InviteToken)
	_ = s.InteractionRespond(i.Interaction, inviteSuccessResponse(meta, url))
}

// createInviteGame is /c2-invite's server call, split from the
// Discord response so it can be tested without a live session. The
// invoking Discord user is passed as host_discord_id: whoever runs
// /c2-invite hosts the table (ADR 0075 §2.1).
func (h *Handler) createInviteGame(ctx context.Context, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) (string, lobby.GameMeta, error) {
	name := stringOption(data.Options, "name")
	if strings.TrimSpace(name) == "" {
		name = defaultGameName(h.now())
	}
	meta, err := h.client.CreateGame(ctx, name, invokerID(i))
	return name, meta, err
}

// invokerID is the Discord user who ran the command: Member.User in a
// guild, User in a DM. Empty when neither is present.
func invokerID(i *discordgo.InteractionCreate) string {
	if i == nil || i.Interaction == nil {
		return ""
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID
	}
	if i.User != nil {
		return i.User.ID
	}
	return ""
}

// handleGames runs /c2-games: fetch the list (invite tokens
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

// defaultGameName is used when /c2-invite is called without a
// name arg. Timestamp lets the playgroup distinguish back-to-
// back games in the lobby list.
func defaultGameName(t time.Time) string {
	return "Discord game · " + t.UTC().Format("2006-01-02 15:04 UTC")
}

// stringOption reads a named string option from the interaction
// options slice. Returns "" if absent — both for "missing" and
// for "present but empty", which is what we want for the
// /c2-invite [name] optional-with-fallback shape.
func stringOption(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name && o.Type == discordgo.ApplicationCommandOptionString {
			return o.StringValue()
		}
	}
	return ""
}

// inviteSuccessResponse is the channel-visible embed posted on
// /c2-invite success. The whole playgroup needs to see the URL,
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

// gamesListResponse is the ephemeral /c2-games output. Empty
// list gets a friendly placeholder rather than an empty embed.
func gamesListResponse(games []lobby.GameMeta) *discordgo.InteractionResponse {
	if len(games) == 0 {
		return ephemeralResponse("No games right now. Run `/c2-invite` to start one.")
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
// Used for errors and the /c2-games output — both are
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
