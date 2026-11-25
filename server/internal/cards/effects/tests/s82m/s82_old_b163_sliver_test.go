package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oldB163SliverOracle = "f9726123-f87b-41bb-9b6f-6d2fbcf0eae2"

func TestOldB163EnduringSliverHasOutlastAndGrantsItToOtherSlivers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	enduring := effectstest.PushCatalogPermanent(g, me.ID, "Enduring Sliver", "Creature — Sliver", oldB163SliverOracle, false)
	mine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Sliver", TypeLine: "Creature — Sliver", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "My Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	theirs := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Sliver", TypeLine: "Creature — Sliver", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	effectstest.Unsick(g, mine)
	effectstest.Unsick(g, bear)
	effectstest.Unsick(g, enduring)
	effectstest.AdvanceToMain(t, g)
	for _, c := range []struct {
		id   uuid.UUID
		name string
		want bool
	}{{enduring, "Enduring Sliver", true}, {mine, "My Sliver", true}, {bear, "My Bear", false}, {theirs, "Their Sliver", false}} {
		effectstest.FillPoolColored(me, "W", 2)
		err := g.ActivateCatalogAbility(me.ID, c.id, 0, game.ActivateAbilityParams{})
		if c.want && err != nil {
			t.Errorf("%s could not outlast: %v", c.name, err)
		}
		if !c.want && err == nil {
			t.Errorf("%s has outlast but should not", c.name)
		}
		if c.want {
			effectstest.PassPriorityAroundTable(t, g)
			if got := effectstest.CountersOn(g, c.id, game.CounterPlusOne); got != 1 {
				t.Errorf("%s has %d +1/+1 counters, want 1", c.name, got)
			}
		}
	}
}

func TestOldB163EnduringSliverOutlastIsSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	enduring := effectstest.PushCatalogPermanent(g, me.ID, "Enduring Sliver", "Creature — Sliver", oldB163SliverOracle, false)
	effectstest.Unsick(g, enduring)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "W", 2)
	if err := g.ActivateCatalogAbility(me.ID, enduring, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("outlast was activated in combat")
	}
}
