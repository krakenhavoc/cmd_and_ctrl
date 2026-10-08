package legal_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// #2109 (ADR 0063's amendment of 2026-10-08), the #544 half. A player
// who "can't attack <player> or permanents they control during their
// next turn" is never offered an attack at that player or their
// planeswalker, every other attack is offered and accepted, and the
// engine refuses what was withheld.
func TestCantAttackPlayerMovesAgreeWithTheEngine(t *testing.T) {
	g := newTable(t)
	s := g.Turn.ActiveSeat
	me, protected, other := g.Seats[s], g.Seats[(s+1)%4], g.Seats[(s+2)%4]
	clearHand(me)
	raider := enteredCard(g, me, game.Card{
		Name: "Raider", TypeLine: "Creature — Test", Power: 3, Toughness: 3,
	})
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.InstanceID == raider {
			c.SummonedThisTurn = false
		}
	}
	walker := battlefieldCard(g, protected, game.Card{
		Name: "Protected Walker", TypeLine: "Legendary Planeswalker — Test",
		Counters: map[string]int{game.CounterLoyalty: 3},
	})
	g.WithWriteLock(func() {
		// The window is open now: start 0, this turn's end.
		me.Statics = append(me.Statics, game.PlayerStatic{
			CantAttack: game.CantAttackGrant{Protected: protected.ID},
			Label:      "Test grant",
			Duration:   g.UntilEndOfTurnDuration(),
		})
	})
	advanceTo(t, g, game.StepDeclareAttackers)

	moves := legal.EnumerateFor(g, me.ID)
	dispatchAll(t, g, me.ID, moves)
	offered := 0
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source != raider {
			continue
		}
		ap := decodeAttack(t, m)
		if ap.Target == protected.ID.String() || ap.Target == walker.String() {
			t.Errorf("offered the creature at the protected player or their planeswalker: %q", m.Label)
		}
		offered++
	}
	if offered == 0 {
		t.Fatalf("no attack offered at all: %v", labels(moves))
	}
	var sawOther bool
	for _, m := range moves {
		if m.Kind == legal.KindAttack && m.Source == raider && decodeAttack(t, m).Target == other.ID.String() {
			sawOther = true
		}
	}
	if !sawOther {
		t.Errorf("the unprotected opponent was not offered: %v", labels(moves))
	}
	for _, refused := range []uuid.UUID{protected.ID, walker} {
		if err := g.Clone().DeclareAttacker(raider, refused); !errors.Is(err, game.ErrIllegalAttackTarget) {
			t.Errorf("engine accepted the withheld attack at %s: %v", refused, err)
		}
	}
}
