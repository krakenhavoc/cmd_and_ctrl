package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// alaB1Settle passes priority until the stack is empty, putting a batch of
// simultaneous triggers on the stack in the order offered. It stops at any
// other prompt, which the test answers.
func AlaB1Settle(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 48; i++ {
		var order *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Kind == game.PendingChoiceTriggerOrder {
				order = c
			}
		}
		if order != nil {
			if err := g.ResolveTriggerOrder(order.ID, order.Chooser, append([]uuid.UUID(nil), order.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
}

// alaB1CastCreature casts a vanilla creature card of the given size from
// the active player's hand and resolves it, stopping at any prompt its
// arrival raises.
func AlaB1CastCreature(t *testing.T, g *game.Game, name string, power, toughness int) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Beast",
		Power: power, Toughness: toughness, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// alaB1Attack declares attackers against defender and settles the attack
// triggers.
func AlaB1Attack(t *testing.T, g *game.Game, defender uuid.UUID, attackers ...uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range attackers {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	AlaB1Settle(t, g)
}
