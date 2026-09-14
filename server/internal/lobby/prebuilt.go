package lobby

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
)

// prebuilt.go is the human half of the pre-built deck catalog: the
// list a player picks from, and the honest statement of what the
// engine will do with each one.
//
// # Why this is a route and not a constant in the client
//
// The list is fixed at build time, so a static JSON file in the client
// bundle would serve it. It would also drift: the coverage numbers
// come out of the effects registry, which changes on most days, and a
// client-side copy would keep claiming yesterday's coverage after the
// server learned better. The one place that can answer "what does THIS
// build implement" is this build.
//
// # Why it does not reuse GET /bot/options
//
// That route is the bot picker's, gated on a bot host being
// configured, and its aiseat.DeckInfo has no room for a coverage
// profile (it is the bot's wire shape, and a bot does not read
// disclosures). A deployment with no model endpoint still wants a
// human deck picker. The two routes read the same internal/decks
// catalog, which is the part that must not fork.

// prebuiltDeckInfo is one deck as the picker renders it: enough to
// choose between four decks, plus the coverage profile that is the
// whole reason a player would trust one.
type prebuiltDeckInfo struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Archetype string `json:"archetype,omitempty"`
	Summary   string `json:"summary,omitempty"`
	Commander string `json:"commander,omitempty"`
	// Colors is the deck's colour identity as single WUBRG letters,
	// so the picker can render pips without parsing the list.
	Colors []string `json:"colors,omitempty"`
	// CardCount is the physical size — 100 for a legal Commander
	// deck, counting every copy of a basic land.
	CardCount int `json:"card_count"`
	// Coverage is what the engine promises about this deck's cards.
	// Counted per distinct card, which is why its totals do not add
	// up to CardCount: thirty Forests are one basic-land row.
	Coverage decks.Coverage `json:"coverage"`
}

// prebuiltDecksResponse is the body of GET /decks.
type prebuiltDecksResponse struct {
	Decks []prebuiltDeckInfo `json:"decks"`
}

// prebuiltDeckCatalog projects the deck registry onto the picker's
// wire shape, coverage included. One catalog build for all four decks
// — see decks.Coverages.
func prebuiltDeckCatalog(c Config) []prebuiltDeckInfo {
	covs := decks.Coverages(c.Cards)
	all := decks.All()
	out := make([]prebuiltDeckInfo, 0, len(all))
	for _, d := range all {
		colors := make([]string, 0, len(d.Identity))
		for _, ch := range d.Identity {
			colors = append(colors, string(ch))
		}
		out = append(out, prebuiltDeckInfo{
			ID:        d.ID,
			Name:      d.Name,
			Archetype: d.Archetype,
			Summary:   d.Summary,
			Commander: d.Commander.Name,
			Colors:    colors,
			CardCount: d.Size(),
			Coverage:  covs[d.ID],
		})
	}
	return out
}

// prebuiltDecks handles GET /decks: the pre-built decks this build
// offers, with each one's engine-coverage profile.
//
// Session-gated but game-independent — the same answer for every
// table, so the lobby fetches it once on mount. It is served even
// when no Scryfall dump is loaded: the coverage grades come from the
// effects registry and are true regardless, and the honest failure
// for a deployment with no dump is the 503 on installing a deck, not
// an empty picker that implies there are none.
func prebuiltDecks(c Config, w http.ResponseWriter, _ *http.Request) error {
	return writeJSON(w, http.StatusOK, prebuiltDecksResponse{Decks: prebuiltDeckCatalog(c)})
}

// prebuiltDeckName is the display name for a deck id, or "" for one
// this build does not have. Used to label the seat, since a
// plain-text decklist carries no name of its own.
func prebuiltDeckName(id string) string {
	d, ok := decks.Lookup(id)
	if !ok {
		return ""
	}
	return d.Name
}

// uploadDeckChoice turns POST /games/{id}/decks' body into the
// (format, source, deckID) triple resolveDeckSource takes. The
// pre-built branch hands over the deck's TEXT, so the request
// continues down the one pipeline an uploaded list uses — see
// uploadDeckRequest.Deck, and botDeckSource, which is the same
// function for the bot route.
//
// Exactly one of deck and source may be set. Sending both is a 400
// rather than a precedence rule, because either answer would be a
// guess about which one the player meant.
func uploadDeckChoice(body uploadDeckRequest) (format, source, deckID string, err error) {
	named := strings.TrimSpace(body.Deck)
	raw := strings.TrimSpace(body.Source)
	switch {
	case named != "" && raw != "":
		return "", "", "", httpError(http.StatusBadRequest, "send either deck or source, not both")
	case named != "":
		d, ok := decks.Lookup(named)
		if !ok {
			return "", "", "", httpError(http.StatusUnprocessableEntity,
				fmt.Sprintf("unknown deck %q (have %s)", named, strings.Join(decks.IDs(), ", ")))
		}
		return "text", d.Decklist(), d.ID, nil
	case raw != "":
		return body.Format, body.Source, "", nil
	default:
		return "", "", "", httpError(http.StatusBadRequest, "deck or source is required")
	}
}
