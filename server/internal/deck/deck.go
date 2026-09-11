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

// printedLoyalty parses Scryfall's printed starting loyalty to an
// int. Scryfall puts loyalty at the top level for ordinary cards and
// on the FACE for double-faced ones, so the face list is the
// fallback — otherwise every transforming planeswalker would import
// with zero loyalty and die to CR 704.5i the instant it resolved.
//
// Non-numeric values ("X" on the handful of X-loyalty designs) parse
// to zero, matching how power / toughness handle "*". Those cards
// stay a manual-sandbox case: the player adds counters by hand.
func printedLoyalty(c cards.Card) int {
	if n, err := strconv.Atoi(strings.TrimSpace(c.Loyalty)); err == nil {
		return n
	}
	for _, f := range c.CardFaces {
		if n, err := strconv.Atoi(strings.TrimSpace(f.Loyalty)); err == nil {
			return n
		}
	}
	return 0
}

// printedKeywords translates Scryfall's `keywords` array into the
// engine's canonical keyword tokens. Scryfall capitalises as the
// card prints ("Flying", "First strike") and lists every mechanic
// on the card, including set-specific ones the engine has never
// heard of ("Prepared", "Waterbend"), so game.CanonicalKeyword
// lowercases and drops anything outside the enforced set.
//
// Issues #317 / #319 / #320: before this, printed keywords existed
// ONLY in the opt-in effect catalog, which covers a few hundred
// cards. Vigilance, flash, flying, trample, deathtouch, lifelink,
// first / double strike, menace, reach, haste and defender were all
// inert on every other card — the attacker tapped, the flash spell
// was refused at instant speed, the flier was blocked by ground
// creatures. Keywords are printed card data, like power / toughness
// and starting loyalty (#274), so they travel the same road.
//
// The multi-face narrowing is the one subtlety. Scryfall puts
// `keywords` at the TOP level even for double-faced cards, where it
// is the UNION over every face and the faces carry no arrays of
// their own: Aang, Swift Savior // Aang and La, Ocean's Fury lists
// flash and flying (front) next to reach and trample (back). A
// game.Card is the front face for every purpose that reads keywords
// today, so a keyword the front face's oracle text does not print
// is dropped. The match is against comma-separated entries of a
// line, which is how keyword abilities are printed ("Reach,
// trample"), so prose that merely mentions a keyword ("target
// creature gains trample") can't smuggle one in.
func printedKeywords(c cards.Card) []string {
	if len(c.Keywords) == 0 {
		return nil
	}
	var front map[string]bool
	if len(c.CardFaces) > 1 {
		front = keywordLines(c.CardFaces[0].OracleText)
	}
	var confirm map[string]bool
	out := make([]string, 0, len(c.Keywords))
	for _, raw := range c.Keywords {
		kw, ok := game.CanonicalKeyword(raw)
		if !ok {
			continue
		}
		if front != nil && !front[kw] {
			continue
		}
		// A keyword with a NARROWER printed variant has to be
		// confirmed against the oracle text, because Scryfall tags
		// the variant with the broad name as well. 21 cards —
		// Knight of Grace, Garruk's Harbinger, Sphinx of the
		// Guildpact, six Jaheiras — print only "Hexproof from
		// black" / "Hexproof from monocolored" and carry BOTH
		// "Hexproof" and "Hexproof from" in the array. Taking the
		// array at its word would make every one of them fully
		// untargetable by opponents: STRONGER than printed, which
		// is the direction this repo never errs in. The line scan
		// finds "Hexproof from black" as its own entry, which is
		// not the bare keyword, so the broad grant is dropped and
		// the card keeps the narrow ability it prints — as a
		// simplification, that narrow ability is then not enforced
		// at all, which errs weaker.
		if narrowVariantKeywords[kw] {
			if confirm == nil {
				confirm = keywordLinesOf(c)
			}
			if !confirm[kw] {
				continue
			}
		}
		out = append(out, kw)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// narrowVariantKeywords lists the canonical tokens whose Scryfall
// entry can be a superset of what the card prints, so a bare
// keyword-ability line is required before the token is stamped.
// "Hexproof" is the only one today ("Hexproof from <quality>");
// shroud and the combat keywords have no parameterised form.
var narrowVariantKeywords = map[string]bool{"hexproof": true}

// keywordLinesOf unions the keyword-ability lines across every
// oracle text a printing carries — the top-level one for a
// single-faced card, each face's for a multi-faced one. The
// multi-face NARROWING above is a separate, stricter check; this is
// only asked whether the bare keyword is printed anywhere at all.
func keywordLinesOf(c cards.Card) map[string]bool {
	out := keywordLines(c.OracleText)
	for _, f := range c.CardFaces {
		for kw := range keywordLines(f.OracleText) {
			out[kw] = true
		}
	}
	return out
}

// keywordLines collects the canonical keywords printed as keyword
// abilities in one face's oracle text. A keyword ability occupies
// its own line, alone or comma-separated from its neighbours
// ("Flash", "Flying, vigilance", "Reach, trample"); reminder text
// in parentheses is stripped first so a reminder that names another
// keyword doesn't count. Only exact matches after the split are
// kept, so a sentence is never mistaken for a keyword line.
func keywordLines(text string) map[string]bool {
	out := map[string]bool{}
	for _, line := range strings.Split(text, "\n") {
		if i := strings.IndexByte(line, '('); i >= 0 {
			line = line[:i]
		}
		for _, part := range strings.Split(line, ",") {
			if kw, ok := game.CanonicalKeyword(part); ok {
				out[kw] = true
			}
		}
	}
	return out
}

// oracleTexts returns every oracle text a printing carries: the
// top-level one, then one per face. Scryfall fills exactly one of
// those in — a single-faced card has top-level text and no faces, a
// transform / modal card has null at the top level and text on each
// face — so the union is the whole of the card's printed rules
// without the caller having to know which shape it got.
func oracleTexts(c cards.Card) []string {
	out := make([]string, 0, 1+len(c.CardFaces))
	out = append(out, c.OracleText)
	for _, f := range c.CardFaces {
		out = append(out, f.OracleText)
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
		InstanceID: uuid.New(),
		Name:       c.Name,
		ScryfallID: c.ID.String(),
		OracleID:   c.OracleID.String(),
		TypeLine:   c.TypeLine,
		Power:      power,
		Toughness:  toughness,
		// CR 306.5b — printed starting loyalty. The engine turns
		// this into loyalty counters on battlefield entry; without
		// it the 704.5i SBA eats the walker on the next priority
		// boundary (issue #274).
		StartingLoyalty: printedLoyalty(c),
		// CR 702 — printed keyword abilities. The engine's combat
		// and cast-timing gates read these through
		// game.HasKeyword; before they were carried here they
		// existed only for catalog cards (#317 / #319 / #320).
		Keywords: printedKeywords(c),
		// Does the engine have a generic path for what this card
		// prints, or does it need a hand-written Spec? Answered here
		// because this is the last place the Scryfall record is in
		// scope — game.Card carries no oracle text. See
		// game.NeedsCatalogEffect; the catalog half of the join
		// happens in game.Unimplemented.
		NeedsEffect:  game.NeedsCatalogEffect(c.TypeLine, oracleTexts(c)...),
		ManaCost:     c.ManaCost,
		ProducedMana: append([]string(nil), c.ProducedMana...),
		Colors:       append([]string(nil), c.Colors...),
		// CR 903.4 Commander colour identity. Already parsed off
		// the Scryfall record and already trusted by
		// deck/validate.go; issue #276 was that it had no path onto
		// game.Card, so commanderIdentityFor fell back to the
		// printed mana cost — empty for a double-faced commander,
		// whose cost Scryfall puts on card_faces[0].
		ColorIdentity: append([]string(nil), c.ColorIdentity...),
		IsCommander:   isCommander,
	}
}
