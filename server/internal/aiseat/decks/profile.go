package decks

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
)

// profile.go is the model tiers' view of a curated deck: the static,
// prompt-cached half of ADR 0033 §5's Layer C prompt — "decklist with
// oracle text, archetype plan" — built from the deck this package
// already describes.
//
// It lives here, next to the decks, for the same reason the picker's
// adapter does NOT (that one is boot wiring, in cmd/server): this is
// a projection of the deck data, it changes when a deck changes, and
// a deck author adding an archetype should not have to know which
// package assembles prompts.

// Profile is the static half of a model tier's prompt for one curated
// deck: the deck's name, its plan in prose, and its list with oracle
// text. ok is false for an unknown deck ID.
//
// idx supplies the oracle text, mana cost and type line — the same
// Scryfall index the deck is resolved against when the seat is
// installed. A nil index, or a card the index does not have, yields a
// name-only entry rather than an error: the prompt is then thinner
// than it should be, which is a worse bot and not a broken one, and a
// server with no dump loaded has louder problems to report than this.
//
// Nothing here varies with the board, or with time. That is
// load-bearing: this profile is built once per seat and carries the
// prompt-cache breakpoint, so one byte of per-decision state in it
// would silently cost the cache on every call of the game and nothing
// would report it (model/prompt.go says the same at more length).
func Profile(idx *cards.Index, id string) (model.DeckProfile, bool) {
	d, ok := Lookup(id)
	if !ok {
		return model.DeckProfile{}, false
	}
	p := model.DeckProfile{
		Name:      d.Name,
		Archetype: d.plan(),
		Cards:     make([]model.DeckCard, 0, len(d.Mainboard)+1),
	}
	seen := make(map[string]bool, len(d.Mainboard)+1)
	for _, c := range d.Cards() {
		if seen[c.Name] {
			// One entry per distinct card: the profile describes the
			// list, and copy counts (which only basics have) are not
			// something the model is asked to reason about.
			continue
		}
		seen[c.Name] = true
		p.Cards = append(p.Cards, deckCard(idx, c))
	}
	return p, true
}

// plan is the deck's archetype and summary as one short paragraph —
// ADR 0033 §5's "archetype plan". The curated decks carry both halves
// already; this is the join, not new prose.
func (d Deck) plan() string {
	var b strings.Builder
	if d.Archetype != "" {
		b.WriteString("This is a ")
		b.WriteString(d.Archetype)
		b.WriteString(" deck. ")
	}
	b.WriteString(strings.TrimSpace(d.Summary))
	if d.Commander.Name != "" {
		b.WriteString(" Your commander is ")
		b.WriteString(d.Commander.Name)
		b.WriteString(".")
	}
	return strings.TrimSpace(b.String())
}

// deckCard looks one curated card up in the Scryfall index for its
// printed text. Missing from the index (or no index at all) degrades
// to the name alone.
func deckCard(idx *cards.Index, c Card) model.DeckCard {
	out := model.DeckCard{Name: c.Name}
	if idx == nil {
		return out
	}
	found, ok := idx.FindByName(c.Name)
	if !ok {
		return out
	}
	out.Cost = found.ManaCost
	out.Type = found.TypeLine
	out.Oracle = found.OracleText
	return out
}
