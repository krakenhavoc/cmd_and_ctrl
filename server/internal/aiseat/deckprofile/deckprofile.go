// Package deckprofile is the model tiers' view of a pre-built deck:
// the static, prompt-cached half of ADR 0033 §5's Layer C prompt —
// "decklist with oracle text, archetype plan" — built from the deck
// internal/decks already describes.
//
// # Why it is its own package
//
// It used to live beside the decks, as internal/aiseat/decks/profile.go,
// and the argument for that was that a projection of deck data belongs
// next to the data. That argument survived only while the decks were
// bot-only. They are not: a human picks from the same four in the
// lobby, so the deck catalog moved to internal/decks and became a
// neutral package, and a neutral package must not import
// internal/aiseat/model to carry an LLM prompt's struct back out.
//
// So the projection moved instead, and the dependency now points the
// way the rest of aiseat already points: this package imports the deck
// catalog and the model types; nothing in internal/decks knows either
// exists. It is not in cmd/server with the other boot wiring because
// it is a tested projection with rules of its own (one row per
// distinct card, index misses degrade to a bare name), and not in
// internal/aiseat/tiers because factory.go is deliberate about not
// depending on the deck catalog — see DeckProfileFunc's comment.
package deckprofile

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/model"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/decks"
)

// Build returns the static half of a model tier's prompt for one
// pre-built deck: the deck's name, its plan in prose, and its list
// with oracle text. ok is false for an unknown deck ID.
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
func Build(idx *cards.Index, id string) (model.DeckProfile, bool) {
	d, ok := decks.Lookup(id)
	if !ok {
		return model.DeckProfile{}, false
	}
	p := model.DeckProfile{
		Name:      d.Name,
		Archetype: plan(d),
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
// ADR 0033 §5's "archetype plan". The pre-built decks carry both
// halves already; this is the join, not new prose.
func plan(d decks.Deck) string {
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

// deckCard looks one card up in the Scryfall index for its printed
// text. Missing from the index (or no index at all) degrades to the
// name alone.
func deckCard(idx *cards.Index, c decks.Card) model.DeckCard {
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
