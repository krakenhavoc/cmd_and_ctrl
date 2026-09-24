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
	"sort"
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

// ToGameCard converts one indexed Scryfall card into the game.Card
// the engine plays with. Exported so the develop environment's card
// spawner (ADR 0023) produces instances identical to imported ones
// rather than a lookalike that drifts from this function.
func ToGameCard(c cards.Card, isCommander bool) game.Card {
	return toGameCard(c, isCommander)
}

// printedLoyalty parses Scryfall's top-level printed starting
// loyalty to an int (CR 306.5b).
//
// It used to fall back to "the first face that prints a loyalty",
// because Scryfall puts loyalty on the FACE for double-faced cards
// and leaves the top-level field empty — without which every
// transforming planeswalker imported with zero loyalty and died to
// the CR 704.5i SBA the instant it resolved (#274).
//
// ADR 0034 makes that fallback both unnecessary and wrong. Per-face
// loyalty now lives on game.Face and SetFace(0) materialises the
// FRONT face's, which for Nissa, Vastwood Seer // Nissa, Sage
// Animist is correctly zero: the front face is a 4/4 Elf Scout, and
// handing it the back face's 3 was the whole-card fallback papering
// over the missing face model.
//
// Non-numeric values ("X" on the handful of X-loyalty designs) parse
// to zero, matching how power / toughness handle "*". Those cards
// stay a manual-sandbox case: the player adds counters by hand.
func printedLoyalty(c cards.Card) int {
	n, _ := strconv.Atoi(strings.TrimSpace(c.Loyalty))
	return n
}

