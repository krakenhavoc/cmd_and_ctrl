package aiseat

// decks.go is the seam between "a human picked a bot deck in the
// lobby" and "a decklist the existing upload path can install".
//
// THE CONTRACT FOR S31 SUB-PR 5. That sub-PR builds the curated
// archetype decks in server/internal/aiseat/decks/ and the test that
// walks them against effects.All(). All it has to do to be reachable
// from the lobby is satisfy DeckSource and get wired into main.go in
// place of PlaceholderDecks(). Specifically:
//
//   - List() returns one DeckInfo per curated deck, in picker order.
//   - Decklist(id) returns the plain-text decklist for that ID, in
//     exactly the format deck.ParseText already accepts (the same
//     bytes a human could paste into the deck-upload box).
//
// Returning TEXT rather than resolved cards is deliberate. The bot
// seat then runs the identical parse → resolve → validate → install
// pipeline POST /games/{id}/decks runs, against the same Scryfall
// index, so a bot cannot be seated with a deck a human could not
// upload, and a curated deck that rots (a renamed card, a new ban)
// fails loudly at the same place a human's would.

// DeckInfo describes one bot deck to a picker.
type DeckInfo struct {
	// ID is the stable wire identifier sent as `deck` on POST
	// /games/{id}/seats/bot. Lowercase, hyphenated.
	ID string `json:"id"`
	// Name is the display name.
	Name string `json:"name"`
	// Description is one sentence on what the deck tries to do.
	Description string `json:"description,omitempty"`
	// Colors is the deck's colour identity as single WUBRG letters,
	// so the picker can render pips without parsing the list.
	Colors []string `json:"colors,omitempty"`
	// Commander is the deck's commander, for the picker subtitle.
	Commander string `json:"commander,omitempty"`
}

// DeckSource is the named-deck lookup the lobby needs. Implemented by
// PlaceholderDecks today and by server/internal/aiseat/decks in
// sub-PR 5. Implementations must be safe for concurrent use and are
// expected to be immutable after construction.
type DeckSource interface {
	// List returns the catalog in picker order.
	List() []DeckInfo
	// Decklist returns the plain-text decklist for id (deck.ParseText
	// format) along with its DeckInfo. ok is false for an unknown ID.
	Decklist(id string) (DeckInfo, string, bool)
}

// placeholderDeckID is the one deck this sub-PR ships. It is
// deliberately not a real archetype: sub-PR 5 owns deckbuilding, and
// two competing deck registries would be worse than a stand-in.
const placeholderDeckID = "placeholder-mono-red"

// placeholderDecklist is a commander and ninety-nine Mountains: a
// legal Commander deck by every rule deck.Validate enforces
// (singleton exempts basics, colour identity is satisfied, the count
// is 100) that required no deckbuilding judgement to write down. It
// exists so the end-to-end flow — pick a deck, seat a bot, press
// Start, watch it play — is real before the curated decks land.
const placeholderDecklist = `Commander:
1 Krenko, Mob Boss
Mainboard:
99 Mountain
`

type placeholderSource struct{}

// PlaceholderDecks is the DeckSource this sub-PR wires into main.go.
// Replace the wiring with the curated registry when sub-PR 5 lands;
// nothing else has to change.
func PlaceholderDecks() DeckSource { return placeholderSource{} }

func (placeholderSource) List() []DeckInfo {
	return []DeckInfo{{
		ID:          placeholderDeckID,
		Name:        "Mono-red placeholder",
		Description: "A commander and ninety-nine Mountains. A stand-in until the curated archetype decks land (S31 sub-PR 5).",
		Colors:      []string{"R"},
		Commander:   "Krenko, Mob Boss",
	}}
}

func (p placeholderSource) Decklist(id string) (DeckInfo, string, bool) {
	if id != placeholderDeckID {
		return DeckInfo{}, "", false
	}
	return p.List()[0], placeholderDecklist, true
}
