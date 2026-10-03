package aiseat_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// multi_block_test.go — the bot half of #1706. A defender with Palace
// Guard (any number) and Two-Headed Giant of Foriys (one extra) against
// three 4/4s: whatever the policy, every enumerated move is accepted,
// the CR 510.1d division prompts a multi-blocker queues are answered by
// the seat that owes them, and the combat ends. The aggressive policy,
// which blocks with everything it is offered, must actually put a
// creature on more than one attacker.

const (
	botPalaceGuardOracle    = "5c92f375-ae6a-4be5-a499-ab87d6bdc49b"
	botTwoHeadedGiantOracle = "38aa31bd-7145-43b9-9409-463d9ad6cd69"
)

// multiBlockPush is limitPush plus the zone move a real entry emits, so
// the layer pass reaches the creature's statics.
func multiBlockPush(g *game.Game, owner *game.Player, name, oracle string, power, tough int) uuid.UUID {
	id := limitPush(g, owner, name, "Creature — Test", oracle, power, tough)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventZoneMove, CardID: id, OldZone: game.ZoneHand, NewZone: game.ZoneBattlefield})
	})
	return id
}

// driveMultiBlockCombat is driveOneCombat that also answers prompts: a
// seat that owes a pending choice (a damage assignment or division)
// acts first. It returns the most attackers any one creature blocked,
// and whether a CR 510.1d division prompt was answered.
func driveMultiBlockCombat(t *testing.T, g *game.Game, pol aiseat.Policy) (most int, divided bool) {
	t.Helper()
	advanceToStep(t, g, game.StepDeclareAttackers)
	for step := 0; step < 200; step++ {
		switch g.Turn.Step {
		case game.StepDeclareAttackers, game.StepDeclareBlockers,
			game.StepFirstStrikeDamage, game.StepCombatDamage, game.StepEndCombat:
		default:
			return most, divided
		}
		for i := range g.Battlefield.Cards {
			most = max(most, len(g.Battlefield.Cards[i].BlockedAttackers()))
		}
		seat := uuid.Nil
		for _, c := range g.PendingChoices {
			if c != nil {
				seat = c.Chooser
				if c.DamageAssignment != nil && c.DamageAssignment.BlockerDivides {
					divided = true
				}
				break
			}
		}
		choosing := seat != uuid.Nil
		declaring := false
		if seat == uuid.Nil && g.Turn.Step == game.StepDeclareBlockers {
			seat = declaringDefender(g)
			declaring = seat != uuid.Nil
		}
		if seat == uuid.Nil {
			if g.Turn.PriorityHolder == game.NoPriority {
				t.Fatalf("step %d: nobody holds priority in %s, and nobody is declaring", step, g.Turn.Step)
			}
			seat = g.Seats[g.Turn.PriorityHolder].ID
		}
		moves := legal.EnumerateFor(g, seat)
		if len(moves) == 0 {
			t.Fatalf("step %d: %s was offered NOTHING in %s — the #544 wedge", step, seat, g.Turn.Step)
		}
		d, err := pol.Decide(context.Background(), aiseat.Input{
			View:  protocol.ViewOfGameFor(g, seat.String()),
			Seat:  seat,
			Moves: moves,
		})
		if err != nil {
			t.Fatalf("step %d: policy: %v", step, err)
		}
		if d.Index == aiseat.Decline {
			if choosing || len(g.PendingChoices) > 0 {
				t.Fatalf("step %d: a seat that owes the table a decision declined in %s", step, g.Turn.Step)
			}
			d.Index = declineAnswer(t, step, declaring, moves)
		}
		if d.Index < 0 || d.Index >= len(moves) {
			t.Fatalf("step %d: index %d of %d", step, d.Index, len(moves))
		}
		mv := moves[d.Index]
		if err := actions.Dispatch(g, actions.Action{
			Type: actions.Type(mv.Type), Player: mv.Player, Caller: seat, Params: mv.Params,
		}); err != nil {
			t.Fatalf("step %d: the engine refused an ENUMERATED move %q: %v — #544", step, mv.Label, err)
		}
	}
	t.Fatalf("the combat did not end in 200 moves (at %s)", g.Turn.Step)
	return most, divided
}

func TestBotsFinishACombatWithMultiBlockers(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 2)
			def := g.Seats[(g.Turn.ActiveSeat+1)%2]
			multiBlockPush(g, def, "Palace Guard", botPalaceGuardOracle, 1, 4)
			multiBlockPush(g, def, "Two-Headed Giant of Foriys", botTwoHeadedGiantOracle, 4, 4)
			most, divided := driveMultiBlockCombat(t, g, pol)
			if name == "aggressive" && (most < 2 || !divided) {
				t.Errorf("the aggressive policy never blocked two attackers with one creature and divided its damage (most %d, divided %v)", most, divided)
			}
		})
	}
}
