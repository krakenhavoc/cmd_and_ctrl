package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_independence_taplands_test.go — a land that enters tapped is
// not a reactor once it is in play. Its replacement ("This land enters
// tapped", EntryOnly) watches zone moves, so before 2026-10-14 a
// Myriad Landscape or Shipwreck Marsh on the battlefield kept every
// token-making batch asking. The arena replays found it in both rounds.

const aureliaTheLawAboveOracle = "2a800427-ff8c-4b3c-baee-85211b70656d"

// The order check sets an EntryOnly replacement aside on a permanent in
// play, which is only sound if the replacement applies to nothing but its
// own object entering. Every catalog row that declares EntryOnly is
// probed with another card moving onto the battlefield.
func TestEntryOnlyReplacementsApplyOnlyToTheirOwnEntry(t *testing.T) {
	if len(productionDefKeys) == 0 {
		t.Fatal("TestMain did not capture the production catalog keys")
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	seen := 0
	for _, key := range productionDefKeys {
		d := defs[key]
		if d == nil {
			continue
		}
		for i, r := range d.Replacements {
			if !r.EntryOnly {
				continue
			}
			seen++
			if r.AppliesTo == nil {
				t.Errorf("%s replacement %d: EntryOnly with no AppliesTo applies to everything", key, i)
				continue
			}
			src := &game.Card{InstanceID: uuid.New(), Owner: me.ID, Controller: me.ID}
			other := &game.ReplacementEvent{
				Kind: game.RepEventMove, CardID: uuid.New(), NewZone: game.ZoneBattlefield,
			}
			var applies bool
			g.WithWriteLock(func() { applies = r.AppliesTo(other, g, src) })
			if applies {
				t.Errorf("%s replacement %d (%q) is EntryOnly but applies to another card entering", key, i, r.RowLabel)
			}
		}
	}
	if seen == 0 {
		t.Fatal("no EntryOnly replacement in the catalog: the probe reads the wrong registry")
	}
}

// Two Pitiless Plunderers each make a Treasure for one death. A Myriad
// Landscape in play no longer keeps that batch asking.
func TestATaplandInPlayIsNotAReactor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Myriad Landscape", "Land", myriadLandscapeOracle, false)
	pushCatalogPermanent(g, me.ID, "Pitiless Plunderer", "Creature — Human Pirate", pitilessPlundererOracle, false)
	pushCatalogPermanent(g, me.ID, "Pitiless Plunderer", "Creature — Human Pirate", pitilessPlundererOracle, false)
	bear := pushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	treasures := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.Name == "Treasure" {
				n++
			}
		}
		return n
	}
	before := treasures()

	wipe(g, bear)
	if ch := settleUntilOrderPrompt(t, g); ch != nil {
		t.Fatalf("a tapland in play kept the batch asking: %+v", ch)
	}
	settleWithoutPrompts(t, g)
	if got := treasures() - before; got != 2 {
		t.Errorf("made %d Treasures, want 2", got)
	}
}

func TestAureliasTriggersDeclareFootprints(t *testing.T) {
	rows := game.CatalogTriggers(aureliaTheLawAboveOracle)
	if len(rows) != 2 {
		t.Fatalf("%d rows, want 2", len(rows))
	}
	for i, want := range [][]game.FootprintKind{
		{game.FootprintDraw},
		{game.FootprintDamageEachOpponent, game.FootprintGainLife},
	} {
		got := rows[i].Footprint
		if len(got) != len(want) {
			t.Errorf("row %d: footprint %+v, want kinds %v", i, got, want)
			continue
		}
		for j := range want {
			if got[j].Kind != want[j] {
				t.Errorf("row %d step %d: %s, want %s", i, j, got[j].Kind, want[j])
			}
		}
	}
}
