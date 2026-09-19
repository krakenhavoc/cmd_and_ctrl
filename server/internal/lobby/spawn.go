package lobby

// spawn.go is the PRODUCTION spawner (ADR 0075 §2.4), which amends
// ADR 0023's "spawning is a cheat on a live table".
//
// The dev spawner in dev.go is unchanged and still lives behind
// requireDevFeature: on a preview box anyone at the table is a
// tester. This route is the opposite posture in every respect. It is
// not environment-gated, so it needs two gates of its own — the
// caller must be able to manage the table (CanManageTable) AND the
// table must have opted in (Settings.AllowSpawn, off by default) —
// and because it runs at a real table it is ANNOUNCED in the game log
// and UNDOABLE, neither of which the dev route bothers with.
//
// The two gates are reported separately on refusal. "You are not the
// host" and "this table has spawning switched off" are different
// problems with different fixes, and a single "forbidden" would send
// the host to ask the admin for a permission they already have.

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/ws"
)

// ErrSpawnNotAllowed is returned when the table has not switched
// AllowSpawn on. Distinct from ErrNotTableManager on purpose — see
// the file comment.
var ErrSpawnNotAllowed = errors.New("lobby: this table has spawning switched off")

// TokenTemplates is the spawner's source of token templates. Satisfied
// by effects.Tokens(), wired into Config in main.
//
// An interface rather than a map so the lobby does not import
// internal/cards/effects: the catalog is blank-imported by main and
// imported by internal/catalog, and nothing else depends on it (see
// catalog.go's package comment). A nil Tokens is a supported
// configuration — the token half of the spawner is simply not offered,
// which is what a test that never wires it gets.
type TokenTemplates interface {
	// TokenKeys lists every spawnable template key, sorted.
	TokenKeys() []string
	// TokenTemplate returns a fresh copy of the named template.
	TokenTemplate(key string) (game.Card, bool)
}

// SpawnOptions is one spawn, as the lobby performs it.
type SpawnOptions struct {
	// Actor is who asked: a player ID, or uuid.Nil for the server
	// admin. Named in the log line and stamped on the undo entry.
	Actor uuid.UUID
	// Controller is the seat the cards belong to. Required even for
	// the shared zones, because control matters there.
	Controller uuid.UUID
	Zone       game.ZoneKind
	Template   game.Card
	Count      int
	// Managed marks the production route. It requires the table's
	// AllowSpawn setting and commits through the undo stack; false is
	// the dev route, which requires neither.
	Managed bool
}

// Spawn inserts Count copies of the template into a zone. Mirrors
// SetDeck: mutate under the room so seq bumps and the replay stream
// records it, then broadcast after releasing l.mu.
//
// Routing this through the room rather than poking Game directly is
// what makes a spawn behave like every other mutation — it lands in
// the replay, so a bug found with a spawned board is still
// reproducible from the recording.
//
// A MANAGED spawn commits as a ws.Bundle rather than an
// ApplyExternal, which is the only difference in the commit path and
// the whole of what makes it undoable: the bundle leaves an undo
// entry attributed to Actor. The entry is FreeUndo, so correcting a
// typo in a card name does not cost the host the one take-back they
// get per turn — the spawn was never a play, and charging for its
// reversal would make the host ration their own repairs.
//
// Requires an active game: spawning into a lobby-state game would be
// undone by Start dealing opening hands, which reads as the feature
// being broken rather than misused.
func (l *Lobby) Spawn(gameID uuid.UUID, opt SpawnOptions) ([]uuid.UUID, error) {
	// Registered before the lock defer so it runs after l.mu is
	// released — see applyLocked.
	var broadcast func()
	defer func() {
		if broadcast != nil {
			broadcast()
		}
	}()
	l.mu.Lock()
	defer l.mu.Unlock()

	entry, ok := l.games[gameID]
	if !ok {
		return nil, ErrGameNotFound
	}
	if entry.room.Game.CurrentState() != game.StateActive {
		return nil, ErrGameNotActiveForSpawn
	}
	if opt.Managed && !entry.room.Game.TableSettingsSnapshot().AllowSpawn {
		return nil, ErrSpawnNotAllowed
	}

	var ids []uuid.UUID
	mutate := func() error {
		var innerErr error
		ids, innerErr = entry.room.Game.SpawnCards(opt.Actor, opt.Controller, opt.Zone, opt.Template, opt.Count)
		return innerErr
	}

	var err error
	if opt.Managed {
		broadcast, err = l.applyBundleLocked(gameID, entry, ws.Bundle{
			Caller:   opt.Actor,
			FreeUndo: true,
			Steps:    []func() error{mutate},
		})
	} else {
		broadcast, err = l.applyLocked(gameID, entry, mutate)
	}
	if err != nil {
		return nil, err
	}
	return ids, nil
}

