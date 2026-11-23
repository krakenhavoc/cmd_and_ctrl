package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_shared_helpers.go — test helpers that S82 roll-up 36's tests share with
// the test packages they were first written in (#5948).

// StdB42Attack declares one attacker and locks the declaration in, leaving
// the attack trigger's prompt open.
func StdB42Attack(t *testing.T, g *game.Game, defender, attacker uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	LockInAttacks(t, g)
}

// MatWardPrompt casts Doom Blade at victim as opp and reports whether the
// opponent is asked to pay a ward cost; the spell is then settled either way.
func MatWardPrompt(t *testing.T, g *game.Game, opp *game.Player, victim uuid.UUID) bool {
	t.Helper()
	CastAtWardedCreature(t, g, opp, victim)
	asked := false
	for i := 0; i < 8; i++ {
		if HasPayUnlessFor(g, opp.ID) {
			asked = true
			AnswerPayUnless(t, g, opp.ID, false)
			break
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	PassPriorityAroundTable(t, g)
	return asked
}

// ToNextUpkeep passes the turn and walks the next player's turn to its
// upkeep, where that player is the planar controller (CR 901.6). It
// returns the new active player.
func ToNextUpkeep(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	EndTurn(t, g)
	if g.Turn.Step != game.StepUpkeep {
		AdvanceTo(t, g, game.StepUpkeep)
	}
	return g.Seats[g.Turn.ActiveSeat]
}

// TspB1Sliver seeds a non-catalog Sliver creature that has been under its
// controller's control since the turn began.
func TspB1Sliver(g *game.Game, owner uuid.UUID, name string, power, toughness int) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Sliver",
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// RsConspireTriggers counts the conspire triggers of `spell` on the stack.
func RsConspireTriggers(g *game.Game, spell uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == spell {
			n++
		}
	}
	return n
}
