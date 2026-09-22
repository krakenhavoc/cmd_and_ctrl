package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	_ "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects" // Greenbelt Guardian
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exhaust_view_test.go — the wire half of #1181. `exhausted` is
// present on an activated ability exactly while this OBJECT has
// already activated it, and never on any other ability of the same
// permanent.
//
// It matters that the row is still THERE. An exhausted ability is
// still printed on the permanent, so the menu greys it and says why,
// the way it greys condition_unmet — unlike an ADR 0071 designation
// gate, which makes the ability absent from the view entirely.

const greenbeltGuardianOracle = "2d8aa053-289d-40d9-baa7-9bd1c5b8e957"

func TestExhaustedIsStampedPerAbilityAndOnlyOnTheSpentOne(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	elf := game.NewCard("Greenbelt Guardian", me)
	elf.TypeLine = "Creature — Elf Ranger"
	elf.OracleID = greenbeltGuardianOracle
	elf.Controller = me
	elf.Power, elf.Toughness = 2, 2
	g.Battlefield.PushTop(elf)
	for range 6 {
		forest := game.NewCard("Forest", me)
		forest.TypeLine = "Basic Land — Forest"
		forest.Controller = me
		g.Battlefield.PushTop(forest)
	}

	c := conditionCardView(t, g, elf.InstanceID)
	if len(c.ActivatedAbilities) != 2 {
		t.Fatalf("got %d activated abilities, want 2", len(c.ActivatedAbilities))
	}
	for i, a := range c.ActivatedAbilities {
		if a.Exhausted {
			t.Errorf("ability %d is exhausted before anything was activated", i)
		}
	}
	raw, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(raw), "exhausted") {
		t.Errorf("an unspent exhaust ability must leave the flag off the wire: %s", raw)
	}

	// Index 1 is "Exhaust — {3}{G}: Put three +1/+1 counters on this
	// creature."; index 0 is the repeatable trample pump.
	if err := g.ActivateCatalogAbility(me, elf.InstanceID, 1, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the exhaust ability: %v", err)
	}

	c = conditionCardView(t, g, elf.InstanceID)
	if len(c.ActivatedAbilities) != 2 {
		t.Fatalf("got %d activated abilities after the activation, want 2 — an exhausted ability "+
			"is greyed, not removed", len(c.ActivatedAbilities))
	}
	if c.ActivatedAbilities[0].Exhausted {
		t.Error("the repeatable ability was stamped exhausted — the record is per ability")
	}
	if !c.ActivatedAbilities[1].Exhausted {
		t.Error("the spent exhaust ability is not stamped exhausted")
	}
	raw, _ = json.Marshal(c)
	if got := strings.Count(string(raw), `"exhausted":true`); got != 1 {
		t.Errorf("exhausted on the wire %d times, want exactly 1: %s", got, raw)
	}
}

// A permanent that never spent an exhaust ability and a permanent that
// did are different OBJECTS, so the flag follows the instance and not
// the card (CR 707.2 — what a permanent has done is not copiable).
func TestExhaustedFollowsTheObjectAndNotTheCard(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0].ID
	push := func() uuid.UUID {
		c := game.NewCard("Greenbelt Guardian", me)
		c.TypeLine = "Creature — Elf Ranger"
		c.OracleID = greenbeltGuardianOracle
		c.Controller = me
		c.Power, c.Toughness = 2, 2
		g.Battlefield.PushTop(c)
		return c.InstanceID
	}
	first, second := push(), push()
	for range 6 {
		forest := game.NewCard("Forest", me)
		forest.TypeLine = "Basic Land — Forest"
		forest.Controller = me
		g.Battlefield.PushTop(forest)
	}

	if err := g.ActivateCatalogAbility(me, first, 1, game.ActivateAbilityParams{AutoTap: true}); err != nil {
		t.Fatalf("activate the first Elf's exhaust ability: %v", err)
	}
	if !conditionCardView(t, g, first).ActivatedAbilities[1].Exhausted {
		t.Error("the Elf that activated it is not stamped")
	}
	if conditionCardView(t, g, second).ActivatedAbilities[1].Exhausted {
		t.Error("the other Elf was stamped too — the record is per object")
	}
}