// printedDefense is printedLoyalty for battles (CR 310.4). Same
// posture on a non-numeric value: zero, and a manual sandbox case.
// Nothing printed puts a non-numeric defense on a card, but the
// parse is the same shape so the failure mode is too.
func printedDefense(c cards.Card) int {
	n, _ := strconv.Atoi(strings.TrimSpace(c.Defense))
	return n
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
	wantProtection := false
	out := make([]string, 0, len(c.Keywords))
	for _, raw := range c.Keywords {
		kws, ok := game.CanonicalKeywords(raw)
		if !ok {
			// PROTECTION lands here, always: Scryfall's array carries
			// the bare family name and the quality lives only in the
			// oracle text, so the tokens come from the line scan
			// below instead (#662). Everything else Scryfall names
			// that the engine does not enforce is simply dropped.
			if strings.EqualFold(strings.TrimSpace(raw), "protection") {
				wantProtection = true
			}
			continue
		}
		kw := kws[0]
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
	// Protection (CR 702.16) is PARAMETERISED, so it cannot come off
	// the array the way every other keyword does: Scryfall says only
	// "Protection" and the quality is in the oracle text. The tokens
	// are read from the keyword-ability LINES instead, through the
	// same scan the narrow-variant confirmation uses — which also
	// gives the multi-face narrowing for free, because a keyword the
	// front face does not print is not in `front`.
	//
	// A quality protection.go's closed grammar cannot parse mints
	// nothing, so the card carries no protection at all and stays
	// flagged by the ADR 0037 coverage signal. Errs weaker, exactly
	// as ADR 0038 §6 chose for "Hexproof from".
	if wantProtection {
		lines := front
		if lines == nil {
			lines = keywordLinesOf(c)
		}
		var protections []string
		for kw := range lines {
			if _, ok := game.ParseProtectionQuality(kw); ok && !containsString(out, kw) {
				protections = append(protections, kw)
			}
		}
		// Map iteration order is random and Card.Keywords is what the
		// client's badge row renders, so the tokens are sorted — the
		// printed order is not recoverable from a set, and a stable
		// order is what stops a card's badges shuffling between
		// imports.
		sort.Strings(protections)
		out = append(out, protections...)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// containsString is a plain membership test for the small token
// slices in this file.
func containsString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
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
			// CanonicalKeywords, not the singular form: "protection
			// from Demons and from Dragons" is one comma-separated
			// part and two abilities (CR 702.16m, #662).
			kws, ok := game.CanonicalKeywords(part)
			if !ok {
				continue
			}
			for _, kw := range kws {
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

// nonNumericStat reports whether a printed power / toughness string is
// present but not a number ("*", "1+*", "?", "X"), which the parse
// turns into a stand-in 0. An empty string is a card with no such
// stat, not a variable one.
func nonNumericStat(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	_, err := strconv.Atoi(s)
	return err != nil
}

// variableToughness answers game.Card.VariableToughness for one set of
// printed values: "the engine has no number for this toughness, and
// the 0 it parsed is a stand-in".
//
// Two ways that happens, and the second is #691's safety net:
//
//   - the string is present and is not a number ("*", "1+*", "?") —
//     nonNumericStat's case, and the whole of #683;
//   - the string is ABSENT on a card or face whose type line says
//     creature. Every creature has a printed toughness, so a missing
//     one is a record the engine cannot read rather than a printed 0,
//     and the 0 this importer writes must not be taken for one.
//
// The second matters because the toughness state-based action now
// kills a printed 0/0 that carries a printing (Card.ToughnessIsKnown
// branch 5). Without this line a Scryfall record that omits a creature
// face's power and toughness — a layout nobody has taught the dump
// reader about yet — would put that creature into the graveyard the
// moment it landed. Flagging it keeps the pre-#683 posture instead:
// the skip covers it, and it lives as a 0/0 nothing can kill, which is
// the failure that can be seen and reported.
//
// A card with no toughness at all and no creature type — every land,
// instant, sorcery and artifact — is not flagged, as before. Nor is a
// JOINED type line ("Creature — Human Wizard // Legendary Planeswalker
// — Jace"): Scryfall leaves the top-level toughness of a multi-face
// printing empty on purpose, the number lives on the face, and
// SetFace(0) overwrites this flag from there anyway.
func variableToughness(typeLine, toughness string) bool {
	if strings.TrimSpace(toughness) != "" {
		return nonNumericStat(toughness)
	}
	if strings.Contains(typeLine, "//") {
		return false
	}
	return strings.Contains(strings.ToLower(typeLine), "creature")
}

// PrintedVariableToughness returns the game.PrintedVariableToughness
// lookup over idx: the answer toGameCard and printedFaces stamp,
// recomputed for a printing by its Scryfall ID. Restoring a game
// written before #683 uses it to backfill Card.VariableToughness (see
// game/snapshot_backfill.go). Faces are reported only for a printing
// with two or more, as printedFaces builds them. Nil for a nil index.
func PrintedVariableToughness(idx *cards.Index) func(scryfallID string) (top bool, faces []bool, ok bool) {
	if idx == nil {
		return nil
	}
	return func(scryfallID string) (bool, []bool, bool) {
		id, err := uuid.Parse(scryfallID)
		if err != nil {
			return false, nil, false
		}
		c, ok := idx.Get(id)
		if !ok {
			return false, nil, false
		}
		var faces []bool
		if len(c.CardFaces) >= 2 {
			faces = make([]bool, len(c.CardFaces))
			for i, f := range c.CardFaces {
				faces[i] = variableToughness(f.TypeLine, f.Toughness)
			}
		}
		return variableToughness(c.TypeLine, c.Toughness), faces, true
	}
}

func toGameCard(c cards.Card, isCommander bool) game.Card {
	// Parse Scryfall's printed power/toughness strings to ints.
	// Non-numeric values ("*", "1+*", "?", empty) parse to zero —
	// good enough for the combat-damage auto-resolve, and the
	// sandbox lets players manually adjust life for the exotic cases.
	power, _ := strconv.Atoi(strings.TrimSpace(c.Power))
	toughness, _ := strconv.Atoi(strings.TrimSpace(c.Toughness))
	out := game.Card{
		InstanceID: uuid.New(),
		Name:       c.Name,
		ScryfallID: c.ID.String(),
		OracleID:   c.OracleID.String(),
		TypeLine:   c.TypeLine,
		Power:      power,
		Toughness:  toughness,
		// #683: the toughness SBA keeps skipping a `*` creature's
		// stand-in 0 after it loses its last counter, where a
		// printed 0/0 dies.
		VariableToughness: variableToughness(c.TypeLine, c.Toughness),
		// CR 306.5b — printed starting loyalty. The engine turns
		// this into loyalty counters on battlefield entry; without
		// it the 704.5i SBA eats the walker on the next priority
		// boundary (issue #274).
		StartingLoyalty: printedLoyalty(c),
		// CR 310.4 — printed defense. Same road as starting loyalty
		// and for the same reason: without it a battle enters with
		// zero defense counters and the 704.5v/w SBA sweeps it before
		// anybody can attack it. Top-level only for the day a
		// single-faced battle is printed; every battle in the game
		// today carries its number on the front FACE, which
		// SetFace(0) materialises a few lines below.
		StartingDefense: printedDefense(c),
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
		// ADR 0034 — the printing's layout and its printed faces.
		// Layout decides what "choose a face" MEANS at announce time
		// (modal DFC: either half; transform: front only), and Faces
		// carries the per-face cost / type / colours that Scryfall
		// leaves null at the TOP level for exactly those layouts.
		Layout: c.Layout,
		Faces:  printedFaces(c),
	}
	// Materialise face 0. For the ~33,000 single-faced oracle IDs
	// this is a no-op and every field above stands as written; for a
	// multi-face card it OVERWRITES Name / TypeLine / ManaCost /
	// Colors / Power / Toughness / StartingLoyalty / StartingDefense
	// with the front face's — which is the point, because the top-level values it
	// replaces are the ones that were null ("" cost ⇒ a free spell)
	// or joined ("Sorcery // Land" ⇒ a sorcery that passed IsLand()
	// and skipped the cost gate entirely, #289).
	out.SetFace(0)
	return out
}

// printedFaces builds the engine's face list from Scryfall's
// card_faces array. Returns nil for a single-faced printing, which
// leaves Card.Faces nil and SetFace a no-op — the entire
// single-faced world is untouched by this change.
//
// A one-element card_faces array is treated as single-faced too.
// Scryfall does not ship those today, but a face list that cannot
// offer a choice is indistinguishable from no face list, and nil is
// the cheaper representation of the same fact.
func printedFaces(c cards.Card) []game.Face {
	if len(c.CardFaces) < 2 {
		return nil
	}
	out := make([]game.Face, 0, len(c.CardFaces))
	for _, f := range c.CardFaces {
		// Same posture as the top-level parse: non-numeric printed
		// values ("*", "1+*", "X") land as zero and stay a manual
		// sandbox case.
		power, _ := strconv.Atoi(strings.TrimSpace(f.Power))
		toughness, _ := strconv.Atoi(strings.TrimSpace(f.Toughness))
		loyalty, _ := strconv.Atoi(strings.TrimSpace(f.Loyalty))
		defense, _ := strconv.Atoi(strings.TrimSpace(f.Defense))
		out = append(out, game.Face{
			Name:              f.Name,
			TypeLine:          f.TypeLine,
			ManaCost:          f.ManaCost,
			Colors:            faceColors(f),
			Power:             power,
			Toughness:         toughness,
			VariableToughness: variableToughness(f.TypeLine, f.Toughness),
			StartingLoyalty:   loyalty,
			StartingDefense:   defense,
			OracleText:        f.OracleText,
		})
	}
	return out
}

// faceColors resolves one face's colours. The order matters, because
// the two Scryfall multi-face shapes populate different fields:
//
//	transform / modal_dfc  `colors` is populated per face.
//	adventure / split      `colors` is null per face; the face's
//	                       mana cost is the only signal.
//	transform BACK faces   have no mana cost at all, and Scryfall
//	                       expresses their colour as CR 105.2c's
//	                       colour indicator — Jace, Telepath Unbound
//	                       is blue by indicator and by nothing else.
//
// So: colors, then color_indicator, then derive from the cost.
// Returns nil rather than an empty slice for a genuinely colourless
// face, because game.Card reads a nil Colors as "not stamped" and
// falls back to the cost — which for a colourless face yields nil
// again, so the two agree.
func faceColors(f cards.CardFace) []string {
	if len(f.Colors) > 0 {
		return append([]string(nil), f.Colors...)
	}
	if len(f.ColorIndicator) > 0 {
		return append([]string(nil), f.ColorIndicator...)
	}
	return game.ColorsInManaCost(f.ManaCost)
}
