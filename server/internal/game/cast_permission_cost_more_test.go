package game

import "testing"

// CastPermission.CostMore (Zero, Incorporeal Sensate, #4973): a spell cast
// through the permission costs {N} more, the increase survives a snapshot,
// and a restored table prices the cast the same way.
func TestCostMorePermissionSurvivesASnapshotAndPricesTheCast(t *testing.T) {
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	source, id := exiledWithSetup(t, g, Card{Name: "Boxed Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}"})
	g.WithWriteLock(func() {
		g.GrantCastPermissionForEffect(CastPermission{
			Player:   me.ID,
			Zone:     ZoneExile,
			Scope:    ScopeStanding,
			Filter:   PermissionFilter{ExiledWithSource: true, ExiledWith: source, OwnedByHolder: true},
			CostMore: 2,
		})
	})

	price := func(g *Game) string {
		var out string
		g.ReadSnapshot(func() {
			c, ok := hiddenZoneCard(g.Exile, id)
			if !ok {
				t.Fatal("card not in exile")
			}
			p, err := g.PriceCastForEffect(me.ID, c, CastSpellParams{FromZone: "exile"})
			if err != nil {
				t.Fatal(err)
			}
			out = p.Total.String()
		})
		return out
	}
	if got := price(g); got != "{3}{G}" {
		t.Fatalf("price = %s, want {3}{G}", got)
	}
	_, restored := roundTrip(t, g)
	if got := price(restored); got != "{3}{G}" {
		t.Errorf("restored price = %s, want {3}{G}", got)
	}
}
