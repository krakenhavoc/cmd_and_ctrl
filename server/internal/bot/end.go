package bot

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/lobby"
)

// confirmTTL is how long a /cc-end confirmation stays valid. Chosen
// to be long enough to read the prompt and click, short enough that
// a forgotten prompt doesn't sit around as a standing "archive this
// table" button someone could stumble into.
const confirmTTL = 60 * time.Second

// ccEndCustomIDPrefix namespaces this command's button CustomIDs so
// dispatchComponent can ignore any other component this bot might
// grow later without guessing at its shape.
const ccEndCustomIDPrefix = "cc-end:"

// errGameAmbiguous is returned by matchGameByName when more than one
// active game's name starts with the given prefix. Bot-side only —
// the server has no such concept — so it isn't a ServerClient
// sentinel like ErrGameNotFound.
var errGameAmbiguous = errors.New("more than one game matches that name")

// pendingEnd is a live /cc-end confirmation awaiting a button press.
type pendingEnd struct {
	GameID    uuid.UUID
	GameName  string
	InvokerID string
	ExpiresAt time.Time
}

// endConfirmations holds outstanding /cc-end confirmations, keyed by
// a random token embedded in the button CustomIDs. A token, not the
// message ID, is the correlation key because InteractionRespond
// doesn't hand the created message back to the caller.
type endConfirmations struct {
	mu      sync.Mutex
	pending map[string]pendingEnd
}

func newEndConfirmations() *endConfirmations {
	return &endConfirmations{pending: make(map[string]pendingEnd)}
}

// put stores a new pending confirmation. It also sweeps every
// already-expired entry it finds along the way — cheap, and keeps
// the map from growing forever across a long-running bot process
// when a confirmation is shown but never clicked.
func (e *endConfirmations) put(token string, p pendingEnd, now time.Time) {
	e.mu.Lock()
	defer e.mu.Unlock()
	for k, v := range e.pending {
		if now.After(v.ExpiresAt) {
			delete(e.pending, k)
		}
	}
	e.pending[token] = p
}

func (e *endConfirmations) get(token string) (pendingEnd, bool) {
	e.mu.Lock()
	defer e.mu.Unlock()
	p, ok := e.pending[token]
	return p, ok
}

func (e *endConfirmations) delete(token string) {
	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.pending, token)
}

// randomToken returns a URL-safe token for a /cc-end button CustomID.
// Not a security boundary by itself (the invoker check is), just
// enough entropy that two concurrent /cc-end calls never collide.
func randomToken() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read failing means the OS RNG is broken; fall
		// back to a timestamp so /cc-end degrades to "usually fine"
		// rather than panicking a slash command.
		return fmt.Sprintf("fallback-%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b)
}

// interactionInvoker returns the Discord user ID and role IDs of
// whoever triggered an interaction. Guild interactions carry Member
// (with Roles); DM interactions carry User instead and have no
// roles. This bot is guild-scoped (ADR 0004 §4), so the DM branch is
// defensive rather than load-bearing.
func interactionInvoker(i *discordgo.Interaction) (userID string, roles []string) {
	if i == nil {
		return "", nil
	}
	if i.Member != nil && i.Member.User != nil {
		return i.Member.User.ID, i.Member.Roles
	}
	if i.User != nil {
		return i.User.ID, nil
	}
	return "", nil
}

// handleEnd runs /cc-end: authorize, resolve the game, and (if it's
// not already archived) show a Confirm/Cancel prompt.
//
// Authorization is "host or admin" (#1098, closing out #1044 and
// #614): a configured admin per Config.IsAdmin, or the Discord user
// who created the table, per the server's GET /games/{id}/creator.
// The creator check is necessarily PER-GAME, so resolveGame now runs
// before authorization rather than after it — an unauthorized caller
// costs one HTTP call it didn't before. That is not a new leak: the
// same active/lobby listing this resolves against is already public
// to anyone in the guild via /cc-games and this command's own
// autocomplete (see dispatchAutocomplete), so nothing is learned here
// that a rejected caller couldn't already see.
func (h *Handler) handleEnd(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, data discordgo.ApplicationCommandInteractionData) {
	raw := stringOption(data.Options, "game")
	meta, err := h.resolveGame(ctx, raw)
	if err != nil {
		h.log.Warn("cc-end resolve game failed", "input", raw, "error", err.Error())
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(resolveGameErrorMessage(err)))
		return
	}

	allowed, err := h.mayEnd(ctx, meta.ID, i.Interaction)
	if err != nil {
		h.log.Warn("cc-end host check failed", "game_id", meta.ID.String(), "error", err.Error())
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(resolveGameErrorMessage(err)))
		return
	}
	if !allowed {
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(adminRefusalMessage(h.cfg)))
		return
	}

	if meta.Archived() {
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse(fmt.Sprintf(
			"**%s** is already archived — nothing to do. Undo with an admin `DELETE /games/%s/archive` if that wasn't intended.",
			meta.Name, meta.ID,
		)))
		return
	}

	uid, _ := interactionInvoker(i.Interaction)
	token := randomToken()
	now := h.now()
	h.confirmations.put(token, pendingEnd{
		GameID:    meta.ID,
		GameName:  meta.Name,
		InvokerID: uid,
		ExpiresAt: now.Add(confirmTTL),
	}, now)

	_ = s.InteractionRespond(i.Interaction, endConfirmResponse(meta, token))
}

