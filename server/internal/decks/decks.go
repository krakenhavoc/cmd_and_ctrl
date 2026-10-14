// Package decks is the pre-built deck catalog: the decks any seat —
// bot or human — may be seated with, and the one lookup the lobby uses
// to turn a deck ID into a validated deck.List.
//
// # Why it is not under internal/aiseat any more
//
// It was, until the lobby grew a human-facing deck picker. The decks
// were never bot-specific; what was bot-specific was that only bots
// could reach them. A player who just wants to play a working game now
// picks from the same catalog, through the same Load, so a package path
// saying `aiseat` would have been a lie about what depends on what.
// The bot picker still reads this catalog, via cmd/server's
// botDeckCatalog adapter, and the model tiers' prompt projection lives
// in internal/aiseat/deckprofile — both import this package; it
// imports neither.
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
// A human gets the weaker version of the same problem — a blank they
// CAN play by hand, because the manual sandbox fallback is a thing a
// player has hands for — plus something a bot never needed: an honest
// statement of what they are picking. Coverage (coverage.go) computes
// that from the same effects registry the catalog page reads, and the
// picker shows it before the deck is installed rather than after a
// card fails to do what it prints.
//
// Basic lands are the one exemption, and they are exempt for a real
// reason rather than convenience: they carry no catalog entry by design.
// game.ActivateManaAbility derives a basic's mana ability from its type
// line when no Spec declares one, so a Forest is fully played by the
// engine while being absent from effects.All().
//
// # What the catalog can and cannot support
//
// Eight decks in eight archetypes:
//
//   - aggro, ramp-stompy (Simic), control (Esper) — the three the
//     sprint planned for. The aggro deck was Izzet (izzet-aggro) until
//     #3071 rebuilt it as mono-red under Purphoros (mono-red-aggro).
//
//   - aristocrats (mono-black) — added because S21's sacrifice outlets
//     and death payoffs and S23's board wipes both landed. The sprint
//     doc deferred this on the grounds that they had not.
//
//   - stompy (mono-green), weenie (mono-white) and voltron (Boros) —
//     added by #554 when the picker opened to human players and four
//     decks stopped being enough of a shop window. Two are mono-coloured
//     on purpose (easier to read, and what a new player reaches for),
//     and all three are built to a stricter bar than the first four:
//     every non-basic card is graded full, and none is on
//     full_test.go's minor-caveats list. Per distinct non-basic card
//     they measure 65 of 65 (Stampede), 68 of 68 (Strength in Numbers)
//     and 72 of 72 (Armed and Dangerous).
//
//   - tribal (five-colour Dragons) — added by #3401 for ADR 0172
//     decision 8: the one built-in deck with a five-colour commander,
//     so the one that can carry Jegantha, the Wellspring as its
//     companion. Every card in it also fits Jegantha's condition.
//
// Voltron was held back for a count problem rather than a missing
// mechanic: the attachment relation shipped (#374, game/attach.go), and
// the first attachments were Bonesplitter, Lightning Greaves,
// Skullclamp, Swiftfoot Boots, two Swords and Rancor, which is a
// creature deck that happens to own a Bonesplitter. The catalog has
// since grown to over two hundred Equipment and Auras graded full,
// which is what let Armed and Dangerous in.
//
// Cards considered and rejected while building the three, each for a
// caveat on the clause its deck would play it for:
//
//   - Stoneforge Mystic — no put-onto-the-battlefield ability, which is
//     the whole reason to play it in Armed and Dangerous.
//   - Brimaz, King of Oreskos — the Cat Soldier it makes on a block is
//     not blocking. Land Tax — the land-count check is made as the
//     trigger goes on the stack. Both are minor against a weenie deck,
//     but the bar here is no caveats at all.
//   - Cultivate — on the minor-caveats list for Deep Roots, so not
//     repeated for Stampede.
//   - Raise the Alarm and Increasing Devotion — full, but they price a
//     printed token count and declare no Purpose (ADR 0126 §6,
//     purpose_test.go), so the bot would price them as blanks.
//   - Thraben Inspector, Skyshroud Claim, Wood Elves, Wayfarer's
//     Bauble — registered but never graded.
//
// The three added here carry no accepted simplifications.
//
// Combo stays out on ADR 0033's judgement that it is a bad idea for a
// bot regardless of coverage.
//
// # Registered is not the same as fully implemented
//
// The coverage test is a floor, not a guarantee. Many catalog cards
// ship with a declared simplification, graded `caveats` in the
// catalog. A card whose RELEVANT clause is unimplemented is a bad pick
// even though it registers, so since #2436 full_test.go holds every
// card here to the catalog's `full` grade, except a short list the
// owner named, each a caveat on a clause its deck does not lean on:
//
//   - Path of Ancestry and Delighted Halfling — with strict mana off,
//     the scry rider and "can't be countered" are inert. Both are here
//     for their mana.
//   - Cultivate — the hand-or-battlefield choice is lost only with one
//     basic left in the library.
//
// Before #2436 the decks also carried cards nobody had graded, and
// cards like Victimize whose simplification made them STRONGER than
// printed. Neither is allowed any more.
//
// # Power level
//
// The decks are "upgraded precons" (#2436): fetchlands and dual lands
// stay, but most tutors and fast mana (Vampiric Tutor, Dark Ritual,
// Lotus Petal, Gaea's Cradle, and Ancient Tomb outside mono-black) are
// cut, and each deck's Summary names the one mechanic it teaches.
// Mono-black keeps Demonic Tutor, Ancient Tomb and Cabal Coffers with
// Urborg: the issue named them as the next trims if the deck still
// dominated, and on the rebalanced lists it did not. A saved table
// setup names a deck by ID, so it seats the current list, whatever it
// held when saved.
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

