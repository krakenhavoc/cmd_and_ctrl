// Package deck owns deck-list parsing and Commander-format
// validation. A decklist comes in as either JSON (Moxfield export)
// or plain text; the parser resolves each name against the
// cards.Index, and the validator enforces the Commander rules the
// sandbox cares about at S05: 100-card singleton, exactly one
// commander, color identity subset, and format legality pinned to
// the game's banlist snapshot.
//
// Banlist pinning is an explicit NON-goal at S05: Validate reads
// legalities live out of the cards.Index, and the server loads its
// Scryfall dump exactly once at startup and never hot-reloads. That
// pins the banlist to whatever the operator shipped with — good
// enough for a home server. When a future sprint introduces
// mid-process refreshes, this validator will need to either freeze
// a per-game snapshot of Legalities at Create() time or consult the
// timestamp stored on List.ParsedAt; neither is wired today.
package deck

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// List is a parsed, resolved decklist. Cards are by reference to the
// Scryfall index; the UI and the game package dereference them via
// cards.Card.ID. Instance IDs are allocated on game-start, not here.
type List struct {
	// Name is the deck's display name (from the source file, if any).
	// Not load-bearing — the UI shows it alongside the player name in
	// the lobby.
	Name string `json:"name,omitempty"`

	// Commanders is the set of cards placed in the command zone. At
	// S05 this must have exactly one entry; partner / companion
	// support is deferred (and the parser emits an explicit
	// "unsupported" error when it sees those mechanics).
	Commanders []cards.Card `json:"commanders"`

	// Mainboard is the 99 (or more, pre-validation) cards placed in
	// the library. Order is the original decklist order; the game's
	// Start step shuffles it.
	Mainboard []cards.Card `json:"mainboard"`

	// Sideboard is carried through from the source but not used at
	// S05 — Commander doesn't use a sideboard in casual play. We
	// keep it so we can surface warnings if a player uploads a
	// Standard-style decklist by mistake.
	Sideboard []cards.Card `json:"sideboard,omitempty"`

	// ParsedAt is wall-clock time when this list was resolved against
	// the Scryfall index. The validator uses this to cross-check
	// against the game's format-pin date.
	ParsedAt time.Time `json:"parsed_at"`
}

// Entry is a pre-resolution decklist row: a card name + quantity +
// which board it lives in. Parsers produce []Entry; Resolve turns
// that into a List by looking up each name in the cards.Index.
type Entry struct {
	Name        string
	Count       int
	IsCommander bool
	IsSideboard bool
}

// ErrUnsupportedMechanic is returned when a decklist advertises
// partner, companion, or similar mechanics that S05 doesn't model.
// Surfaced as a parse error (not a validation error) because the
// parser bails out before producing a List.
var ErrUnsupportedMechanic = errors.New("deck: unsupported mechanic (partner/companion) — deferred to a later sprint")

// UnknownCardError is returned when a decklist references a card the
// Scryfall index doesn't know about. Surfaces all unknown names at
// once so the user can fix the whole deck in one round-trip rather
// than rebuild N times.
type UnknownCardError struct {
	Names []string
}

func (e *UnknownCardError) Error() string {
	if len(e.Names) == 1 {
		return fmt.Sprintf("deck: unknown card %q", e.Names[0])
	}
	return fmt.Sprintf("deck: %d unknown cards: %s", len(e.Names), strings.Join(e.Names, ", "))
}

// Violations projects the unknown-name list into the same Violation
// shape Validate emits, so the HTTP layer can render a single 422
// `{"error", "violations": [...]}` body regardless of whether the
// failure came from Resolve or Validate.
func (e *UnknownCardError) Violations() []Violation {
	out := make([]Violation, 0, len(e.Names))
	for _, n := range e.Names {
		out = append(out, Violation{
			Code:    CodeUnknownCard,
			Card:    n,
			Message: fmt.Sprintf("%q was not found in the card index", n),
		})
	}
	return out
}

// UnsupportedMechanicError is returned when Resolve finds a card
// whose oracle text leans on a mechanic (partner/companion) the
// server doesn't model yet. Carries the offending card name so the
// HTTP layer can surface it as a structured violation.
type UnsupportedMechanicError struct {
	Card string
}

func (e *UnsupportedMechanicError) Error() string {
	return fmt.Sprintf("%s: %s", ErrUnsupportedMechanic.Error(), e.Card)
}

func (e *UnsupportedMechanicError) Unwrap() error { return ErrUnsupportedMechanic }

func (e *UnsupportedMechanicError) Violations() []Violation {
	return []Violation{{
		Code:    CodeUnsupportedMechanic,
		Card:    e.Card,
		Message: fmt.Sprintf("%q uses partner/companion, which is deferred to a later sprint", e.Card),
	}}
}