// mayEnd reports whether i's invoker may run /cc-end against game id:
// a configured admin (cheap, no HTTP call), or — only when they are
// not — the Discord user the server says created that table (#1098).
// A game with no creator (an admin session created it, or it was
// restored from a pre-ADR-0051 file import) has the server answer
// false for every Discord user, so the two allowlists remain the only
// route for those games; empty allowlists and no creator therefore
// still refuse everyone, the same "never fail open" rule #614
// established for the allowlists alone.
func (h *Handler) mayEnd(ctx context.Context, id uuid.UUID, i *discordgo.Interaction) (bool, error) {
	if h.cfg.IsAdmin(i) {
		return true, nil
	}
	uid, _ := interactionInvoker(i)
	if uid == "" {
		return false, nil
	}
	return h.client.IsCreator(ctx, id, uid)
}

// dispatchAutocomplete answers the /cc-end "game" option's
// autocomplete requests by reusing the /cc-games listing — active
// and lobby games only, same as the channel command, so a typed
// prefix suggests exactly what /cc-games would have shown. Not
// authorization-gated: the same information is already visible to
// anyone who can run /cc-games in this guild.
func (h *Handler) dispatchAutocomplete(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()
	if data.Name != CmdEnd {
		_ = s.InteractionRespond(i.Interaction, autocompleteResponse(nil))
		return
	}

	var focused string
	for _, opt := range data.Options {
		if opt.Focused {
			focused = opt.StringValue()
		}
	}

	games, err := h.client.ListGames(ctx)
	if err != nil {
		h.log.Warn("cc-end autocomplete list failed", "error", err.Error())
		_ = s.InteractionRespond(i.Interaction, autocompleteResponse(nil))
		return
	}
	_ = s.InteractionRespond(i.Interaction, autocompleteResponse(gameChoices(games, focused)))
}

// dispatchComponent handles a Confirm/Cancel button click. Any
// CustomID this bot doesn't recognize is ignored — discordgo fans
// InteractionMessageComponent events out to every registered
// handler, and this bot may grow other components later.
func (h *Handler) dispatchComponent(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	customID := i.MessageComponentData().CustomID
	if !strings.HasPrefix(customID, ccEndCustomIDPrefix) {
		return
	}
	clickerID, _ := interactionInvoker(i.Interaction)
	outcome, pending, token := h.evaluateComponentClick(customID, clickerID, h.now())

	switch outcome {
	case outcomeInvalidCustomID:
		h.log.Warn("malformed cc-end custom id", "custom_id", customID)
	case outcomeNotFound:
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("This confirmation is no longer available. Run `/cc-end` again."))
	case outcomeWrongUser:
		_ = s.InteractionRespond(i.Interaction, ephemeralResponse("Only the person who ran `/cc-end` can respond to this confirmation."))
	case outcomeExpired:
		h.confirmations.delete(token)
		_ = s.InteractionRespond(i.Interaction, updateMessageResponse(fmt.Sprintf(
			"This confirmation for **%s** expired after %ds. Run `/cc-end` again.", pending.GameName, int(confirmTTL.Seconds()),
		)))
	case outcomeCancel:
		h.confirmations.delete(token)
		_ = s.InteractionRespond(i.Interaction, updateMessageResponse(fmt.Sprintf("Cancelled — **%s** was not archived.", pending.GameName)))
	case outcomeConfirm:
		h.confirmations.delete(token)
		h.finishConfirm(ctx, s, i, pending)
	}
}

