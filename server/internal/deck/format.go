package deck

import (
	"fmt"
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// format.go is the deck check of every format but Commander (ADR 0222
// §2). It is one function driven by the format's game.DeckRules row, so
// Pioneer, Modern and the rest are rows, not code.
//
// The rules it enforces, from the Comprehensive Rules:
//
//	CR 100.2a: "In constructed play (a way of playing in which each
//	player creates their own deck ahead of time), each deck has a
//	minimum deck size of 60 cards. A constructed deck may contain any
//	number of basic land cards and no more than four of any card with a
//	particular English name other than basic land cards."
//
//	CR 100.4a: "In constructed play, a sideboard may contain no more
//	than fifteen cards. The four-card limit (see rule 100.2a) applies to
//	the combined deck and sideboard."
//
// and the format's card pool, read from Scryfall's legalities map: a
// card is in when its entry is "legal". Nothing here computes rotation.

// validateConstructed checks list against spec's deck rules.
//
// The companion, when the list has one, is a sideboard card (CR
// 702.139a: "from outside the game", which in constructed play is the
// sideboard, CR 100.4), so it counts toward the sideboard's size and
// toward its name's copies.
func validateConstructed(list *List, spec game.FormatSpec) error {
	rules := spec.Deck
	var vs []Violation

	// A commander: the format has none.
	if !spec.CommanderExpected {
		seen := map[string]bool{}
		for _, c := range list.Commanders {
			if seen[c.Name] {
				continue
			}
			seen[c.Name] = true
			vs = append(vs, Violation{
				Code:    CodeCommanderNotAllowed,
				Card:    c.Name,
				Message: fmt.Sprintf("%q is listed as a commander; a %s deck has none, so list it with the rest of the deck or take it out", c.Name, spec.Name),
			})
		}
	}

	// The deck's size (CR 100.2a). The commanders are counted with the
	// deck so the number matches what the player listed.
	if total := len(list.Commanders) + len(list.Mainboard); total < rules.MinCards {
		vs = append(vs, Violation{
			Code:    CodeWrongCardCount,
			Message: fmt.Sprintf("deck has %d cards; a %s deck needs at least %d", total, spec.Name, rules.MinCards),
		})
	}

	// The sideboard's size (CR 100.4a).
	side := constructedSideboard(list)
	if n := len(side); n > rules.SideboardSize {
		vs = append(vs, Violation{
			Code:    CodeSideboardOverLimit,
			Message: fmt.Sprintf("sideboard has %d cards; a %s sideboard can have at most %d", n, spec.Name, rules.SideboardSize),
		})
	}

	// Copies by English name, across the deck and the sideboard (CR
	// 100.2a, 100.4a). Basic lands are exempt, and so is a card whose
	// own text says how many a deck may hold (CR 100.2a's limit is the
	// default that text overrides, as in Commander).
	pool := make([]cards.Card, 0, len(list.Commanders)+len(list.Mainboard)+len(side))
	pool = append(pool, list.Commanders...)
	pool = append(pool, list.Mainboard...)
	pool = append(pool, side...)
	seen := make(map[string]int, len(pool))
	limits := make(map[string]copyLimit, len(pool))
	for _, c := range pool {
		if isBasicLand(c) {
			continue
		}
		seen[c.Name]++
		limits[c.Name] = deckCopyLimit(c)
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		n, lim := seen[name], limits[name]
		limit := rules.MaxCopies
		if lim.max > limit {
			limit = lim.max
		}
		if lim.any || n <= limit {
			continue
		}
		vs = append(vs, Violation{
			Code:    CodeTooManyCopies,
			Card:    name,
			Message: fmt.Sprintf("%q appears %d times across the deck and sideboard; a %s deck can have at most %d", name, n, spec.Name, limit),
		})
	}

	// The card pool: every card, sideboard included, is legal in the
	// format. One violation per distinct card. The companion is left to
	// companionViolations, which says the same of it once.
	reported := map[string]bool{}
	legality := func(cs []cards.Card) {
		for _, c := range cs {
			if reported[c.Name] {
				continue
			}
			if legal := c.Legalities[rules.Legality]; legal != "legal" {
				reported[c.Name] = true
				vs = append(vs, Violation{
					Code:    CodeNotLegalInFormat,
					Card:    c.Name,
					Message: fmt.Sprintf("%q is not legal in the %s format (status: %q)", c.Name, rules.Legality, legal),
				})
			}
		}
	}
	legality(list.Commanders)
	legality(list.Mainboard)
	legality(list.Sideboard)
	// An Attraction or a sticker sheet is in no constructed format this
	// build knows; the same line says so if a list names one.
	legality(list.Attractions)
	legality(list.StickerSheets)

	vs = append(vs, unsupportedLayoutViolations(list)...)
	vs = append(vs, meldBackFaceViolations(list)...)
	vs = append(vs, companionViolations(list, spec)...)

	if len(vs) == 0 {
		return nil
	}
	return &ValidationError{Violations: vs}
}

// constructedSideboard is the list's sideboard as constructed play
// counts it: the sideboard rows plus the companion, which is one of the
// fifteen.
func constructedSideboard(list *List) []cards.Card {
	side := make([]cards.Card, 0, len(list.Sideboard)+1)
	side = append(side, list.Sideboard...)
	if list.Companion != nil {
		side = append(side, *list.Companion)
	}
	return side
}

// EntriesFor adapts parsed rows to a format before they are resolved.
// Commander's rows are returned as they are. For a format with no
// commander, the rows ParseText only GUESSED were commanders (a short
// unheaded last block, the shape of a Commander export) are read as the
// sideboard instead, which is what an unheaded last block means in a
// constructed list. A row the list itself marked as a commander is left
// alone, for the deck check to refuse.
func EntriesFor(format game.Format, entries []Entry) []Entry {
	if format.Spec().CommanderExpected {
		return entries
	}
	out := append([]Entry(nil), entries...)
	for i := range out {
		if out[i].IsCommander && out[i].CommanderInferred {
			out[i].IsCommander = false
			out[i].CommanderInferred = false
			out[i].IsSideboard = true
		}
	}
	return out
}
