// Package decks is the curated bot-deck catalog: the decks a bot seat
// may be seated with, and the one lookup the lobby uses to turn a deck
// ID into a validated deck.List.
//
// # Why curated, and why a test enforces it
//
// A bot is capped by the engine it plays inside, not by its policy. A
// deck containing a card the catalog does not implement gives the bot a
// card that does nothing when cast — and nothing in the protocol tells
// it so. That is strictly worse than a smaller deck, because the bot
// spends mana and a card on a blank and cannot learn not to. So every
// non-basic card in every deck here resolves to a registered
// effects.Spec, and decks_test.go fails the build the moment one stops
// doing so. ADR 0033 §7 is the decision; this package is its enforcement.
//
// Basic lands are the one exemption, and they are exempt for a real
// reason rather than convenience: they carry no catalog entry by design.
// game.ActivateManaAbility derives a basic's mana ability from its type
// line when no Spec declares one, so a Forest is fully played by the
// engine while being absent from effects.All().
//
// # What the catalog can and cannot support
//
// Four archetypes, which is one more than docs/sprints.md originally
// scoped and two fewer than issue #89 hoped for:
//
//   - aggro (Izzet), ramp-stompy (Simic), control (Esper) — the three
//     the sprint planned for.
//   - aristocrats (mono-black) — added because S21's sacrifice outlets
//     and death payoffs and S23's board wipes both landed. The sprint
//     doc deferred this on the grounds that they had not.
//
// Voltron and Equipment/Aura decks stay out, but the reason has moved
// and is worth stating precisely. The attachment relation shipped
// (#374, game/attach.go), and so did the first attachments (#379, #380)
// — Bonesplitter, Lightning Greaves, Skullclamp, Swiftfoot Boots, two
// Swords and Rancor. Seven. A Voltron deck is fifteen to twenty-five
// attachments; seven of them plus ninety-three other cards is a
// creature deck that happens to own a Bonesplitter. So this is now a
// COUNT problem rather than a missing-machinery problem, and it clears
// on its own as the catalog grows — no engine work is blocking it.
//
// Combo stays out on ADR 0033's judgement that it is a bad idea for a
// bot regardless of coverage.
//
// # Registered is not the same as fully implemented
//
// The coverage test is a floor, not a guarantee. Many catalog cards
// ship with a declared simplification — a clause the engine does not
// model, named in a comment on the card's file. A card whose RELEVANT
// clause is unimplemented is a bad pick even though it registers, so
// every card here was read before it was picked. Two were rejected on
// exactly that test:
//
//   - Gemcutter Buccaneer. "The second half is not modelled" — turning
//     Treasures into Equipment, which is most of why you play it.
//   - Teferi, Time Raveler. Its +1's continuous effect and its static
//     both need machinery the engine does not have, leaving loyalty
//     gain and a bounce.
//
// The simplifications that were accepted are all either cosmetic or
// weaker-than-printed on a clause the deck does not lean on:
//
//   - Solphim, Mayhem Dominus — activated ability not modelled. The
//     noncombat-damage doubling, which is the reason it is in the aggro
//     deck, works.
//   - Farseek — fetches a basic that isn't a Forest rather than any
//     land with those subtypes. The Simic deck runs Islands, so it
//     still ramps.
//   - Filigree Familiar — the sacrifice-for-mana ability is missing;
//     the ETB life and the dies-draw are what it is here for.
//   - Krosan Grip — no split second. Still destroys the artifact.
//   - Path of Ancestry, Raffine's Tower — scry rider and cycling
//     missing. Both still tap for the colours.
//   - Exotic Orchard, Reflecting Pool — one colour short when chained
//     into each other. Weaker than printed, never illegal.
//   - Delighted Halfling — "can't be countered" is inert; it is in the
//     Simic deck as a mana creature.
//   - Rhystic Study — the optional draw is taken automatically, which
//     saves a prompt the bot would always answer the same way.
//   - Victimize — the returned creatures enter untapped, which is
//     STRONGER than printed. The only accepted simplification in that
//     direction, and it favours whichever seat is playing the deck
//     rather than the bot specifically.
//
// # Adding a deck
//
// Copy one of the per-deck files, register it in all, and run the
// package tests. The tests check card count, singleton, colour identity
// and catalog coverage offline; the CMDCTRL_SCRYFALL_DUMP-gated test in
// realdump_manual_test.go does the full deck.Validate against a real
// Scryfall dump, and is the only thing that checks the declared colour
// identities are the ones Scryfall actually prints. Run it.
package decks

