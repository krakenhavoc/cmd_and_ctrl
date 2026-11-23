package testsl

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// lander_not_a_land_test.go — #5213. A Lander token ("Token Artifact —
// Lander") is an artifact: its entering is not a land entering, so it
// fires no landfall trigger and does not count toward "lands that
// entered this turn". A real land still does both.

func TestALanderTokenDoesNotTriggerLandfall(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Tireless Provisioner", "Creature — Elf Scout", effectstest.TirelessProvisionerOracle, false)
	base := g.TurnTallyFor(me.ID).LandsEntered

	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(me.ID, effects.B16LanderToken(), 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Lander"); got != 1 {
		t.Fatalf("%d Landers, want 1", got)
	}
	if c := effectstest.LatestOptionPickFor(g, me.ID); c != nil {
		t.Fatalf("a Lander token entering fired landfall: %+v", c)
	}
	if got := g.TurnTallyFor(me.ID).LandsEntered; got != base {
		t.Errorf("a Lander token moved the lands-entered tally %d -> %d", base, got)
	}
	if c, ok := g.LookupCardForEffect(landerTokenID(t, g)); !ok || c.IsLand() || !c.IsArtifact() || !c.HasSubtype("Lander") {
		t.Errorf("the Lander token on the battlefield (%q): want an artifact with subtype Lander that is not a land", c.TypeLine)
	}

	// A real land still triggers, and still counts.
	effectstest.PlayLandFromHand(t, g, "Forest", "")
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LatestOptionPickFor(g, me.ID); c == nil || len(c.PickOptions) != 2 {
		t.Fatalf("a Forest entering should fire landfall's Treasure-or-Food choice: %+v", c)
	}
	if got := g.TurnTallyFor(me.ID).LandsEntered; got != base+1 {
		t.Errorf("a Forest moved the lands-entered tally %d -> %d, want %d", base, got, base+1)
	}
}

// landerTokenID is the InstanceID of the Lander on the battlefield.
func landerTokenID(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Lander" {
			return c.InstanceID
		}
	}
	t.Fatal("no Lander on the battlefield")
	return uuid.Nil
}