// Resolve turns a slice of Entry rows into a List by looking each
// name up in idx. Unknown names are collected into a single
// UnknownCardError so a bad decklist fails in one shot rather than
// name-by-name. ResolveName is case-insensitive (see cards.Index).
func Resolve(idx *cards.Index, name string, entries []Entry) (*List, error) {
	if idx == nil {
		return nil, errors.New("deck: no card index available — run scryfall-refresh.sh")
	}
	list := &List{Name: strings.TrimSpace(name), ParsedAt: time.Now().UTC()}
	var unknown []string

	for _, e := range entries {
		if e.Count <= 0 {
			continue
		}
		c, ok := idx.FindByName(e.Name)
		if !ok {
			unknown = append(unknown, e.Name)
			continue
		}
		// Unsupported-mechanic sniff: if the oracle text mentions
		// partner or companion at all, the deck is probably leaning
		// on a mechanic we don't model yet. Bail rather than silently
		// coerce the second commander into the mainboard.
		if e.IsCommander && mentionsUnsupportedMechanic(c) {
			return nil, &UnsupportedMechanicError{Card: c.Name}
		}
		switch {
		case e.IsCommander:
			for range e.Count {
				list.Commanders = append(list.Commanders, c)
			}
		case e.IsSideboard:
			for range e.Count {
				list.Sideboard = append(list.Sideboard, c)
			}
		default:
			for range e.Count {
				list.Mainboard = append(list.Mainboard, c)
			}
		}
	}

	if len(unknown) > 0 {
		return nil, &UnknownCardError{Names: unknown}
	}
	return list, nil
}

// mentionsUnsupportedMechanic returns true if a card advertises a
// mechanic we defer to a later sprint. Match is on the top-level
// oracle text as well as each face's oracle text (partner cards
// like Thrasios have the clause on the card's main face; meld /
// split cards keep their legal text on one of the faces). We use a
// coarse substring match against "Partner" and "Companion —"
// (with the em-dash separator Scryfall uses), which is precise
// enough to cover both partner variants and every companion.
func mentionsUnsupportedMechanic(c cards.Card) bool {
	if hasUnsupportedPhrase(c.OracleText) {
		return true
	}
	for _, face := range c.CardFaces {
		if hasUnsupportedPhrase(face.OracleText) {
			return true
		}
	}
	return false
}

func hasUnsupportedPhrase(text string) bool {
	// "Partner with" is the named-partner variant; plain "Partner"
	// anchors the generic one. Companion uses the em-dash delimiter
	// that Scryfall ships (U+2014) to introduce the restriction.
	if strings.Contains(text, "Partner with ") {
		return true
	}
	if strings.Contains(text, "Companion \u2014 ") {
		return true
	}
	// Word-boundary match for plain "Partner". Avoid false positives
	// on card names or flavor text by requiring a non-letter
	// character on each side.
	needle := "Partner"
	idx := strings.Index(text, needle)
	for idx != -1 {
		leftOK := idx == 0 || !isLetter(text[idx-1])
		rightIdx := idx + len(needle)
		rightOK := rightIdx == len(text) || !isLetter(text[rightIdx])
		if leftOK && rightOK {
			return true
		}
		text = text[idx+1:]
		idx = strings.Index(text, needle)
	}
	return false
}

func isLetter(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z')
}

// ToGameCards converts a resolved List into the []game.Card slice
// that game.AddPlayer / game.ReplaceDeck expects. Each card gets a
// fresh InstanceID (game uses these to distinguish physical copies
// of the same printed card), the Scryfall UUID is stamped for
// client-side image lookup, and IsCommander is set on each
// commander card. Owner / Controller are left as uuid.Nil — the
// game package re-stamps them at seat time.
func (l *List) ToGameCards() []game.Card {
	out := make([]game.Card, 0, len(l.Commanders)+len(l.Mainboard))
	for _, c := range l.Commanders {
		out = append(out, toGameCard(c, true))
	}
	for _, c := range l.Mainboard {
		out = append(out, toGameCard(c, false))
	}
	return out
}

func toGameCard(c cards.Card, isCommander bool) game.Card {
	// Parse Scryfall's printed power/toughness strings to ints.
	// Non-numeric values ("*", "1+*", "?", empty) parse to zero —
	// good enough for the combat-damage auto-resolve, and the
	// sandbox lets players manually adjust life for the exotic cases.
	power, _ := strconv.Atoi(strings.TrimSpace(c.Power))
	toughness, _ := strconv.Atoi(strings.TrimSpace(c.Toughness))
	return game.Card{
		InstanceID:  uuid.New(),
		Name:        c.Name,
		ScryfallID:  c.ID.String(),
		OracleID:    c.OracleID.String(),
		TypeLine:    c.TypeLine,
		Power:       power,
		Toughness:   toughness,
		IsCommander: isCommander,
	}
}