import (
	"fmt"
	"sort"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// Card is one row of a curated decklist.
//
// Name is the join key at runtime — deck.Resolve looks it up in the
// Scryfall index by name, exactly as a player-uploaded list is
// resolved. OracleID is the join key at test time, where no Scryfall
// dump exists and the effects registry is the only thing to check
// against. Load re-joins the two and fails if they disagree, which is
// how a renamed or re-keyed card is caught rather than silently
// resolving to something else.
type Card struct {
	// Name is the exact Scryfall front-face name.
	Name string

	// OracleID is the Scryfall oracle ID, and the key the effect
	// catalog registers under. Empty for basic lands, which have no
	// catalog entry by design (see the package comment).
	OracleID string

	// Identity is the card's Commander colour identity as WUBRG
	// letters — "UR", "B", or "" for colourless. Declared here so the
	// colour-identity rule is checked offline; the real-dump test
	// cross-checks these against Scryfall.
	Identity string

	// Count is the number of copies, and is only ever set above one for
	// basic lands. Zero means one copy — see Copies.
	Count int

	// Basic marks a basic land: exempt from the singleton rule and from
	// the catalog-coverage rule.
	Basic bool
}

// Copies is the number of copies of this card in the deck. The zero
// Count means one, so the common singleton row stays free of noise.
func (c Card) Copies() int {
	if c.Count <= 0 {
		return 1
	}
	return c.Count
}

// Deck is one curated bot deck: a legal Commander deck whose every
// non-basic card the engine actually plays.
type Deck struct {
	// ID is the stable wire identifier the lobby sends to seat a bot.
	// Kebab-case, never renamed once shipped — a saved lobby config
	// referring to a deck by ID must keep working.
	ID string

	// Name is the display name.
	Name string

	// Archetype is the one-word shape of the deck: aggro, ramp-stompy,
	// control, aristocrats. Shown in the deck picker and used by the
	// heuristic policy in a later sub-PR to pick a weight set.
	Archetype string

	// Identity is the deck's colour identity as WUBRG letters — the
	// commander's identity, which every mainboard card must be a subset
	// of.
	Identity string

	// Summary is one sentence for the deck picker.
	Summary string

	// Commander is the single command-zone card. Partner is not
	// modelled (deck.Validate rejects a second commander), so this is
	// one card rather than a slice.
	Commander Card

	// Mainboard is the other 99 cards.
	Mainboard []Card
}

// all is the registry, in picker order. Adding a deck means adding it
// here; nothing else discovers decks.
var all = []Deck{
	izzetAggro,
	simicRamp,
	esperControl,
	monoBlackAristocrats,
}

// All returns every curated deck, in a stable picker order. The
// returned slice is freshly allocated, but the Deck values share their
// Mainboard backing arrays with the registry — treat them as read-only.
func All() []Deck {
	out := make([]Deck, len(all))
	copy(out, all)
	return out
}

// IDs returns every deck ID in picker order. The lobby's deck picker
// and the bot-seat endpoint's validation both want this and neither
// wants the card lists.
func IDs() []string {
	out := make([]string, 0, len(all))
	for _, d := range all {
		out = append(out, d.ID)
	}
	return out
}

// Lookup returns the deck with the given ID. The second return is false
// for an unknown ID — that is a client sending a deck name this build
// does not have, which the caller should surface as a 400 rather than
// substituting a default.
func Lookup(id string) (Deck, bool) {
	for _, d := range all {
		if d.ID == id {
			return d, true
		}
	}
	return Deck{}, false
}

// Cards returns the commander followed by the mainboard, one entry per
// row (not per copy). Used by the tests and by Decklist.
func (d Deck) Cards() []Card {
	out := make([]Card, 0, len(d.Mainboard)+1)
	out = append(out, d.Commander)
	out = append(out, d.Mainboard...)
	return out
}

// Size is the total card count including the commander and every copy
// of a basic land. A legal Commander deck has Size 100.
func (d Deck) Size() int {
	n := d.Commander.Copies()
	for _, c := range d.Mainboard {
		n += c.Copies()
	}
	return n
}

// Decklist renders the deck in the plain-text dialect deck.ParseText
// accepts — the same format a player pastes into the deck uploader.
// That is deliberate: the bot-seat path runs the identical
// ParseText → Resolve → Validate pipeline a human upload runs, so a bot
// deck cannot be legal by a rule a human deck is not held to.
func (d Deck) Decklist() string {
	var b strings.Builder
	b.WriteString("Commander:\n")
	fmt.Fprintf(&b, "%d %s\n", d.Commander.Copies(), d.Commander.Name)
	b.WriteString("\nDeck:\n")
	for _, c := range d.Mainboard {
		fmt.Fprintf(&b, "%d %s\n", c.Copies(), c.Name)
	}
	return b.String()
}

// Load turns a deck ID into a resolved, validated deck.List, ready to
// seat. This is the whole interface the lobby needs:
//
//	list, err := decks.Load(catalog.Cards, "izzet-aggro")
//
// It runs the same pipeline as an uploaded decklist — ParseText,
// Resolve against the Scryfall index, Validate — and then adds one
// check a player upload does not get: every resolved card's oracle ID
// must match the one this package declares. A curated deck whose name
// resolved to a different card than the curator picked is exactly the
// silent failure the package exists to prevent, so it is an error and
// not a warning.
//
// Errors are wrapped, so a caller rendering violations can still reach
// *deck.ValidationError and *deck.UnknownCardError with errors.As.
func Load(idx *cards.Index, id string) (*deck.List, error) {
	d, ok := Lookup(id)
	if !ok {
		return nil, fmt.Errorf("decks: unknown bot deck %q (have %s)", id, strings.Join(IDs(), ", "))
	}
	entries, err := deck.ParseText(d.Decklist())
	if err != nil {
		return nil, fmt.Errorf("decks: %s: parse: %w", id, err)
	}
	list, err := deck.Resolve(idx, d.Name, entries)
	if err != nil {
		return nil, fmt.Errorf("decks: %s: resolve: %w", id, err)
	}
	if err := d.checkOracleIDs(list); err != nil {
		return nil, err
	}
	if err := deck.Validate(list); err != nil {
		return nil, fmt.Errorf("decks: %s: %w", id, err)
	}
	return list, nil
}

// checkOracleIDs asserts that each declared non-basic card resolved to
// the printing this package meant. Compares by name because Resolve
// keeps decklist order per section but the caller should not have to
// rely on that.
func (d Deck) checkOracleIDs(list *deck.List) error {
	want := make(map[string]string, len(d.Mainboard)+1)
	for _, c := range d.Cards() {
		if c.Basic {
			continue
		}
		want[c.Name] = c.OracleID
	}
	resolved := make([]cards.Card, 0, len(list.Commanders)+len(list.Mainboard))
	resolved = append(resolved, list.Commanders...)
	resolved = append(resolved, list.Mainboard...)

	var mismatches []string
	for _, c := range resolved {
		id, ok := want[c.Name]
		if !ok {
			continue // basic land, or a name the index normalised
		}
		if got := c.OracleID.String(); got != id {
			mismatches = append(mismatches, fmt.Sprintf("%s (declared %s, index has %s)", c.Name, id, got))
		}
	}
	if len(mismatches) > 0 {
		sort.Strings(mismatches)
		return fmt.Errorf("decks: %s: %d card(s) resolved to a different oracle ID than declared: %s",
			d.ID, len(mismatches), strings.Join(mismatches, "; "))
	}
	return nil
}
