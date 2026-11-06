package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// conspire_view_test.go — the VIEW half of #4690. A conspire offer's
// tap_others_options are the viewer's untapped creatures that share a
// colour with the spell, off the engine's own walk asked with the spell
// as the cost's source, so the picker offers exactly what CastSpell
// accepts (#544).

const viewConspireOracle = "test-view-conspire"

func TestConspireOfferListsOnlyCreaturesSharingAColor(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stubViewOptionalCosts(t, viewConspireOracle, []game.AdditionalCost{{
		Optional: true, Key: game.ConspireKey, Label: "Conspire",
		TapOthers: &game.TapOthersCost{
			Count: 2, SharesColorWithSource: true,
			Label: "two untapped creatures you control that share a color with it",
			Filter: &game.TargetSpec{Mode: "permanent", Label: "creature",
				Zones: []game.ZoneKind{game.ZoneBattlefield},
				CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
					return c.IsCreature()
				}},
		},
	}})
	spell := uuid.New()
	red, redGreen, green, colorless, tappedRed, theirs := uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New(), uuid.New()
	creature := func(id uuid.UUID, owner uuid.UUID, tapped bool, colors ...string) game.Card {
		return game.Card{InstanceID: id, Name: "Creature", TypeLine: "Creature — Test", Colors: colors,
			Power: 1, Toughness: 1, Owner: owner, Controller: owner, Tapped: tapped}
	}
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: spell, Name: "Red Plot", TypeLine: "Sorcery", OracleID: viewConspireOracle,
			ManaCost: "{1}{R}", Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true},
		})
		g.Battlefield.PushTop(creature(red, me.ID, false, "R"))
		g.Battlefield.PushTop(creature(redGreen, me.ID, false, "R", "G"))
		g.Battlefield.PushTop(creature(green, me.ID, false, "G"))
		g.Battlefield.PushTop(creature(colorless, me.ID, false))
		g.Battlefield.PushTop(creature(tappedRed, me.ID, true, "R"))
		g.Battlefield.PushTop(creature(theirs, opp.ID, false, "R"))
	})

	c := handCardIn(t, ViewOfGameFor(g, me.ID.String()), me.ID.String(), spell.String())
	if len(c.OptionalCosts) != 1 {
		t.Fatalf("optional_costs = %+v, want conspire", c.OptionalCosts)
	}
	oc := c.OptionalCosts[0]
	if oc.TapOthersOptions == nil {
		t.Fatal("no tap_others_options on the conspire offer")
	}
	got := map[string]bool{}
	for _, s := range oc.TapOthersOptions.Cards {
		got[s] = true
	}
	if len(got) != 2 || !got[red.String()] || !got[redGreen.String()] {
		t.Errorf("tap_others_options = %v, want only the two untapped red creatures I control", oc.TapOthersOptions.Cards)
	}
	if oc.TapOthersOptions.Min != 2 || oc.TapOthersOptions.Max != 2 {
		t.Errorf("min/max = %d/%d, want 2/2", oc.TapOthersOptions.Min, oc.TapOthersOptions.Max)
	}
}
