package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// attack_direction_test.go — ADR 0206 (#4197), the #544 half. Under
// "each player may attack only the nearest opponent in the chosen
// direction" the enumerator offers each attacker only that opponent and
// their planeswalkers, every offered move is accepted, and the withheld
// targets really are refused.
func TestAttackDirectionMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me := g.Seats[s]
	left, across, right := g.Seats[(s+1)%4], g.Seats[(s+2)%4], g.Seats[(s+3)%4]
	clearHand(me)
	bear := enteredCard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == bear {
			c.SummonedThisTurn = false
		}
	}
	leftWalker := battlefieldCard(g, left, game.Card{
		Name: "Left Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	acrossWalker := battlefieldCard(g, across, game.Card{
		Name: "Across Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	g.WithWriteLock(func() {
		g.GrantAttackDirectionForEffect(game.AttackDirectionLeft, "Test Tactician", uuid.Nil, g.UntilYourNextTurnDuration(me.ID))
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	offered := map[string]bool{}
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != bear {
			continue
		}
		offered[decodeAttack(t, m).Target] = true
	}
	if len(offered) != 2 || !offered[left.ID.String()] || !offered[leftWalker.String()] {
		t.Fatalf("offered %v, want only the player to the left and their planeswalker: %v", offered, labels(moves))
	}
	for _, refused := range []uuid.UUID{across.ID, right.ID, acrossWalker} {
		if err := g.Clone().DeclareAttacker(bear, refused); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("engine accepted the withheld attack at %s: %v", refused, err)
		}
	}
}
