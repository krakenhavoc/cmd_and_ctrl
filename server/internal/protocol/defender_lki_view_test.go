package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TestAttackerWhoseWalkerLeftKeepsItsDefendingPlayer — #1364. CR
// 506.4c: a creature whose planeswalker has left combat is still
// attacking and "may be blocked", by the walker's former controller
// (CR 802.2a). The wire names that seat, so the client's block pickers
// offer the block the server now accepts; the kind is absent because
// the creature attacks nothing.
func TestAttackerWhoseWalkerLeftKeepsItsDefendingPlayer(t *testing.T) {
	g := buildActiveGame(t)
	owner, other := g.Seats[0].ID, g.Seats[1].ID
	walker := seatWalker(g, other, 4)
	attacker := uuid.New()
	g.WithWriteLock(func() {
		g.Battlefield.PushTop(game.Card{
			InstanceID: attacker, Name: "Attacker", TypeLine: "Creature — Test",
			Power: 2, Toughness: 2, Owner: owner, Controller: owner,
		})
	})
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	var err error
	g.WithWriteLock(func() { err = g.DestroyPermanentForEffect(walker) })
	if err != nil {
		t.Fatalf("destroy the walker: %v", err)
	}

	c := cardInView(t, ViewOfGame(g), attacker)
	if c.DefendingPlayer != other.String() {
		t.Errorf("defending_player = %q, want the walker's former controller", c.DefendingPlayer)
	}
	if c.AttackingTargetKind != "" {
		t.Errorf("attacking_target_kind = %q, want absent — it attacks nothing (CR 506.4c)", c.AttackingTargetKind)
	}
}
