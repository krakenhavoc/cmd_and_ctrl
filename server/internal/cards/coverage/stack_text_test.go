package coverage

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/oracletext"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stack_text_test.go — #3830: the stack shows an ability's printed
// text, derived from the Oracle text by oracletext rather than written
// per card. These tests run the derivation over the whole catalog
// against the oracle fixtures, the text production reads from the
// Scryfall index.

// fixtureSource answers oracletext from the oracle fixtures.
func fixtureSource(fx map[string]OracleCard) oracletext.Source {
	return func(id uuid.UUID) (oracletext.Card, bool) {
		c, ok := fx[id.String()]
		if !ok {
			return oracletext.Card{}, false
		}
		out := oracletext.Card{Name: c.Name}
		for _, f := range c.AllFaces() {
			out.Faces = append(out.Faces, oracletext.Face{Name: f.Name, Text: f.Text})
		}
		return out, true
	}
}

func useFixtureSource(t *testing.T) {
	t.Helper()
	oracletext.SetSource(fixtureSource(loadOracleFixture(t)))
	t.Cleanup(func() { oracletext.SetSource(nil) })
}

// ownRef is the ref a stack item of one of key's own rows carries.
func ownRef(key, slot, name string, i int) game.AbilityRef {
	return game.AbilityRef{Key: key, Slot: slot, Ref: game.OwnAbilityRef(i), Name: name}
}

func TestStackAbilityTextOfNamedCards(t *testing.T) {
	useFixtureSource(t)
	cases := []struct {
		card, slot string
		row        int
		want       string
	}{
		{"Purphoros, God of the Forge", game.AbilitySlotTriggered, 0,
			"Whenever another creature you control enters, Purphoros deals 2 damage to each opponent."},
		{"Impact Tremors", game.AbilitySlotTriggered, 0,
			"Whenever a creature you control enters, this enchantment deals 1 damage to each opponent."},
		// An activated ability with a cost: the whole printed line.
		{"Purphoros, God of the Forge", game.AbilitySlotActivated, 0,
			"{2}{R}: Creatures you control get +1/+0 until end of turn."},
	}
	for _, c := range cases {
		s, ok := specNamed(c.card)
		if !ok {
			t.Fatalf("%s is not in the catalog", c.card)
		}
		def := game.CatalogLookup(s.OracleID)
		var name string
		if c.slot == game.AbilitySlotTriggered {
			name = def.Triggered[c.row].Key
		} else {
			name = def.Activated[c.row].Label
		}
		if got := oracletext.AbilityText(ownRef(s.OracleID, c.slot, name, c.row)); got != c.want {
			t.Errorf("%s %s row %d: got %q, want %q", c.card, c.slot, c.row, got, c.want)
		}
	}
}

// TestStackAbilityTextFallsBack: a row with no clause of its own on
// the card's face answers "", and the stack shows the label.
func TestStackAbilityTextFallsBack(t *testing.T) {
	useFixtureSource(t)
	s, ok := specNamed("Purphoros, God of the Forge")
	if !ok {
		t.Fatal("Purphoros is not in the catalog")
	}
	label := game.CatalogLookup(s.OracleID).Triggered[0].Key
	for _, ref := range []game.AbilityRef{
		// A granted bundle's row: the text is the grantor's.
		{Key: game.GrantKey("some-bundle"), Slot: game.AbilitySlotTriggered, Ref: "grant:some-bundle:0:0", Name: label},
		// A label the card has no row for.
		ownRef(s.OracleID, game.AbilitySlotTriggered, "Purphoros, God of the Forge — not a row", 0),
		// A key that is not a card's oracle ID (a token's, an emblem's).
		ownRef("token:goblin", game.AbilitySlotTriggered, label, 0),
	} {
		if got := oracletext.AbilityText(ref); got != "" {
			t.Errorf("%+v: got %q, want no text", ref, got)
		}
	}
}

// TestStackAbilityTextCoverage measures how many of the catalog's own
// triggered and activated rows get printed text, and holds the share
// above a floor so a parser change that stops matching fails loudly.
// CMDCTRL_STACK_TEXT_DUMP=<file> writes every row and its answer.
func TestStackAbilityTextCoverage(t *testing.T) {
	useFixtureSource(t)
	var total, matched int
	var lines []string
	for _, s := range effects.All() {
		def := game.CatalogLookup(s.OracleID)
		if def == nil {
			continue
		}
		for i, row := range def.Triggered {
			total++
			got := oracletext.AbilityText(ownRef(s.OracleID, game.AbilitySlotTriggered, row.Key, i))
			if got != "" {
				matched++
			}
			lines = append(lines, fmt.Sprintf("T | %s | %s | %q", s.Name, row.Key, got))
		}
		for i, row := range def.Activated {
			total++
			got := oracletext.AbilityText(ownRef(s.OracleID, game.AbilitySlotActivated, row.Label, i))
			if got != "" {
				matched++
			}
			lines = append(lines, fmt.Sprintf("A | %s | %s | %q", s.Name, row.Label, got))
		}
	}
	share := float64(matched) / float64(max(total, 1))
	t.Logf("%d of %d catalog ability rows (%.1f%%) show their printed text on the stack", matched, total, 100*share)
	if path := os.Getenv("CMDCTRL_STACK_TEXT_DUMP"); path != "" {
		sort.Strings(lines)
		if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if share < stackTextFloor {
		t.Errorf("only %.1f%% of ability rows map to printed text, under the %.0f%% floor", 100*share, 100*stackTextFloor)
	}
}

// specNamed finds a catalog card's front-face Spec by name.
func specNamed(name string) (effects.Spec, bool) {
	for _, s := range effects.All() {
		if s.Name == name && !strings.Contains(s.OracleID, "#") {
			return s, true
		}
	}
	return effects.Spec{}, false
}

// stackTextFloor is a little under the share measured when #3830 (97.6%)
// landed, so the test fails on a parser that stops matching, not on a
// new card whose labels the matcher cannot place.
const stackTextFloor = 0.95
