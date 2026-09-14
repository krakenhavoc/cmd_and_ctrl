package decks

import (
	"sort"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
)

// basicLandNames are the only cards allowed into a pre-built deck without a
// catalog entry. They have none by design: game.ActivateManaAbility
// synthesises a basic's mana ability from its type line when no Spec
// declares one, so the engine plays a Forest completely while
// effects.All() has never heard of it.
var basicLandNames = map[string]bool{
	"Plains": true, "Island": true, "Swamp": true,
	"Mountain": true, "Forest": true, "Wastes": true,
}

// TestEveryCardResolvesToARegisteredSpec is the test ADR 0033 §7 asks
// for, and the reason this package exists.
//
// A pre-built deck is only as good as the engine's coverage of it. A card
// with no effects.Spec is not a slightly worse card — it is a blank
// that costs mana and a card and tells nobody, because the manual
// sandbox fallback is a thing a HUMAN uses to play the card by hand and
// a bot has no hands. So this fails the build rather than warning.
func TestEveryCardResolvesToARegisteredSpec(t *testing.T) {
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			var missing []string
			for _, c := range d.Cards() {
				if c.Basic {
					continue
				}
				if _, ok := effects.Lookup(c.OracleID); !ok {
					missing = append(missing, c.Name+" ("+c.OracleID+")")
				}
			}
			if len(missing) > 0 {
				sort.Strings(missing)
				t.Fatalf("%d card(s) have no registered effects.Spec:\n\t%s\n\n"+
					"Either the card was removed from the catalog (replace it here) or "+
					"its OracleID is wrong. A bot cannot play a card the engine does not "+
					"implement — do not silence this by exempting the card.",
					len(missing), strings.Join(missing, "\n\t"))
			}
		})
	}
}

// TestDeclaredNamesMatchTheCatalog guards the other half of the join.
// The OracleID is what the coverage test above checks; the Name is what
// deck.Resolve actually looks up at runtime. A row where those two
// disagree passes the coverage test and then resolves to the wrong card
// — or to nothing — in a real game.
func TestDeclaredNamesMatchTheCatalog(t *testing.T) {
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			for _, c := range d.Cards() {
				if c.Basic {
					continue
				}
				spec, ok := effects.Lookup(c.OracleID)
				if !ok {
					continue // reported by the coverage test
				}
				if spec.Name != c.Name {
					t.Errorf("oracle %s is %q in the catalog but %q here", c.OracleID, spec.Name, c.Name)
				}
			}
		})
	}
}

// TestNoBackFaceOracleKeys pins the MDFC trap. Sixty registry entries
// are MDFC land backs keyed "<oracle_id>#1" whose front faces are
// deliberately unregistered — TestBackFaceSpecsAreKeyedByFace in the
// effects package pins that arrangement. len(effects.All()) is
// therefore not a card count, and a card is NOT covered just because
// grepping the registry for its name hit something.
//
// A bot deck must never reach a card through a back-face key
// regardless of whether the front is registered. The Sieges (S32) do
// register both halves, and even there the deck names the FRONT: a
// back face is not a card you can put in a decklist, it is a face the
// engine reaches by casting or transforming.
func TestNoBackFaceOracleKeys(t *testing.T) {
	for _, d := range All() {
		for _, c := range d.Cards() {
			if strings.Contains(c.OracleID, "#") {
				t.Errorf("%s: %s uses a back-face key %q; the front face is what gets cast",
					d.ID, c.Name, c.OracleID)
			}
		}
	}
}

// TestDeckIsOneHundredCards — CR 903.5a, checked offline. deck.Validate
// checks it too, but only once a Scryfall index has resolved the names,
// and CI has no dump.
func TestDeckIsOneHundredCards(t *testing.T) {
	for _, d := range All() {
		if got := d.Size(); got != 100 {
			t.Errorf("%s: %d cards, want 100", d.ID, got)
		}
	}
}

// TestSingleton — CR 903.5b. Basic lands are exempt; nothing else is.
func TestSingleton(t *testing.T) {
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			seen := make(map[string]int, len(d.Mainboard)+1)
			for _, c := range d.Cards() {
				if c.Basic {
					if c.Copies() < 1 {
						t.Errorf("%s: basic land with %d copies", c.Name, c.Copies())
					}
					continue
				}
				if c.Copies() != 1 {
					t.Errorf("%s: %d copies of a non-basic", c.Name, c.Copies())
				}
				seen[c.Name]++
			}
			for name, n := range seen {
				if n > 1 {
					t.Errorf("%s appears %d times", name, n)
				}
			}
			// A basic land declared without Basic:true would slip past
			// the singleton rule in deck.Validate (which keys off the
			// type line) while being flagged here — and vice versa. Keep
			// the two in agreement.
			for _, c := range d.Cards() {
				if basicLandNames[c.Name] != c.Basic {
					t.Errorf("%s: Basic=%v but basicLandNames says %v", c.Name, c.Basic, basicLandNames[c.Name])
				}
				if c.Basic && c.OracleID != "" {
					t.Errorf("%s: basic lands carry no OracleID, has %q", c.Name, c.OracleID)
				}
				if !c.Basic && c.OracleID == "" {
					t.Errorf("%s: non-basic with no OracleID", c.Name)
				}
			}
		})
	}
}

