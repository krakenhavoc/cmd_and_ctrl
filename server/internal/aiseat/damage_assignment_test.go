package aiseat_test

import (
	"context"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// TestHeuristicAssignsDamageToKillTheOneItCan is review game 2's seq
// 325 (#2692): the heuristic's 3/3, blocked by a 2/4 and then a 1/1,
// kills the 1/1 instead of putting all 3 on the 2/4.
func TestHeuristicAssignsDamageToKillTheOneItCan(t *testing.T) {
	g := newRoom(t, 2, 41).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	def := g.Seats[(g.Turn.ActiveSeat+1)%2]
	mary := limitPush(g, me, "Mary", "Creature — Pirate", "", 3, 3)
	yshtola := limitPush(g, def, "Y'shtola", "Creature — Cat Wizard", "", 2, 4)
	soldier := limitPush(g, def, "Soldier", "Creature — Soldier", "", 1, 1)
	advanceToStep(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(mary, def.ID); err != nil {
		t.Fatal(err)
	}
	advanceToStep(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlockers([]game.BlockDeclaration{
		{Blocker: yshtola, Attacker: mary},
		{Blocker: soldier, Attacker: mary},
	}); err != nil {
		t.Fatal(err)
	}
	advanceToStep(t, g, game.StepCombatDamage)
	if len(g.PendingChoices) == 0 || g.PendingChoices[0].Kind != game.PendingChoiceDamageAssignment {
		t.Fatalf("no damage-assignment prompt: %v", g.PendingChoices)
	}
	moves := legal.EnumerateFor(g, me.ID)
	d, err := heuristic.New().Decide(context.Background(), aiseat.Input{
		View:  protocol.ViewOfGameFor(g, me.ID.String()),
		Seat:  me.ID,
		Moves: moves,
	})
	if err != nil || d.Index < 0 || d.Index >= len(moves) {
		t.Fatalf("Decide = %+v, %v over %d moves", d, err, len(moves))
	}
	mv := moves[d.Index]
	if mv.Type != legal.TypeResolveChoice {
		t.Fatalf("the heuristic picked %q, want the damage assignment", mv.Label)
	}
	if err := actions.Dispatch(g, actions.Action{
		Type: actions.Type(mv.Type), Player: mv.Player, Caller: me.ID, Params: mv.Params,
	}); err != nil {
		t.Fatalf("the engine refused %q: %v", mv.Label, err)
	}
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == soldier {
			t.Fatalf("the 1/1 survived: %s", mv.Params)
		}
	}
}
