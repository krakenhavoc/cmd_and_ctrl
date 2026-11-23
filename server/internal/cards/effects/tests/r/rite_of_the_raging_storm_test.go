package testsr

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const riteOfTheRagingStormOracle = "984f92dd-cb9b-4421-a8e6-75bfe566be10"

func TestRiteOfTheRagingStormIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(riteOfTheRagingStormOracle)
	if !ok || spec.Name != "Rite of the Raging Storm" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func TestRiteOfTheRagingStormMakesEachPlayersRagerAndBarsItFromTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.PushPermanentForTest(g, me.ID, "Rite of the Raging Storm", riteOfTheRagingStormOracle, "Enchantment")
	var walker uuid.UUID
	g.WithWriteLock(func() {
		walker = effectstest.PushPermanent(g, me.ID, game.Card{Name: "My Walker", TypeLine: "Legendary Planeswalker — Test",
			Counters: map[string]int{game.CounterLoyalty: 3}})
	})
	bear := effectstest.TdCreature(g, opp.ID, "Plain Bear", 2, 2)

	// The upkeep's player makes the token, not the Rite's controller.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	var rager uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Lightning Rager" {
			rager = c.InstanceID
			if c.Controller != opp.ID || !c.IsToken() || c.Power != 5 || c.Toughness != 1 {
				t.Fatalf("Rager = %+v, want a 5/1 token under the upkeep player", c)
			}
			if !game.HasKeyword(&c, "trample") || !game.HasKeyword(&c, "haste") {
				t.Errorf("Rager lacks trample/haste")
			}
		}
	}
	if rager == uuid.Nil {
		t.Fatal("no Lightning Rager was created")
	}
	if n := effectstest.CountNamed(g, "Lightning Rager"); n != 1 {
		t.Fatalf("%d Ragers, want 1", n)
	}

	effectstest.AdvanceToDeclareAttackersOf(t, g, 1)
	for _, target := range []uuid.UUID{me.ID, walker} {
		err := g.Clone().DeclareAttacker(rager, target)
		var refused *game.AttackTargetRestrictionError
		if !errors.As(err, &refused) {
			t.Fatalf("Rager attacking %s: err = %v, want an attack-target refusal", target, err)
		}
	}
	if err := g.Clone().DeclareAttacker(rager, third.ID); err != nil {
		t.Fatalf("Rager attacking another player: %v", err)
	}
	// Only creatures named Lightning Rager are barred.
	if err := g.Clone().DeclareAttacker(bear, me.ID); err != nil {
		t.Fatalf("a Bear attacking the controller: %v", err)
	}

	// "At the beginning of the end step, sacrifice this creature."
	effectstest.AdvanceToEndStepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rager) {
		t.Error("the Rager survived the end step")
	}
}
