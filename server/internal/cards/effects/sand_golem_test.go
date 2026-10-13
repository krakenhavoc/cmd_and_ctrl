package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sand_golem_test.go — ADR 0153 PR 2 (#3168).

const sandGolemOracle = "0f9b78ca-1556-4157-a91c-55276fd78127"

// An opponent's discard: it sits in the graveyard until the next end
// step, then returns under its owner's control with a +1/+1 counter.
func TestSandGolemReturnsWithACounterAtTheNextEndStep(t *testing.T) {
	d := newDBOTable(t)
	golem := d.handCard(d.victim, "Sand Golem", "Artifact Creature — Golem", "{5}", sandGolemOracle, 3, 3)
	filler := d.filler(d.victim)

	d.theyDiscard(t, golem, filler)

	if !d.victim.Graveyard.Contains(golem) {
		t.Fatalf("before the end step the Golem is in %q, want the graveyard", d.zoneOf(golem))
	}
	d.toEndStep(t)

	c, ok := d.onBattlefield(golem)
	if !ok {
		t.Fatalf("at the end step the Golem is in %q, want the battlefield", d.zoneOf(golem))
	}
	if c.Controller != d.victim.ID {
		t.Errorf("the Golem is controlled by %v, want its owner", c.Controller)
	}
	if n := c.Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	if p := apaLive(d.g, golem).CurrentPower(); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

func TestSandGolemStaysInTheGraveyardWhenYouDiscardItYourself(t *testing.T) {
	d := newDBOTable(t)
	golem := d.handCard(d.caster, "Sand Golem", "Artifact Creature — Golem", "{5}", sandGolemOracle, 3, 3)
	filler := d.filler(d.caster)

	d.iDiscard(t, golem, filler)
	d.toEndStep(t)

	if !d.caster.Graveyard.Contains(golem) {
		t.Errorf("the Golem is in %q, want the graveyard", d.zoneOf(golem))
	}
}

// A Golem that left the graveyard before the end step is not returned.
func TestSandGolemDoesNotReturnFromExile(t *testing.T) {
	d := newDBOTable(t)
	golem := d.handCard(d.victim, "Sand Golem", "Artifact Creature — Golem", "{5}", sandGolemOracle, 3, 3)
	filler := d.filler(d.victim)

	d.theyDiscard(t, golem, filler)
	d.g.WithWriteLock(func() {
		if err := d.g.ExileCardForEffect(golem); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	d.toEndStep(t)

	if _, ok := d.onBattlefield(golem); ok {
		t.Error("the exiled Golem came back")
	}
}
