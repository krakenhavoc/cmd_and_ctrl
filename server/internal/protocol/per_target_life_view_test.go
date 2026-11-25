package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// per_target_life_view_test.go — #6206, the view's half of "This spell
// costs 3 life more to cast for each target" (Phyrexian Purge, CR
// 601.2f, 119.4).
//
// The same stamp a permanent's target tax gets (target_tax_view_test.go),
// for a cost the SPELL charges itself: every candidate carries what
// naming it costs, so the picker can say "3 life" before the click that
// commits the cast, and a candidate the viewer cannot pay for is named
// as such.

func TestLegalTargetsCarryTheSpellsOwnPerTargetLife(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const purge = "test-6206-view-purge"

	defs := map[string]*game.CardDef{
		purge: {LifePerTarget: game.LifePerTargetCost{
			Life:  3,
			Label: "This spell costs 3 life more to cast for each target.",
		}},
	}
	prevDef, prevSpec := game.CatalogLookup, game.CatalogTargetSpec
	game.CatalogLookup = func(key string) *game.CardDef { return defs[key] }
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != purge {
			return nil
		}
		return &game.TargetSpec{
			Mode: "creature", Label: "any number of target creatures",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature()
			},
			Min: 1, Max: 0,
		}
	}
	t.Cleanup(func() { game.CatalogLookup, game.CatalogTargetSpec = prevDef, prevSpec })

	bear := game.NewCard("Their Bear", opp.ID)
	bear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(bear)
	spell := game.NewCard("Test Purge", me.ID)
	spell.TypeLine = "Sorcery"
	spell.OracleID = purge
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	find := func(t *testing.T) *LegalTargetsView {
		t.Helper()
		view := ViewOfGameFor(g, me.ID.String())
		for i := range view.Seats[0].Hand.Cards {
			c := &view.Seats[0].Hand.Cards[i]
			if c.InstanceID == spell.InstanceID.String() {
				if c.LegalTargets == nil {
					t.Fatal("the spell ships its legal targets")
				}
				return c.LegalTargets
			}
		}
		t.Fatal("the spell is in the viewer's hand")
		return nil
	}

	me.Life = 20
	lt := find(t)
	if got := lt.LifeTaxes[bear.InstanceID.String()]; got != 3 {
		t.Errorf("naming the Bear costs %d life, want 3 (%+v)", got, lt.LifeTaxes)
	}
	if len(lt.LifeTaxesUnpayable) != 0 {
		t.Errorf("at 20 life it is payable: %v", lt.LifeTaxesUnpayable)
	}

	// CR 119.4: at 2 life the seat cannot pay the 3, so the candidate is
	// still offered and marked unpayable — the picker greys what the
	// engine would refuse.
	me.Life = 2
	lt = find(t)
	if len(lt.LifeTaxesUnpayable) != 1 || lt.LifeTaxesUnpayable[0] != bear.InstanceID.String() {
		t.Errorf("at 2 life the 3 is unpayable: %v", lt.LifeTaxesUnpayable)
	}
}
