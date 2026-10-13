package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dodecapod_test.go — ADR 0153 PR 2 (#3168).

const dodecapodOracle = "2c189e3b-90d3-49b4-bebe-a4f14c8a275b"

// An opponent's discard puts it onto the battlefield with two +1/+1
// counters, so it is a 5/5 as it enters.
func TestDodecapodEntersWithTwoCountersWhenAnOpponentMakesYouDiscardIt(t *testing.T) {
	d := newDBOTable(t)
	pod := d.handCard(d.victim, "Dodecapod", "Artifact Creature — Golem", "{4}", dodecapodOracle, 3, 3)
	filler := d.filler(d.victim)

	d.theyDiscard(t, pod, filler)

	c, ok := d.onBattlefield(pod)
	if !ok {
		t.Fatalf("Dodecapod is in %q, want the battlefield", d.zoneOf(pod))
	}
	if n := c.Counters[game.CounterPlusOne]; n != 2 {
		t.Errorf("+1/+1 counters = %d, want 2", n)
	}
	if p := apaLive(d.g, pod).CurrentPower(); p != 5 {
		t.Errorf("power = %d, want 5", p)
	}
	if c.Controller != d.victim.ID {
		t.Errorf("Dodecapod is controlled by %v, want its owner", c.Controller)
	}
}

func TestDodecapodStaysInTheGraveyardForYourOwnDiscard(t *testing.T) {
	d := newDBOTable(t)
	pod := d.handCard(d.caster, "Dodecapod", "Artifact Creature — Golem", "{4}", dodecapodOracle, 3, 3)
	filler := d.filler(d.caster)

	d.iDiscard(t, pod, filler)

	if !d.caster.Graveyard.Contains(pod) {
		t.Fatalf("Dodecapod is in %q, want the graveyard", d.zoneOf(pod))
	}
}