// SpawnCards is Spawn for the dev route's shape: no actor, no setting
// check, no undo entry. Kept so the develop environment's tooling and
// its tests read exactly as they did before ADR 0075.
func (l *Lobby) SpawnCards(gameID, playerID uuid.UUID, zone game.ZoneKind, template game.Card, n int) ([]uuid.UUID, error) {
	return l.Spawn(gameID, SpawnOptions{Controller: playerID, Zone: zone, Template: template, Count: n})
}

// applyBundleLocked is applyLocked for a commit that must leave an
// undo entry behind. Callers hold l.mu, and must run the returned
// thunk after releasing it — see applyLocked for why.
func (l *Lobby) applyBundleLocked(id uuid.UUID, entry *gameEntry, b ws.Bundle) (func(), error) {
	view, seq, err := entry.room.ApplyBundle(b)
	if err != nil {
		return nil, err
	}
	if l.broadcast == nil {
		return func() {}, nil
	}
	bc := l.broadcast
	return func() { bc.BroadcastState(id, seq, view) }, nil
}

// --- HTTP ---------------------------------------------------------

// spawnRequest is the body of POST /games/{id}/spawn.
//
// Exactly one of ScryfallID, Name or Token identifies what to make.
// The first two are the dev spawner's card identifiers and behave
// identically; Token names a template from GET
// /games/{id}/spawn/tokens.
type spawnRequest struct {
	ScryfallID string `json:"scryfall_id,omitempty"`
	Name       string `json:"name,omitempty"`
	// Token is a token template key, exactly as the tokens route
	// lists it ("Treasure", "1/1 white Soldier").
	Token string `json:"token,omitempty"`
	// PlayerID is the seat that will own and control the spawned
	// cards. Required even for the shared zones (battlefield, exile)
	// because control matters there.
	PlayerID string `json:"player_id"`
	Zone     string `json:"zone"`
	Count    int    `json:"count,omitempty"`
	// Commander stamps IsCommander, so a spawn into the command zone
	// behaves like a real commander rather than a card sitting in the
	// wrong place. Ignored for a token.
	Commander bool `json:"commander,omitempty"`
}

// card projects the request onto the dev spawner's identifier shape,
// so one resolver serves both routes.
func (b spawnRequest) card() devSpawnRequest {
	return devSpawnRequest{ScryfallID: b.ScryfallID, Name: b.Name}
}

type spawnResponse struct {
	Spawned    []string `json:"spawned"`
	Name       string   `json:"name"`
	Zone       string   `json:"zone"`
	ScryfallID string   `json:"scryfall_id,omitempty"`
	Count      int      `json:"count"`
	// Token is true when the request named a token template. The
	// client needs it to know the response's ScryfallID is empty by
	// nature rather than by failure.
	Token bool `json:"token,omitempty"`
}

// spawnCard serves POST /games/{id}/spawn (ADR 0075 §2.4).
func spawnCard(c Config, w http.ResponseWriter, r *http.Request) error {
	gameID, err := requireTableManager(c, r)
	if err != nil {
		return err
	}
	p, _ := auth.PrincipalFromContext(r.Context())

	var body spawnRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}
	playerID, err := uuid.Parse(strings.TrimSpace(body.PlayerID))
	if err != nil {
		return httpError(http.StatusBadRequest, "player_id must be a uuid")
	}
	zone, err := parseSpawnZone(body.Zone)
	if err != nil {
		return err
	}
	count := body.Count
	if count == 0 {
		count = 1
	}

	template, name, scryfallID, isToken, err := resolveSpawnTemplate(c, body)
	if err != nil {
		return err
	}

	ids, err := c.Lobby.Spawn(gameID, SpawnOptions{
		Actor:      p.PlayerID,
		Controller: playerID,
		Zone:       zone,
		Template:   template,
		Count:      count,
		Managed:    true,
	})
	if err != nil {
		return spawnHTTPError(err)
	}

	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return writeJSON(w, http.StatusOK, spawnResponse{
		Spawned:    out,
		Name:       name,
		Zone:       string(zone),
		ScryfallID: scryfallID,
		Count:      len(out),
		Token:      isToken,
	})
}