// finishConfirm calls the archive route and edits the confirmation
// message with the result. Split out of dispatchComponent so the
// network call sits in one small function.
func (h *Handler) finishConfirm(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate, pending pendingEnd) {
	meta, err := h.client.ArchiveGame(ctx, pending.GameID)
	if err != nil {
		h.log.Error("archive game failed", "error", err.Error(), "game_id", pending.GameID.String())
		_ = s.InteractionRespond(i.Interaction, updateMessageResponse(archiveErrorMessage(err)))
		return
	}
	_ = s.InteractionRespond(i.Interaction, updateMessageResponse(fmt.Sprintf(
		"Archived **%s**. It's off the active list; undo with an admin `DELETE /games/%s/archive`.",
		meta.Name, meta.ID,
	)))
}

// componentOutcome is what a button click resolves to, decided
// without touching the Session — kept separate from
// dispatchComponent so the authorization/expiry logic is testable
// without a live discordgo connection.
type componentOutcome int

const (
	outcomeInvalidCustomID componentOutcome = iota
	outcomeNotFound
	outcomeWrongUser
	outcomeExpired
	outcomeCancel
	outcomeConfirm
)

// evaluateComponentClick resolves one cc-end:<action>:<token>
// CustomID click into an outcome. It does NOT mutate
// h.confirmations — callers decide what to delete based on the
// outcome, so a wrong-user or malformed click leaves the real
// confirmation alone for the actual invoker to still use.
func (h *Handler) evaluateComponentClick(customID, clickerID string, now time.Time) (componentOutcome, pendingEnd, string) {
	rest := strings.TrimPrefix(customID, ccEndCustomIDPrefix)
	action, token, ok := strings.Cut(rest, ":")
	if !ok || token == "" {
		return outcomeInvalidCustomID, pendingEnd{}, ""
	}

	pending, found := h.confirmations.get(token)
	if !found {
		return outcomeNotFound, pendingEnd{}, token
	}
	if clickerID == "" || clickerID != pending.InvokerID {
		return outcomeWrongUser, pending, token
	}
	if now.After(pending.ExpiresAt) {
		return outcomeExpired, pending, token
	}
	switch action {
	case "confirm":
		return outcomeConfirm, pending, token
	case "cancel":
		return outcomeCancel, pending, token
	default:
		return outcomeInvalidCustomID, pending, token
	}
}

// resolveGame turns the /cc-end "game" option's value into a
// GameMeta. Autocomplete always hands back a game ID, but a manually
// typed value (autocomplete is a suggestion, not a constraint) may
// be a name instead, so a value that doesn't parse as a UUID is
// matched by name prefix against the active/lobby listing —
// deliberately the same fallback ADR 0004 already documents for
// /cc-invite-style flexibility.
//
// Resolving by ID goes through GetGame, not ListGames, specifically
// because GetGame still finds an archived game — needed so
// handleEnd can tell "already archived" apart from "no such game".
func (h *Handler) resolveGame(ctx context.Context, raw string) (lobby.GameMeta, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return lobby.GameMeta{}, ErrGameNotFound
	}
	if id, err := uuid.Parse(raw); err == nil {
		return h.client.GetGame(ctx, id)
	}
	games, err := h.client.ListGames(ctx)
	if err != nil {
		return lobby.GameMeta{}, err
	}
	return matchGameByName(games, raw)
}

// matchGameByName does a case-insensitive prefix match of name
// against games. Zero matches is ErrGameNotFound; more than one is
// errGameAmbiguous, since silently picking one would risk archiving
// the wrong table.
func matchGameByName(games []lobby.GameMeta, name string) (lobby.GameMeta, error) {
	lower := strings.ToLower(name)
	var match lobby.GameMeta
	found := 0
	for _, g := range games {
		if strings.HasPrefix(strings.ToLower(g.Name), lower) {
			match = g
			found++
			if found > 1 {
				return lobby.GameMeta{}, errGameAmbiguous
			}
		}
	}
	if found == 0 {
		return lobby.GameMeta{}, ErrGameNotFound
	}
	return match, nil
}

// adminRefusalMessage explains why /cc-end refused the caller. An
// empty configuration gets a different message than "you specifically
// aren't on the list" — an operator seeing this for the first time
// should learn immediately which env vars to set, per issue #614.
func adminRefusalMessage(cfg Config) string {
	if !cfg.HasAdmins() {
		return "No admins are configured for /cc-end. Set CMDCTRL_DISCORD_ADMIN_USER_IDS and/or CMDCTRL_DISCORD_ADMIN_ROLE_IDS on the bot to allow it."
	}
	return "You're not authorized to run /cc-end. Ask an admin, or check CMDCTRL_DISCORD_ADMIN_USER_IDS / CMDCTRL_DISCORD_ADMIN_ROLE_IDS."
}

