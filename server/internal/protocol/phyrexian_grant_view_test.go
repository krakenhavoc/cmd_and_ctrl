package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// phyrexian_grant_view_test.go — ADR 0131 (#2531): `phyrexian_symbols`
// counts the symbols the VIEWER can pay with life, so a {B} under their
// K'rrik is in it, and `phyrexian_granted` says how many of those are
// the grant's rather than printed. An opponent's view of the same card
// text, and a viewer without K'rrik, see neither.

const krrikViewOracle = "test-view-life-for-mana"

func withLifeForManaView(t *testing.T) {
	t.Helper()
	prev := game.CatalogLifeForMana
	game.CatalogLifeForMana = func(key string) []game.LifeForManaStatic {
		if key == krrikViewOracle {
			return []game.LifeForManaStatic{{Color: "B"}}
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogLifeForMana = prev })
}

func seedKrrikView(g *game.Game, owner uuid.UUID) {
	g.Battlefield.PushTop(game.Card{
		InstanceID: uuid.New(), Name: "Test K'rrik", TypeLine: "Legendary Creature",
		OracleID: krrikViewOracle, Owner: owner, Controller: owner,
	})
}

func TestHandCardCountsTheSymbolsAGrantMakesPayableWithLife(t *testing.T) {
	withLifeForManaView(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKrrikView(g, me.ID)
	seen := map[uuid.UUID]bool{me.ID: true, opp.ID: true}

	for _, tc := range []struct {
		name        string
		cost        string
		wantSymbols int
		wantGranted int
	}{
		{"two black", "{1}{B}{B}", 2, 2},
		{"hybrid half", "{B/G}{2}", 1, 1},
		{"printed and granted", "{3}{B/P}{B}", 2, 1},
		{"no black", "{2}{G}", 0, 0},
	} {
		c := game.NewCard(tc.name, me.ID)
		c.TypeLine = "Sorcery"
		c.ManaCost = tc.cost
		c.KnownBy = seen
		me.Hand.PushTop(c)

		v := ViewOfGameFor(g, me.ID.String())
		var got *CardView
		for i := range v.Seats[0].Hand.Cards {
			if v.Seats[0].Hand.Cards[i].InstanceID == c.InstanceID.String() {
				got = &v.Seats[0].Hand.Cards[i]
			}
		}
		if got == nil {
			t.Fatalf("%s: hand view missing the seeded card", tc.name)
		}
		if got.PhyrexianSymbols != tc.wantSymbols || got.PhyrexianGranted != tc.wantGranted {
			t.Errorf("%s (%s): phyrexian_symbols %d / granted %d, want %d / %d",
				tc.name, tc.cost, got.PhyrexianSymbols, got.PhyrexianGranted, tc.wantSymbols, tc.wantGranted)
		}
	}
}

// The grant is the viewer's: the seat without K'rrik sees the printed
// count only.
func TestAViewerWithoutTheGrantSeesOnlyPrintedSymbols(t *testing.T) {
	withLifeForManaView(t)
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedKrrikView(g, opp.ID) // the OPPONENT controls it
	c := game.NewCard("Black Spell", me.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = "{3}{B/P}{B}"
	c.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	me.Hand.PushTop(c)

	v := ViewOfGameFor(g, me.ID.String())
	for i := range v.Seats[0].Hand.Cards {
		got := v.Seats[0].Hand.Cards[i]
		if got.InstanceID != c.InstanceID.String() {
			continue
		}
		if got.PhyrexianSymbols != 1 || got.PhyrexianGranted != 0 {
			t.Errorf("phyrexian_symbols %d / granted %d, want 1 / 0 — the opponent's K'rrik grants nothing here",
				got.PhyrexianSymbols, got.PhyrexianGranted)
		}
		return
	}
	t.Fatal("hand view missing the seeded card")
}

func TestActivatedAbilityViewCountsTheGrant(t *testing.T) {
	withLifeForManaView(t)
	g := buildActiveGame(t)
	owner := g.Seats[0].ID
	seedKrrikView(g, owner)
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: "Test Pod", TypeLine: "Artifact",
		OracleID: "00000000-0000-0000-0000-0000000000dd", Owner: owner, Controller: owner,
		ActivatedAbilities: []game.ActivatedAbilityShape{
			{Label: "{1}{B}{B}: black", Cost: game.AbilityCost{Mana: "{1}{B}{B}"}},
			{Label: "{2}: plain", Cost: game.AbilityCost{Mana: "{2}"}},
		},
	})
	var c CardView
	for _, v := range ViewOfGame(g).Battlefield.Cards {
		if v.InstanceID == id.String() {
			c = v
		}
	}
	if len(c.ActivatedAbilities) != 2 {
		t.Fatalf("got %d abilities, want 2", len(c.ActivatedAbilities))
	}
	if a := c.ActivatedAbilities[0]; a.PhyrexianSymbols != 2 || a.PhyrexianGranted != 2 {
		t.Errorf("{1}{B}{B}: symbols %d / granted %d, want 2 / 2", a.PhyrexianSymbols, a.PhyrexianGranted)
	}
	if a := c.ActivatedAbilities[1]; a.PhyrexianSymbols != 0 || a.PhyrexianGranted != 0 {
		t.Errorf("{2}: symbols %d / granted %d, want 0 / 0", a.PhyrexianSymbols, a.PhyrexianGranted)
	}
}