// resolveSpawnTemplate turns the request into a game.Card template
// plus what the response has to echo back. A token wins over a card
// identifier when both are supplied, the way ScryfallID wins over
// Name: the more specific handle is the one the client just clicked.
func resolveSpawnTemplate(c Config, body spawnRequest) (tmpl game.Card, name, scryfallID string, isToken bool, err error) {
	if key := strings.TrimSpace(body.Token); key != "" {
		if c.Tokens == nil {
			return game.Card{}, "", "", false, httpError(http.StatusServiceUnavailable,
				"token templates are not available on this server")
		}
		t, ok := c.Tokens.TokenTemplate(key)
		if !ok {
			return game.Card{}, "", "", false, httpError(http.StatusNotFound,
				"no token template named "+strconv.Quote(key))
		}
		return t, t.Name, "", true, nil
	}
	if c.Cards == nil {
		return game.Card{}, "", "", false, httpError(http.StatusServiceUnavailable,
			"card index not loaded; run scryfall-refresh.sh")
	}
	card, err := resolveDevCard(c.Cards, body.card())
	if err != nil {
		return game.Card{}, "", "", false, err
	}
	return deck.ToGameCard(card, body.Commander), card.Name, card.ID.String(), false, nil
}

// spawnHTTPError maps the mutation's refusals onto statuses. The
// AllowSpawn refusal is a 403 like the host refusal and says which of
// the two it was.
func spawnHTTPError(err error) error {
	switch {
	case errors.Is(err, ErrSpawnNotAllowed):
		return httpError(http.StatusForbidden,
			"spawning is switched off for this table; the host can turn it on in table settings")
	case errors.Is(err, game.ErrSpawnCount):
		return httpError(http.StatusBadRequest, err.Error())
	case errors.Is(err, game.ErrSpawnZoneUnsupported):
		return httpError(http.StatusBadRequest, err.Error())
	case errors.Is(err, game.ErrSpawnTokenZone):
		return httpError(http.StatusBadRequest, err.Error())
	case errors.Is(err, game.ErrPlayerNotFound):
		return ErrPlayerNotInGame
	}
	return err
}

// spawnCardSearch serves GET /games/{id}/spawn/cards?q=&limit=: the
// same Scryfall index read GET /dev/cards does, for a caller who may
// spawn at THIS table.
//
// It exists because /dev/cards is behind requireDevFeature and 404s
// in production, so without it the production spawner would have a
// name field and no way to find a name. Same response shape, so one
// client component serves both.
func spawnCardSearch(c Config, w http.ResponseWriter, r *http.Request) error {
	if _, err := requireTableManager(c, r); err != nil {
		return err
	}
	return devCardSearch(c, w, r)
}

// requireTableManager is the shared gate for the spawn routes: the
// caller must be this table's host or the server admin. Returns the
// game's id so the caller does not parse the path twice.
func requireTableManager(c Config, r *http.Request) (uuid.UUID, error) {
	gameID, err := gameIDFromPath(r)
	if err != nil {
		return uuid.Nil, err
	}
	p, ok := auth.PrincipalFromContext(r.Context())
	if !ok {
		return uuid.Nil, httpError(http.StatusInternalServerError, "missing principal")
	}
	meta, err := c.Lobby.Get(gameID)
	if err != nil {
		return uuid.Nil, err
	}
	if !CanManageTable(p, meta) {
		return uuid.Nil, httpError(http.StatusForbidden,
			"only the table host or the admin may spawn cards")
	}
	return gameID, nil
}

// spawnTokenList is one row of GET /games/{id}/spawn/tokens.
type spawnTokenList struct {
	Tokens []string `json:"tokens"`
}

// spawnTokens serves GET /games/{id}/spawn/tokens: the keys the
// `token` field of a spawn request accepts.
//
// Scoped to a game although the answer is game-independent, for one
// deployment reason: /games is already in deploy/Caddyfile's @api
// matcher, the Vite proxy and the service worker's API_PATH, and a
// new top-level prefix would have to be added to all three or 404 in
// production only. It carries no game state, so the scoping costs
// nothing else.
//
// Gated like the spawn itself, so the picker is not offered to a seat
// that cannot use it.
func spawnTokens(c Config, w http.ResponseWriter, r *http.Request) error {
	if _, err := requireTableManager(c, r); err != nil {
		return err
	}
	keys := []string{}
	if c.Tokens != nil {
		keys = c.Tokens.TokenKeys()
	}
	return writeJSON(w, http.StatusOK, spawnTokenList{Tokens: keys})
}
