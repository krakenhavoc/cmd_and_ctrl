package testscr

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dodecapod_test.go — ADR 0153 PR 2 (#3168).

const dodecapodOracle = "2c189e3b-90d3-49b4-bebe-a4f14c8a275b"

// An opponent's discard puts it onto the battlefield with two +1/+1
// counters, so it is a 5/5 as it enters.
func TestDodecapodEntersWithTwoCountersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	pod := d.HandCard(d.Victim, "Dodecapod", "Artifact Creature — Golem", "{4}", dodecapodOracle, 3, 3)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, pod, filler)

	c, ok := d.OnBattlefield(pod)
	if !ok {
		t.Fatalf("Dodecapod is in %q, want the battlefield", d.ZoneOf(pod))
	}
	if n := c.Counters[game.CounterPlusOne]; n != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", n)
	}
	if p := effectstest.ApaLive(d.G, pod).CurrentPower(); p != 5 {
		t.Errorf("power = %d, want 5", p)
	}
	if c.Controller != d.Victim.ID {
		t.Errorf("Dodecapod is controlled by %v, want its owner", c.Controller)
	}
}

func TestDodecapodStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	pod := d.HandCard(d.Caster, "Dodecapod", "Artifact Creature — Golem", "{4}", dodecapodOracle, 3, 3)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, pod, filler)

	if !d.Caster.Graveyard.Contains(pod) {
		t.Fatalf("Dodecapod is in %q, want the graveyard", d.ZoneOf(pod))
	}
}