// Deck is one pre-built deck: a legal Commander deck whose every
// non-basic card the engine actually plays.
type Deck struct {
	// ID is the stable wire identifier the lobby sends to seat a
	// player or a bot with this deck.
	// Kebab-case, never renamed once shipped — a saved lobby config
	// referring to a deck by ID must keep working.
	ID string

	// Name is the display name.
	Name string

	// Archetype is the one-word shape of the deck: aggro, ramp-stompy,
	// control, aristocrats, tribal. Shown in the deck picker and used by the
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

	// Companion is the card waiting outside the game (CR 702.139,
	// ADR 0172), or nil. It is not one of the 100: Size, Cards and the
	// singleton and identity checks leave it out, and Decklist writes it
	// as a Companion section. Its condition is checked by deck.Validate
	// against the commander plus mainboard, like a player's upload.
	Companion *Card
}

// all is the registry, in picker order. Adding a deck means adding it
// here; nothing else discovers decks.
var all = []Deck{
	monoRedAggro,
	simicRamp,
	esperControl,
	monoBlackAristocrats,
	monoGreenStompy,
	monoWhiteWeenie,
	borosVoltron,
	dragonHoard,
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
// and the seat endpoints' validation both want this and neither
// wants the card lists.
func IDs() []string {
	out := make([]string, 0, len(all))
	for _, d := range all {
		out = append(out, d.ID)
	}
	return out
}

// legacyIDs maps a retired deck ID to the ID of the deck that replaced
// it. Migration 0018 rewrites the IDs the database stores, but a
// guest's browser, an engine snapshot and an operator's script keep the
// old one, so Lookup answers it too. The returned Deck carries the NEW
// ID, so whatever is stored from it from then on is the new one. IDs()
// never lists an old ID: the pickers offer each deck once.
var legacyIDs = map[string]string{
	// #3071: the Izzet Pirates deck was rebuilt as mono-red under
	// Purphoros, and its ID followed.
	"izzet-aggro": "mono-red-aggro",
}

// Lookup returns the deck with the given ID. The second return is false
// for an unknown ID — that is a client sending a deck name this build
// does not have, which the caller should surface as a 400 rather than
// substituting a default.
func Lookup(id string) (Deck, bool) {
	if renamed, ok := legacyIDs[id]; ok {
		id = renamed
	}
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

// WithCompanion is Cards plus the companion, if the deck has one: every
// card the deck can put in front of a player, for the checks that ask
// whether the engine plays it.
func (d Deck) WithCompanion() []Card {
	out := d.Cards()
	if d.Companion != nil {
		out = append(out, *d.Companion)
	}
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
// That is deliberate: the pre-built path runs the identical
// ParseText → Resolve → Validate pipeline a pasted list runs, so a
// pre-built deck cannot be legal by a rule an uploaded deck is not
// held to, and there is exactly one legality path in the server.
func (d Deck) Decklist() string {
	var b strings.Builder
	b.WriteString("Commander:\n")
	fmt.Fprintf(&b, "%d %s\n", d.Commander.Copies(), d.Commander.Name)
	b.WriteString("\nDeck:\n")
	for _, c := range d.Mainboard {
		fmt.Fprintf(&b, "%d %s\n", c.Copies(), c.Name)
	}
	if d.Companion != nil {
		b.WriteString("\nCompanion:\n")
		fmt.Fprintf(&b, "1 %s\n", d.Companion.Name)
	}
	return b.String()
}

// Load turns a deck ID into a resolved, validated deck.List, ready to
// seat. This is the whole interface the lobby needs:
//
//	list, err := decks.Load(cfg.Cards, "mono-red-aggro")
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
		return nil, fmt.Errorf("decks: unknown deck %q (have %s)", id, strings.Join(IDs(), ", "))
	}
	return d.Load(idx)
}

// Load is the package-level Load for a Deck already in hand — the
// tutorial pair (tutorial.go), which Lookup does not see because the
// pickers must not offer it. Same pipeline, same oracle-ID check.
func (d Deck) Load(idx *cards.Index) (*deck.List, error) {
	id := d.ID
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

// frontFace is a resolved card's name as a deck file lists it: a
// double-faced card resolves to "Front // Back" (Fable of the
// Mirror-Breaker // Reflection of Kiki-Jiki), and the list names its
// front face.
func frontFace(name string) string {
	front, _, _ := strings.Cut(name, " // ")
	return front
}

// checkOracleIDs asserts that each declared non-basic card resolved to
// the printing this package meant. Compares by name because Resolve
// keeps decklist order per section but the caller should not have to
// rely on that.
func (d Deck) checkOracleIDs(list *deck.List) error {
	want := make(map[string]string, len(d.Mainboard)+1)
	for _, c := range d.WithCompanion() {
		if c.Basic {
			continue
		}
		want[c.Name] = c.OracleID
	}
	resolved := make([]cards.Card, 0, len(list.Commanders)+len(list.Mainboard)+1)
	resolved = append(resolved, list.Commanders...)
	resolved = append(resolved, list.Mainboard...)
	if list.Companion != nil {
		resolved = append(resolved, *list.Companion)
	}

	var mismatches []string
	for _, c := range resolved {
		id, ok := want[frontFace(c.Name)]
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
