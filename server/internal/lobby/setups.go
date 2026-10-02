package lobby

// setups.go is ADR 0110 section 5 items 1 and 2 (Delivery PR 7): a
// person's last table setup, captured when a table starts and applied
// to a new one.
//
//	GET  /me/setup          — signed in: the caller's remembered setup
//	POST /games/{id}/setup  — host, creator or admin: apply it to a lobby table
//	POST /games {"setup": "last"} — the same, at creation
//
// The setup is built by the server from the game it starts (table
// settings, bot seats, the other signed-in humans), never accepted from
// a client, and stored through internal/tablesetups (migration 0008).
// Applying one goes through the same UpdateSettings path as PATCH
// /games/{id}/settings and the same deck pipeline as POST
// /games/{id}/seats/bot, so a setup can never seat a bot or set a
// value those routes would refuse. What cannot be applied is skipped
// and named, never silently downgraded.

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/tablesetups"
)

// setupFromLast is the one setup source there is: the caller's last
// (ADR 0110 "Out of scope": no named setups).
const setupFromLast = "last"

// setupSkip names one part of a setup that was not applied, and why.
type setupSkip struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}

// setupResult says what applying a setup did.
type setupResult struct {
	// Settings reports whether the table settings were applied.
	Settings bool `json:"settings"`
	// BotsAdded counts the bot seats added.
	BotsAdded int `json:"bots_added"`
	// Skipped names everything that was not applied, with the reason.
	Skipped []setupSkip `json:"skipped"`
}

// applySetupRequest is the body of POST /games/{id}/setup.
type applySetupRequest struct {
	From string `json:"from"`
}

// applySetupResponse is the body of a successful POST
// /games/{id}/setup: the table after the setup, and what was skipped.
type applySetupResponse struct {
	Game GameMeta `json:"game"`
	setupResult
}

// setupView is a setup as GET /me/setup serves it. The settings are a
// complete SettingsPatch (wire names, as PATCH /games/{id}/settings
// takes them).
type setupView struct {
	Settings   json.RawMessage   `json:"settings"`
	Bots       []tablesetups.Bot `json:"bots"`
	Tablemates []uuid.UUID       `json:"tablemates"`
}

// mySetupResponse is the body of GET /me/setup. Setup is null when the
// caller has none yet.
type mySetupResponse struct {
	Setup *setupView `json:"setup"`
	// GameID is the table it was captured from.
	GameID *uuid.UUID `json:"game_id,omitempty"`
	// UpdatedAt is when it was captured, Unix milliseconds.
	UpdatedAt int64 `json:"updated_at,omitempty"`
}

func (c Config) tableSetups() tablesetups.Store {
	if c.TableSetups == nil {
		return tablesetups.NoStore{}
	}
	return c.TableSetups
}

// completeSettingsPatch is ts as a SettingsPatch with every field set,
// which is what a setup stores: applying it reproduces the table's
// settings whatever the new table's defaults are.
func completeSettingsPatch(ts game.TableSettings) game.SettingsPatch {
	return game.SettingsPatch{
		UndoLimit:       &ts.UndoLimit,
		UndoScope:       &ts.UndoScope,
		StartingLife:    &ts.StartingLife,
		CommanderDamage: &ts.CommanderDamage,
		BotPace:         &ts.BotPace,
		AllowSpawn:      &ts.AllowSpawn,
	}
}

// setupOwner is whose row a started table writes (ADR 0110 §5 item 1):
// the game's creator, if it has one, otherwise the person who pressed
// start, if they are a person. uuid.Nil means nobody's.
func setupOwner(meta GameMeta, starter auth.Principal) uuid.UUID {
	if meta.CreatedBy != uuid.Nil {
		return meta.CreatedBy
	}
	return starter.UserID
}

// buildSetup is the setup a started table leaves its owner: the table
// settings, every bot seat in seat order, and the other signed-in
// humans who sat there.
func buildSetup(meta GameMeta, settings game.TableSettings, owner uuid.UUID) (tablesetups.Setup, error) {
	raw, err := json.Marshal(completeSettingsPatch(settings))
	if err != nil {
		return tablesetups.Setup{}, err
	}
	setup := tablesetups.Setup{Settings: raw, Bots: []tablesetups.Bot{}, Tablemates: []uuid.UUID{}}
	seen := map[uuid.UUID]bool{owner: true}
	seats := append([]SeatInfo(nil), meta.Players...)
	sortSeats(seats)
	for _, s := range seats {
		if s.IsBot {
			setup.Bots = append(setup.Bots, tablesetups.Bot{Tier: s.BotTier, DeckID: s.BotDeck, Name: s.Name})
			continue
		}
		u := parseUUIDOrNil(s.UserID)
		if u == uuid.Nil || seen[u] {
			continue
		}
		seen[u] = true
		setup.Tablemates = append(setup.Tablemates, u)
	}
	return setup, nil
}