// TestColourIdentity — CR 903.4. Every mainboard card's identity must be
// a subset of the commander's, and the deck's declared Identity must be
// exactly the commander's.
func TestColourIdentity(t *testing.T) {
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			if d.Identity != d.Commander.Identity {
				t.Fatalf("deck identity %q but commander %s is %q",
					d.Identity, d.Commander.Name, d.Commander.Identity)
			}
			allowed := make(map[rune]bool, len(d.Identity))
			for _, r := range d.Identity {
				if !strings.ContainsRune("WUBRG", r) {
					t.Fatalf("deck identity %q contains %q, which is not a WUBRG symbol", d.Identity, r)
				}
				allowed[r] = true
			}
			for _, c := range d.Mainboard {
				for _, r := range c.Identity {
					if !allowed[r] {
						t.Errorf("%s has identity %q, outside the deck's %q", c.Name, c.Identity, d.Identity)
						break
					}
				}
			}
		})
	}
}

// TestRegistryIsWellFormed — IDs unique and kebab-case, archetypes from
// the known set, no empty display fields. The ID is a wire value the
// lobby sends, so a typo here is a 400 a player cannot fix.
func TestRegistryIsWellFormed(t *testing.T) {
	knownArchetypes := map[string]bool{
		"aggro": true, "ramp-stompy": true, "control": true, "aristocrats": true,
	}
	seen := map[string]bool{}
	for _, d := range All() {
		if seen[d.ID] {
			t.Errorf("duplicate deck ID %q", d.ID)
		}
		seen[d.ID] = true
		if d.ID != strings.ToLower(d.ID) || strings.ContainsAny(d.ID, " _") {
			t.Errorf("deck ID %q should be lower-case kebab-case", d.ID)
		}
		if d.Name == "" || d.Summary == "" {
			t.Errorf("%s: Name and Summary are shown in the picker and must be set", d.ID)
		}
		if !knownArchetypes[d.Archetype] {
			t.Errorf("%s: unknown archetype %q — add it to the heuristic policy's weight sets first", d.ID, d.Archetype)
		}
	}
	if len(IDs()) != len(All()) {
		t.Errorf("IDs() returned %d ids for %d decks", len(IDs()), len(All()))
	}
}

// TestLookupUnknown — an unknown ID is a miss, not a default. Seating a
// bot with a silently-substituted deck is worse than refusing.
func TestLookupUnknown(t *testing.T) {
	if _, ok := Lookup("no-such-deck"); ok {
		t.Fatal("Lookup returned a deck for an unknown ID")
	}
	for _, id := range IDs() {
		if _, ok := Lookup(id); !ok {
			t.Errorf("Lookup(%q) missed a deck IDs() advertises", id)
		}
	}
}

// TestDecklistParsesAsPlayersUploadsDo closes the loop on the format
// claim in Decklist's doc comment: the text this package emits is the
// text deck.ParseText accepts, with the commander in the command zone
// and 99 cards in the mainboard.
func TestDecklistParsesAsPlayersUploadsDo(t *testing.T) {
	for _, d := range All() {
		t.Run(d.ID, func(t *testing.T) {
			entries, err := deck.ParseText(d.Decklist())
			if err != nil {
				t.Fatalf("ParseText: %v", err)
			}
			var commanders, mainboard int
			for _, e := range entries {
				switch {
				case e.IsSideboard:
					t.Errorf("%s landed in the sideboard; Commander has none", e.Name)
				case e.IsCommander:
					commanders += e.Count
				default:
					mainboard += e.Count
				}
			}
			if commanders != 1 {
				t.Errorf("parsed %d commanders, want 1", commanders)
			}
			if mainboard != 99 {
				t.Errorf("parsed %d mainboard cards, want 99", mainboard)
			}
		})
	}
}

// TestLoadRejectsUnknownID — Load's unknown-deck path, which needs no
// card index to reach.
func TestLoadRejectsUnknownID(t *testing.T) {
	if _, err := Load(nil, "no-such-deck"); err == nil {
		t.Fatal("Load accepted an unknown deck ID")
	} else if !strings.Contains(err.Error(), "unknown deck") {
		t.Errorf("unhelpful error: %v", err)
	}
}
