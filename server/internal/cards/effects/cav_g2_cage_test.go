package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grafdigger's Cage sentence one: a creature card in a graveyard can't
// enter the battlefield, so Reanimate leaves it where it was. Without the
// Cage the same line works (the control half).
func TestGrafdiggersCageStopsReanimation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		cage    bool
		entered bool
	}{
		{"with the Cage", true, false},
		{"without the Cage", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			if tc.cage {
				pushCatalogPermanent(g, opp.ID, "Grafdigger's Cage", "Artifact", grafdiggersCageOracle, false)
			}
			victim := seedGraveyardCreature(me, "Bulky Zombie", "{4}{B}")
			castCatalogSpell(t, g, "Reanimate", "Sorcery", reanimateOracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: victim}})
			passPriorityAroundTable(t, g)
			_, onBattlefield := battlefieldCard(g, victim)
			if onBattlefield != tc.entered {
				t.Errorf("on the battlefield = %v, want %v", onBattlefield, tc.entered)
			}
			if !tc.entered && !me.Graveyard.Contains(victim) {
				t.Error("the refused creature should stay in the graveyard")
			}
		})
	}
}

// The Cage names creature cards only, and only graveyards and libraries: a
// noncreature permanent card still comes back from the graveyard.
func TestGrafdiggersCageLeavesNoncreaturesAlone(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, opp.ID, "Grafdigger's Cage", "Artifact", grafdiggersCageOracle, false)
	rock := seedGraveyardCreature(me, "Stone Idol", "{2}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == rock {
				me.Graveyard.Cards[i].TypeLine = "Artifact"
			}
		}
	})
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneBattlefield}, rock); err != nil {
		t.Fatalf("move: %v", err)
	}
	if _, ok := battlefieldCard(g, rock); !ok {
		t.Error("a noncreature card should enter despite the Cage")
	}
}