// sortSeats orders seats by seat number (insertion sort: four seats).
func sortSeats(s []SeatInfo) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j].Seat < s[j-1].Seat; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// captureSetup writes the owner's setup for a table that has just
// started. A failure is logged and never fails the start: the game is
// already running, and a setup is a convenience.
func captureSetup(ctx context.Context, c Config, starter auth.Principal, meta GameMeta) {
	if meta.Practice {
		return
	}
	owner := setupOwner(meta, starter)
	if owner == uuid.Nil {
		return
	}
	g, err := c.Lobby.LookupGame(meta.ID)
	if err != nil {
		return
	}
	setup, err := buildSetup(meta, g.TableSettingsSnapshot(), owner)
	if err == nil {
		err = c.tableSetups().Put(ctx, owner, meta.ID, setup)
	}
	if err != nil && !errors.Is(err, tablesetups.ErrNoStore) {
		c.logger().Warn("tablesetups: capturing a setup failed", "err", err, "game_id", meta.ID, "user_id", owner)
	}
}

// mySetup is GET /me/setup: the caller's remembered setup, for the
// create form's "use my last setup". Same caller rule as the rest of
// /me/*: a signed-in person, else 403 (never 401, #1154).
func mySetup(c Config, w http.ResponseWriter, r *http.Request) error {
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	rec, err := c.tableSetups().Get(r.Context(), p.UserID)
	if errors.Is(err, tablesetups.ErrNotFound) {
		return writeJSON(w, http.StatusOK, mySetupResponse{})
	}
	if err != nil {
		c.logger().Error("reading a user's table setup failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your last setup; try again")
	}
	out := mySetupResponse{Setup: &setupView{
		Settings:   rec.Setup.Settings,
		Bots:       nonNilBots(rec.Setup.Bots),
		Tablemates: nonNilUUIDs(rec.Setup.Tablemates),
	}}
	if len(out.Setup.Settings) == 0 {
		out.Setup.Settings = json.RawMessage(`{}`)
	}
	if rec.GameID != uuid.Nil {
		id := rec.GameID
		out.GameID = &id
	}
	if !rec.UpdatedAt.IsZero() {
		out.UpdatedAt = rec.UpdatedAt.UnixMilli()
	}
	return writeJSON(w, http.StatusOK, out)
}

func nonNilBots(b []tablesetups.Bot) []tablesetups.Bot {
	if b == nil {
		return []tablesetups.Bot{}
	}
	return b
}

func nonNilUUIDs(u []uuid.UUID) []uuid.UUID {
	if u == nil {
		return []uuid.UUID{}
	}
	return u
}

// canApplySetup is who may apply a setup to a table: whoever may
// manage it (CanManageTable: its host or an admin), or its creator.
// The creator is included because a setup is the creator's own and the
// table is theirs before they have sat down, when the first human seat
// would otherwise be its only manager (ADR 0075 §2.1).
func canApplySetup(c Config, p auth.Principal, meta GameMeta) bool {
	admin := c.isAdmin(p)
	return CanManageTable(p, meta, admin) || CanRotateInvites(p, meta, admin)
}

// applySetupRoute is POST /games/{id}/setup with {"from": "last"}:
// apply the caller's last setup to a lobby table. The caller must be a
// signed-in person (the setup is theirs; 403 otherwise) who may manage
// the table (canApplySetup, 403), the table must be unstarted (409),
// and they must have a setup (404).
func applySetupRoute(c Config, w http.ResponseWriter, r *http.Request) error {
	id, err := gameIDFromPath(r)
	if err != nil {
		return err
	}
	p, err := signedInUser(r)
	if err != nil {
		return err
	}
	var body applySetupRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	if strings.TrimSpace(body.From) != setupFromLast {
		return httpError(http.StatusBadRequest, fmt.Sprintf("from must be %q", setupFromLast))
	}
	meta, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	if !canApplySetup(c, p, meta) {
		return httpError(http.StatusForbidden, "only the table's host, its creator or an admin may apply a setup")
	}
	if meta.State != string(game.StateLobby) {
		return ErrGameStarted
	}
	rec, err := c.tableSetups().Get(r.Context(), p.UserID)
	if errors.Is(err, tablesetups.ErrNotFound) {
		return httpError(http.StatusNotFound, "you have no saved setup yet; one is saved when a table you created starts")
	}
	if err != nil {
		c.logger().Error("reading a user's table setup failed", "err", err)
		return httpError(http.StatusInternalServerError, "could not load your last setup; try again")
	}
	res := applySetup(r.Context(), c, p, id, rec.Setup)
	fresh, err := c.Lobby.Get(id)
	if err != nil {
		return err
	}
	return writeJSON(w, http.StatusOK, applySetupResponse{
		Game:        redactMetaFor(p, c.isAdmin(p), id, fresh),
		setupResult: res,
	})
}

