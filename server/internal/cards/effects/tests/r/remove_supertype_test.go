package testsr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remove_supertype_test.go — #6199, Melting's "All lands are no longer
// snow" (CR 613.1d, 205.4g) as a printed static.

const meltingTestOracle = "test-6199-melting"

func TestAllLandsAreNoLongerSnow(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: meltingTestOracle,
		Name:     "Test Melting",
		Static:   []game.StaticAbility{effects.AllLandsAreNoLongerSupertype("Snow")},
	})
	g, me, opp := effectstest.P7Table(t)
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Snow-Covered Forest", TypeLine: "Basic Snow Land — Forest"})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Snow-Covered Island", TypeLine: "Basic Snow Land — Island"})
	golem := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Snow Golem", TypeLine: "Snow Artifact Creature — Golem", Power: 1, Toughness: 1})
	snow := func() (bool, bool, bool) {
		var a, b, c bool
		g.ReadSnapshot(func() {
			a = effectstest.FindBattlefieldCardForTest(g, mine).HasSupertype("snow")
			b = effectstest.FindBattlefieldCardForTest(g, theirs).HasSupertype("snow")
			c = effectstest.FindBattlefieldCardForTest(g, golem).HasSupertype("snow")
		})
		return a, b, c
	}
	if a, b, _ := snow(); !a || !b {
		t.Fatal("snow lands are not snow before Melting")
	}
	melting := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Test Melting", TypeLine: "Enchantment", OracleID: meltingTestOracle})
	if a, b, c := snow(); a || b || !c {
		t.Fatalf("under Melting: my land snow=%v, theirs=%v, the Golem=%v; want lands nonsnow, the Golem snow", a, b, c)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, melting); err != nil {
		t.Fatal(err)
	}
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if a, b, _ := snow(); !a || !b {
		t.Fatal("the lands are not snow again once Melting has left")
	}
}
