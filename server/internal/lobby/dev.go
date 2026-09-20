package lobby

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/auth"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dev.go holds the HTTP surface for the develop environment's card
// spawner (ADR 0023). Every route here is registered behind
// requireDevFeature, so in production they are 404s and this file's
// handlers are unreachable.
//
// Why HTTP rather than a new WS action type: the card index lives in
// lobby.Config, the lobby already owns the mutate-and-broadcast path
// that deck upload uses, and requireDevFeature already gates HTTP.
// Routing it through the action protocol would mean threading the
// index into actions.Dispatch and adding a second, parallel
// environment gate on the WS side — more moving parts for no gain.

// devCardSearchLimit caps results per query regardless of what the
// caller asks for.
const devCardSearchLimit = 40

// devCardResult is one row of GET /dev/cards.
type devCardResult struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TypeLine string `json:"type_line"`
	ManaCost string `json:"mana_cost"`
	SetCode  string `json:"set"`
}

// devCardSearch serves GET /dev/cards?q=&limit=. Not scoped to a
// game: it is a read of the Scryfall index and carries no game
// state. 503 when the index is unloaded, matching the deck-upload
// endpoint's posture on a fresh deployment.
func devCardSearch(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "card index not loaded; run scryfall-refresh.sh")
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		return writeJSON(w, http.StatusOK, map[string][]devCardResult{"cards": {}})
	}
	limit := devCardSearchLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil && n > 0 && n < limit {
			limit = n
		}
	}
	found := c.Cards.Search(q, limit)
	out := make([]devCardResult, 0, len(found))
	for _, card := range found {
		out = append(out, devCardResult{
			ID:       card.ID.String(),
			Name:     card.Name,
			TypeLine: card.TypeLine,
			ManaCost: card.ManaCost,
			SetCode:  card.SetCode,
		})
	}
	return writeJSON(w, http.StatusOK, map[string][]devCardResult{"cards": out})
}

// devSpawnRequest is the body of POST /games/{id}/dev/spawn.
//
// Exactly one of ScryfallID or Name identifies the card. Name is
// accepted because typing one is how the tool is actually used; the
// ID path exists so the client can pin the exact printing it showed
// in the search results rather than re-resolving a name that may be
// ambiguous.
type devSpawnRequest struct {
	ScryfallID string `json:"scryfall_id,omitempty"`
	Name       string `json:"name,omitempty"`
	// PlayerID is the seat that will own and control the spawned
	// cards. Required even for the shared zones (battlefield, exile)
	// because control matters there.
	PlayerID string `json:"player_id"`
	Zone     string `json:"zone"`
	Count    int    `json:"count,omitempty"`
	// Commander stamps IsCommander, so a spawn into the command zone
	// behaves like a real commander rather than a card sitting in the
	// wrong place.
	Commander bool `json:"commander,omitempty"`
}

type devSpawnResponse struct {
	Spawned     []string `json:"spawned"`
	Name        string   `json:"name"`
	Zone        string   `json:"zone"`
	ScryfallID  string   `json:"scryfall_id"`
	InstanceIDs int      `json:"count"`
}

// devSpawnCard serves POST /games/{id}/dev/spawn.
func devSpawnCard(c Config, w http.ResponseWriter, r *http.Request) error {
	if c.Cards == nil {
		return httpError(http.StatusServiceUnavailable, "card index not loaded; run scryfall-refresh.sh")
	}
	gameID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		return httpError(http.StatusBadRequest, "game id must be a uuid")
	}
	var body devSpawnRequest
	if err := decodeJSON(w, r, &body); err != nil {
		return err
	}

	card, err := resolveDevCard(c.Cards, body)
	if err != nil {
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

	template := deck.ToGameCard(card, body.Commander)
	// Actor names the spawner in the game-log line ADR 0075 added.
	// Everything else about this route is unchanged: no AllowSpawn
	// check, no undo entry, still 404 outside a dev deployment.
	var actor uuid.UUID
	if p, ok := auth.PrincipalFromContext(r.Context()); ok {
		actor = p.PlayerID
	}
	ids, err := c.Lobby.Spawn(gameID, SpawnOptions{
		Actor:      actor,
		Controller: playerID,
		Zone:       zone,
		Template:   template,
		Count:      count,
	})
	if err != nil {
		return spawnHTTPError(err)
	}

	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, id.String())
	}
	return writeJSON(w, http.StatusOK, devSpawnResponse{
		Spawned:     out,
		Name:        card.Name,
		Zone:        string(zone),
		ScryfallID:  card.ID.String(),
		InstanceIDs: len(out),
	})
}

// resolveDevCard turns the request's card identifier into an indexed
// card. ScryfallID wins when both are supplied.
func resolveDevCard(idx *cards.Index, body devSpawnRequest) (cards.Card, error) {
	if id := strings.TrimSpace(body.ScryfallID); id != "" {
		parsed, err := uuid.Parse(id)
		if err != nil {
			return cards.Card{}, httpError(http.StatusBadRequest, "scryfall_id must be a uuid")
		}
		card, ok := idx.Get(parsed)
		if !ok {
			return cards.Card{}, httpError(http.StatusNotFound, "no card with that scryfall_id in the index")
		}
		return card, nil
	}
	if name := strings.TrimSpace(body.Name); name != "" {
		card, ok := idx.FindByName(name)
		if !ok {
			return cards.Card{}, httpError(http.StatusNotFound, "no card named "+strconv.Quote(name)+" in the index")
		}
		return card, nil
	}
	return cards.Card{}, httpError(http.StatusBadRequest, "one of scryfall_id or name is required")
}

// parseSpawnZone maps the wire string onto a ZoneKind, rejecting the
// stack explicitly rather than letting it fall through to a generic
// "unknown zone" — a caller asking for the stack is asking for
// something coherent that this tool deliberately does not do.
func parseSpawnZone(raw string) (game.ZoneKind, error) {
	switch game.ZoneKind(strings.ToLower(strings.TrimSpace(raw))) {
	case game.ZoneBattlefield:
		return game.ZoneBattlefield, nil
	case game.ZoneHand:
		return game.ZoneHand, nil
	case game.ZoneGraveyard:
		return game.ZoneGraveyard, nil
	case game.ZoneExile:
		return game.ZoneExile, nil
	case game.ZoneLibrary:
		return game.ZoneLibrary, nil
	case game.ZoneCommand:
		return game.ZoneCommand, nil
	case game.ZoneStack:
		return "", httpError(http.StatusBadRequest,
			"cannot spawn onto the stack; spawn to hand and cast it")
	default:
		return "", httpError(http.StatusBadRequest, "zone must be one of: battlefield, hand, graveyard, exile, library, command")
	}
}