// applyLastSetup is POST /games's "setup": "last". The table was just
// created by p, so there is nothing to authorise; a caller with no
// setup gets a result that says so.
func applyLastSetup(ctx context.Context, c Config, p auth.Principal, id uuid.UUID) setupResult {
	if p.UserID == uuid.Nil {
		return setupResult{Skipped: []setupSkip{{Name: "setup", Reason: "only a signed-in person has a saved setup"}}}
	}
	rec, err := c.tableSetups().Get(ctx, p.UserID)
	if errors.Is(err, tablesetups.ErrNotFound) {
		return setupResult{Skipped: []setupSkip{{Name: "setup", Reason: "you have no saved setup yet"}}}
	}
	if err != nil {
		c.logger().Error("reading a user's table setup failed", "err", err)
		return setupResult{Skipped: []setupSkip{{Name: "setup", Reason: "your last setup could not be loaded"}}}
	}
	return applySetup(ctx, c, p, id, rec.Setup)
}

// applySetup applies setup to the lobby table id: the settings through
// UpdateSettings, then each bot through the bot-seat pipeline, up to
// the free seats. Anything that cannot be applied is skipped and
// named.
func applySetup(ctx context.Context, c Config, p auth.Principal, id uuid.UUID, setup tablesetups.Setup) setupResult {
	res := setupResult{Skipped: []setupSkip{}}

	if len(bytes.TrimSpace(setup.Settings)) > 0 && !bytes.Equal(bytes.TrimSpace(setup.Settings), []byte("{}")) {
		var patch game.SettingsPatch
		switch err := json.Unmarshal(setup.Settings, &patch); {
		case err != nil:
			res.Skipped = append(res.Skipped, setupSkip{Name: "table settings", Reason: "the saved settings could not be read"})
		default:
			if verr := patch.Validate(); verr != nil {
				res.Skipped = append(res.Skipped, setupSkip{Name: "table settings", Reason: verr.Error()})
			} else if _, uerr := c.Lobby.UpdateSettings(id, actorIn(p, id), patch); uerr != nil {
				res.Skipped = append(res.Skipped, setupSkip{Name: "table settings", Reason: uerr.Error()})
			} else {
				res.Settings = true
			}
		}
	}

	for i, b := range setup.Bots {
		name := strings.TrimSpace(b.Name)
		if name == "" {
			name = fmt.Sprintf("Bot %d", i+1)
		}
		if reason := seatSetupBot(ctx, c, id, name, b); reason != "" {
			res.Skipped = append(res.Skipped, setupSkip{Name: name, Reason: reason})
			continue
		}
		res.BotsAdded++
	}
	return res
}

// seatSetupBot adds one bot from a setup, through the same checks POST
// /games/{id}/seats/bot makes. It returns "" on success, or the reason
// the bot was skipped.
func seatSetupBot(ctx context.Context, c Config, id uuid.UUID, name string, b tablesetups.Bot) string {
	if c.Bots == nil {
		return "bot seats are not enabled on this server"
	}
	tier := strings.ToLower(strings.TrimSpace(b.Tier))
	known := false
	for _, t := range c.Bots.Tiers() {
		if t == tier {
			known = true
			break
		}
	}
	if !known {
		return fmt.Sprintf("the %q tier is not available on this server", b.Tier)
	}
	deckID := strings.TrimSpace(b.DeckID)
	if deckID == "" {
		return "it played a pasted decklist, which a setup does not keep"
	}
	if c.BotDecks == nil {
		return "no bot deck catalog is configured on this server"
	}
	_, list, ok := c.BotDecks.Decklist(deckID)
	if !ok {
		return fmt.Sprintf("its deck %q no longer exists", deckID)
	}
	if c.Cards == nil || c.Cards.Count() == 0 {
		return "the card index is not loaded"
	}
	var sink responseSink
	resolved, _, written, err := resolveDeckSource(ctx, c, &sink, "text", list)
	if written {
		return "its deck " + sink.errorMessage()
	}
	if err != nil {
		return "its deck could not be loaded: " + err.Error()
	}
	if _, _, err := c.Lobby.AddBot(id, name, tier, deckID, resolved.Name, resolved.ToGameCards()); err != nil {
		if errors.Is(err, ErrGameFull) {
			return "the table has no free seat"
		}
		return err.Error()
	}
	return ""
}

// responseSink is an http.ResponseWriter that keeps what was written,
// so a setup can reuse resolveDeckSource (which answers a refused deck
// itself) and report the refusal as a skip instead.
type responseSink struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (s *responseSink) Header() http.Header {
	if s.header == nil {
		s.header = http.Header{}
	}
	return s.header
}

func (s *responseSink) Write(b []byte) (int, error) { return s.body.Write(b) }

func (s *responseSink) WriteHeader(code int) { s.status = code }

// errorMessage is the "error" field of what was written, lower-cased
// to read after "its deck".
func (s *responseSink) errorMessage() string {
	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(s.body.Bytes(), &body); err != nil || body.Error == "" {
		return "was refused"
	}
	return "was refused: " + body.Error
}
