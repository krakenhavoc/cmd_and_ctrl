package testss82v

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// An opponent's discard: it sits in the graveyard until the next end
// step, then returns under its owner's control with a +1/+1 counter.
func TestSandGolemReturnsWithACounterAtTheNextEndStep(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	golem := d.HandCard(d.Victim, "Sand Golem", "Artifact Creature — Golem", "{5}", effectstest.SandGolemOracle, 3, 3)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, golem, filler)

	if !d.Victim.Graveyard.Contains(golem) {
		t.Fatalf("before the end step the Golem is in %q, want the graveyard", d.ZoneOf(golem))
	}
	d.ToEndStep(t)

	c, ok := d.OnBattlefield(golem)
	if !ok {
		t.Fatalf("at the end step the Golem is in %q, want the battlefield", d.ZoneOf(golem))
	}
	if c.Controller != d.Victim.ID {
		t.Errorf("the Golem is controlled by %v, want its owner", c.Controller)
	}
	if n := c.Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", n)
	}
	if p := effectstest.ApaLive(d.G, golem).CurrentPower(); p != 4 {
		t.Errorf("power = %d, want 4", p)
	}
}

func TestSandGolemStaysInTheGraveyardWhenYouDiscardItYourself(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	golem := d.HandCard(d.Caster, "Sand Golem", "Artifact Creature — Golem", "{5}", effectstest.SandGolemOracle, 3, 3)
	filler := d.Filler(d.Caster)

	d.IDiscard(t, golem, filler)
	d.ToEndStep(t)

	if !d.Caster.Graveyard.Contains(golem) {
		t.Errorf("the Golem is in %q, want the graveyard", d.ZoneOf(golem))
	}
}

// A Golem that left the graveyard before the end step is not returned.
func TestSandGolemDoesNotReturnFromExile(t *testing.T) {
	d := effectstest.NewDBOTable(t)
	golem := d.HandCard(d.Victim, "Sand Golem", "Artifact Creature — Golem", "{5}", effectstest.SandGolemOracle, 3, 3)
	filler := d.Filler(d.Victim)

	d.TheyDiscard(t, golem, filler)
	d.G.WithWriteLock(func() {
		if err := d.G.ExileCardForEffect(golem); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	d.ToEndStep(t)

	if _, ok := d.OnBattlefield(golem); ok {
		t.Error("the exiled Golem came back")
	}
}
