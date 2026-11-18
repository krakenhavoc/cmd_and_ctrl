package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const unforgivingOneOracle = "9ec2114c-023b-45c3-a44c-6a9f6462bc85"

func TestUnforgivingOneIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(unforgivingOneOracle)
	if !ok || spec.Name != "Unforgiving One" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

// graveCreatureMV puts a creature card with the given mana cost into p's graveyard.
func graveCreatureMV(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Test", ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func TestUnforgivingOneReturnsACreatureUpToTheModifiedCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := b12Push(g, me.ID, "Unforgiving One", "Creature — Spirit", unforgivingOneOracle, 2, 3)
	// Unforgiving One itself carries a counter, so it is modified; X = 2 with the bear.
	s82SetCounters(g, one, map[string]int{"+1/+1": 1})
	modifiedBear(g, me.ID)
	small := graveCreatureMV(me, "Two Drop", "{1}{B}")
	big := graveCreatureMV(me, "Three Drop", "{2}{B}")

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(one, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	pick := latestPickTarget(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if hasID(pick.PickTargetCards, big) || !hasID(pick.PickTargetCards, small) {
		t.Fatalf("offered = %v, want only the two-drop", pick.PickTargetCards)
	}
	pickCard(t, g, me.ID, small)
	passPriorityAroundTable(t, g)
	if !onBattlefield(g, small) {
		t.Error("the two-drop did not return")
	}
	if !onBattlefield(g, small) || onBattlefield(g, big) {
		t.Error("the three-drop returned")
	}
}

func TestUnforgivingOneWithNothingModifiedOnlyReturnsAZeroDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	one := b12Push(g, me.ID, "Unforgiving One", "Creature — Spirit", unforgivingOneOracle, 2, 3)
	one1 := graveCreatureMV(me, "One Drop", "{B}")
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(one, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	if onBattlefield(g, one1) {
		t.Error("a mana value 1 creature returned with no modified creatures")
	}
}
