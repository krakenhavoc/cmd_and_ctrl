package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// ninjutsu_test.go — the enumerator half of #1227. The invariant is
// #544's, read through a cost whose candidate pool is a COMBAT fact:
// offer the activation exactly when an unblocked attacker exists, and
// offer none at all when the board cannot pay — so a bot is never
// handed a ninjutsu move the engine then refuses.

// ninjutsuCard is the keyword's shape without a catalog import: a hand
// ability whose cost is mana plus "return an unblocked attacker you
// control to hand".
func ninjutsuCard(name, cost string) game.Card {
	return game.Card{
		Name:     name,
		TypeLine: "Creature — Human Ninja",
		Power:    2, Toughness: 2,
		ActivatedAbilities: []game.ActivatedAbilityShape{{
			Label: "Ninjutsu " + cost,
			Cost: game.AbilityCost{
				Mana: cost,
				ReturnToHand: &game.ReturnToHandCost{
					Count: 1,
					Filter: &game.TargetSpec{
						Mode:  "permanent",
						Label: "an unblocked attacker you control",
						Zones: []game.ZoneKind{game.ZoneBattlefield},
						CardOK: func(g *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
							return c.IsCreature() && g.UnblockedAttackerForEffect(c.InstanceID)
						},
						Min: 1, Max: 1,
					},
					Label: "an unblocked attacker you control",
				},
			},
			Zones:  []game.ZoneKind{game.ZoneHand},
			Effect: func(*game.Game, *game.StackItem) error { return nil },
		}},
	}
}

// vanillaAttacker is a creature with no text, the thing the cost pays
// with.
func vanillaAttacker(name string) game.Card {
	return game.Card{Name: name, TypeLine: "Creature — Rat", Power: 1, Toughness: 1}
}

// The whole enumerator contract in one test: nothing offered in the
// declare-attackers step (no creature is unblocked yet), exactly one
// move once the cursor reaches declare blockers, and the move the
// enumerator offers is one the dispatcher accepts.
func TestNinjutsuIsEnumeratedOnlyOnceBlockersAreDeclared(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island2", "Island"))
	attacker := battlefieldCard(g, active, vanillaAttacker("Sneaky Rat"))
	ninja := handCard(active, ninjutsuCard("Ninja of the Deep Hours", "{1}{U}"))

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if acts := activationsOf(legal.EnumerateFor(g, active.ID), ninja); len(acts) != 0 {
		t.Fatalf("ninjutsu offered in declare_attackers, where nothing is unblocked yet: %v", labels(acts))
	}

	advanceTo(t, g, game.StepDeclareBlockers)
	moves := legal.EnumerateFor(g, active.ID)
	acts := activationsOf(moves, ninja)
	if len(acts) != 1 {
		t.Fatalf("want exactly one ninjutsu move once blockers are declared, got %v", labels(acts))
	}
	dispatchAll(t, g, active.ID, moves)
}

// A BLOCKED attacker is not a payment, so the ability is not offered
// at all — CR 118.3 read through the enumerator (#544).
func TestNinjutsuIsNotEnumeratedForABlockedAttacker(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[1]
	clearHand(active)
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island2", "Island"))
	attacker := battlefieldCard(g, active, vanillaAttacker("Sneaky Rat"))
	wall := battlefieldCard(g, defender, game.Card{
		Name: "Wall", TypeLine: "Creature — Wall", Power: 0, Toughness: 4,
	})
	ninja := handCard(active, ninjutsuCard("Ninja of the Deep Hours", "{1}{U}"))

	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}

	if acts := activationsOf(legal.EnumerateFor(g, active.ID), ninja); len(acts) != 0 {
		t.Fatalf("ninjutsu offered with the only attacker blocked: %v", labels(acts))
	}
}
