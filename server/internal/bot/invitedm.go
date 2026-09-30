package bot

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

// /c2-invite-dm (#613, ADR 0051 decision 5, ADR 0004 amendment
// 2026-09-30).
//
// The command is a thin client of the server's
// POST /games/{id}/invites/dm. The server opens the DM itself with
// its own bot token, so there is exactly one place that builds and
// sends an invite DM; this bot never sends one from its gateway
// session. The route accepts a raw Discord snowflake (`discord_id`)
// only from an admin session, which is the credential the bot holds —
// that is why a mention is enough here and the target need not have
// signed in to the site.

// InviteDMResult mirrors the route's 200 body. It carries no
// snowflake and no invite token.
type InviteDMResult struct {
	Sent        bool   `json:"sent"`
	UserID      string `json:"user_id"`
	DisplayName string `json:"display_name"`
}

// InviteDMError is a non-200 answer from POST /games/{id}/invites/dm.
// StatusCode drives the user-visible wording (inviteDMErrorMessage).
type InviteDMError struct {
	StatusCode int
	Message    string
}

func (e *InviteDMError) Error() string {
	if e.Message == "" {
		return fmt.Sprintf("invite dm: server returned %d", e.StatusCode)
	}
	return fmt.Sprintf("invite dm: server returned %d: %s", e.StatusCode, e.Message)
}

// InviteDM calls POST /games/{id}/invites/dm with the cached admin
// session (re-logging in once on a 401), naming the target by Discord
// snowflake.
func (c *ServerClient) InviteDM(ctx context.Context, gameID uuid.UUID, discordID string) (InviteDMResult, error) {
	body, err := json.Marshal(map[string]string{"discord_id": discordID})
	if err != nil {
		return InviteDMResult{}, err
	}
	resp, err := c.doAuthorized(ctx, func(token string) (*http.Response, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/games/"+gameID.String()+"/invites/dm", bytes.NewReader(body))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := c.http.Do(req)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrServerUnreachable, err)
		}
		return resp, nil
	})
	if err != nil {
		return InviteDMResult{}, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusUnauthorized {
		return InviteDMResult{}, ErrUnauthorized
	}
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		var eb struct {
			Error string `json:"error"`
		}
		_ = json.Unmarshal(b, &eb)
		msg := eb.Error
		if msg == "" {
			msg = strings.TrimSpace(string(b))
		}
		return InviteDMResult{}, &InviteDMError{StatusCode: resp.StatusCode, Message: msg}
	}
	var out InviteDMResult
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return InviteDMResult{}, fmt.Errorf("decode invite dm: %w", err)
	}
	return out, nil
}

// inviteDMErrorMessage maps an InviteDM error to a player-facing line.
// 503 is the "not configured" answer the route gives while
// CMDCTRL_DISCORD_BOT_TOKEN or the public origin is missing; it never
// fails open, so this is the normal state on a preview host.
func inviteDMErrorMessage(err error) string {
	var de *InviteDMError
	if !errors.As(err, &de) {
		return inviteErrorMessage(err)
	}
	switch de.StatusCode {
	case http.StatusServiceUnavailable:
		return "Invite DMs are not set up on the game server (it needs `CMDCTRL_DISCORD_BOT_TOKEN` and a public base URL). Share the invite link in a channel instead."
	case http.StatusUnprocessableEntity:
		return "Discord would not deliver the DM. The person has to share a server with the bot and allow DMs from server members."
	case http.StatusTooManyRequests:
		return "Too many invite DMs just now. Wait a few seconds and try again."
	case http.StatusConflict:
		return "That table's invite link is no longer held in memory (the server restarted). Rotate the invite from the lobby, then try again."
	case http.StatusNotFound:
		return "The game server could not find that game or that Discord user."
	case http.StatusForbidden:
		return "The game server refused this DM."
	case http.StatusFailedDependency:
		return "Discord rejected the game server's request. Try again in a minute."
	case http.StatusBadRequest:
		return "The game server did not accept that request."
	default:
		return "Something went wrong sending the DM."
	}
}

// userOptionID reads a User-typed option's snowflake without needing
// a Session (discordgo's UserValue resolves through one).
func userOptionID(opts []*discordgo.ApplicationCommandInteractionDataOption, name string) string {
	for _, o := range opts {
		if o.Name == name && o.Type == discordgo.ApplicationCommandOptionUser {
			if id, ok := o.Value.(string); ok {
				return id
			}
		}
	}
	return ""
}

// handleInviteDM runs /c2-invite-dm. The work is deferred (two server
// calls plus the server's two Discord calls can exceed the 3s ack
// window) and the reply is ephemeral: the invite goes to the target's
// DMs, not the channel.
func (h *Handler) handleInviteDM(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	_ = s.InteractionRespond(i.Interaction, deferredEphemeralResponse())
	h.editEphemeral(s, i, h.runInviteDM(ctx, i, data), nil)
}

// runInviteDM is the whole command minus the Discord plumbing, so it
// can be tested: it returns the ephemeral text to show.
//
// Two shapes. With no `game`, it starts a new table exactly as
// /c2-invite does (the invoker hosts it) and DMs its invite. With a
// `game` (id, or a name prefix), it DMs that table's invite, but only
// for the game's creator or a bot admin — the same check as /c2-end.
// That gate is the bot's, not the server's: the server trusts the
// admin session the bot holds, so without it any guild member could
// make the bot DM anyone an invite to any table.
func (h *Handler) runInviteDM(ctx context.Context, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) string {
	target := userOptionID(data.Options, "user")
	if target == "" {
		return "Pick a user to DM."
	}
	if data.Resolved != nil && data.Resolved.Users[target] != nil && data.Resolved.Users[target].Bot {
		return "That is a bot; it can't play. Pick a person."
	}

	var (
		meta    lobby.GameMeta
		created bool
	)
	if raw := strings.TrimSpace(stringOption(data.Options, "game")); raw != "" {
		var err error
		meta, err = h.resolveGame(ctx, raw)
		if err != nil {
			h.log.Warn("c2-invite-dm resolve game failed", "input", raw, "error", err.Error())
			return resolveGameErrorMessage(err)
		}
		ok, err := h.mayEnd(ctx, meta.ID, i.Interaction)
		if err != nil {
			h.log.Error("c2-invite-dm permission check failed", "error", err.Error())
			return inviteErrorMessage(err)
		}
		if !ok {
			return "Only the table's creator or a bot admin can DM invites for an existing game. Leave `game` out to start a new one."
		}
	} else {
		var err error
		_, meta, err = h.createInviteGame(ctx, i, data)
		if err != nil {
			h.log.Error("c2-invite-dm create game failed", "error", err.Error())
			return inviteErrorMessage(err)
		}
		created = true
	}

	res, err := h.client.InviteDM(ctx, meta.ID, target)
	if err != nil {
		h.log.Warn("c2-invite-dm failed", "game", meta.ID.String(), "error", err.Error())
		msg := inviteDMErrorMessage(err)
		if created && meta.InviteToken != "" {
			// The table exists now; hand the invoker the link rather than
			// leaving an orphan game they can't reach. Ephemeral, so only
			// they see it.
			msg += fmt.Sprintf("\nThe game **%s** was created anyway: %s", meta.Name, buildInviteURL(h.cfg.ClientBaseURL, meta.ID, meta.InviteToken))
		}
		return msg
	}

	who := res.DisplayName
	if who == "" {
		who = "them"
	}
	if created {
		return fmt.Sprintf("Created **%s** and sent %s the invite by DM.", meta.Name, who)
	}
	return fmt.Sprintf("Sent %s the invite to **%s** by DM.", who, meta.Name)
}