// resolveGameErrorMessage maps a resolveGame error to a user-visible
// string.
func resolveGameErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrGameNotFound):
		return "No game matches that. Try `/cc-games` for the active list, or paste the id."
	case errors.Is(err, errGameAmbiguous):
		return "More than one game name starts with that. Try a longer prefix, or paste the id from `/cc-games`."
	case errors.Is(err, ErrServerUnreachable):
		return "Game server is not reachable right now."
	case errors.Is(err, ErrUnauthorized):
		return "Bot is not authorized against the game server — check CMDCTRL_ADMIN_TOKEN."
	default:
		return "Something went wrong talking to the game server."
	}
}

// archiveErrorMessage maps an ArchiveGame error to a user-visible
// string, distinct from resolveGameErrorMessage's wording because by
// this point the caller already saw the confirmation prompt and
// needs to know the archive itself didn't happen.
func archiveErrorMessage(err error) string {
	switch {
	case errors.Is(err, ErrGameNotFound):
		return "That game is gone now — nothing was archived."
	case errors.Is(err, ErrServerUnreachable):
		return "Game server is not reachable right now — nothing was archived."
	case errors.Is(err, ErrUnauthorized):
		return "Bot is not authorized against the game server — check CMDCTRL_ADMIN_TOKEN. Nothing was archived."
	default:
		return "Something went wrong archiving the game — check `/cc-games` before retrying."
	}
}

// gameChoices builds the /cc-end autocomplete choices from the
// active/lobby listing, filtered by a case-insensitive substring
// match on name so a partial word anywhere in the table name works,
// not just a prefix. Discord caps autocomplete results at 25 and
// choice names at 100 characters.
func gameChoices(games []lobby.GameMeta, query string) []*discordgo.ApplicationCommandOptionChoice {
	lower := strings.ToLower(strings.TrimSpace(query))
	var out []*discordgo.ApplicationCommandOptionChoice
	for _, g := range games {
		if lower != "" && !strings.Contains(strings.ToLower(g.Name), lower) {
			continue
		}
		seats := len(g.Players)
		plural := "s"
		if seats == 1 {
			plural = ""
		}
		label := fmt.Sprintf("%s (%s, %d seat%s)", g.Name, g.State, seats, plural)
		if len(label) > 100 {
			label = label[:100]
		}
		out = append(out, &discordgo.ApplicationCommandOptionChoice{Name: label, Value: g.ID.String()})
		if len(out) == 25 {
			break
		}
	}
	return out
}

// endConfirmResponse builds the ephemeral Confirm/Cancel prompt,
// naming the table, its seat count and when it was created, per
// issue #614's confirmation-step requirement.
func endConfirmResponse(meta lobby.GameMeta, token string) *discordgo.InteractionResponse {
	seats := len(meta.Players)
	plural := "s"
	if seats == 1 {
		plural = ""
	}
	content := fmt.Sprintf(
		"Archive **%s**? %d player%s, created %s.\nThis is reversible (`DELETE /games/%s/archive` undoes it). This prompt expires in %ds.",
		meta.Name, seats, plural, meta.CreatedAt.UTC().Format("2006-01-02 15:04 UTC"), meta.ID, int(confirmTTL.Seconds()),
	)
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
			Components: []discordgo.MessageComponent{
				discordgo.ActionsRow{
					Components: []discordgo.MessageComponent{
						discordgo.Button{
							Label:    "Confirm",
							Style:    discordgo.DangerButton,
							CustomID: ccEndCustomIDPrefix + "confirm:" + token,
						},
						discordgo.Button{
							Label:    "Cancel",
							Style:    discordgo.SecondaryButton,
							CustomID: ccEndCustomIDPrefix + "cancel:" + token,
						},
					},
				},
			},
		},
	}
}

// updateMessageResponse edits the message the clicked button was
// attached to (InteractionResponseUpdateMessage) and clears its
// components — the confirm/cancel buttons should not still be
// clickable once the interaction is resolved.
func updateMessageResponse(content string) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseUpdateMessage,
		Data: &discordgo.InteractionResponseData{
			Content:    content,
			Components: []discordgo.MessageComponent{},
		},
	}
}

// autocompleteResponse wraps choices in the autocomplete-result
// response type. A nil/empty slice is valid — Discord just shows no
// suggestions.
func autocompleteResponse(choices []*discordgo.ApplicationCommandOptionChoice) *discordgo.InteractionResponse {
	return &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{Choices: choices},
	}
}
