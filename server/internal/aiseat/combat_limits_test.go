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

// combat_limits_test.go — the bot half of #1507: a CR 508.1c / 509.1b
// whole-combat limit (Silent Arbiter, Crawlspace) must never wedge a
// bot table. The enumerator is the bot's whole world, so the two ways
// this could go wrong are the #544 pair: an attack or block OFFERED
// that the verb refuses (the seat re-picks the same argmax forever),
// or a seat left with nothing to do.
//
// Each test drives one combat move by move through legal.EnumerateFor
// and actions.Dispatch, with two policies: the heuristic, and a
// scripted policy that attacks and blocks with everything it is
// offered — the policy most likely to walk into a limit.
//
// Deliberately NOT behind AISEAT_GAME_TESTS: no whole game, a bounded
// loop over one combat, in milliseconds.

const (
	botSilentArbiterOracle = "1cdf30de-d88c-421a-80df-4917bbd2f09e"
	botCrawlspaceOracle    = "2296370c-fe34-4df6-92a5-260f1634bede"
)

// limitTable is a started table of `seats` past the mulligans, parked
// at the active seat's declare attackers, with three 4/4s for the
// active seat and two 2/2s for every other seat.
func limitTable(t *testing.T, seats int) *game.Game {
	t.Helper()
	g := newRoom(t, seats, 41).Game
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	active := g.Seats[g.Turn.ActiveSeat]
	for i := 0; i < 3; i++ {
		limitPush(g, active, "Ogre", "Creature — Ogre", "", 4, 4)
	}
	for _, p := range g.Seats {
		if p.ID == active.ID {
			continue
		}
		limitPush(g, p, "Bear", "Creature — Bear", "", 2, 2)
		limitPush(g, p, "Bear", "Creature — Bear", "", 2, 2)
	}
	return g
}

func limitPush(g *game.Game, owner *game.Player, name, typeLine, oracle string, power, tough int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(),
		Name:       name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: tough,
		Owner: owner.ID, Controller: owner.ID,
	}
	c.AddKnowersAll(seatIDs(g))
	g.Battlefield.PushTop(c)
	return c.InstanceID
}

func hasKind(moves []legal.Move, k legal.Kind) bool {
	for _, m := range moves {
		if m.Kind == k {
			return true
		}
	}
	return false
}

// combatTally is the most creatures seen attacking, attacking one
// target, blocking, and blocking for one seat, at any point of the
// drive.
type combatTally struct {
	attackers, blockers int
	perDefender         map[uuid.UUID]int
	blockersBy          map[uuid.UUID]int
}

// driveOneCombat plays from declare attackers until the cursor leaves
// combat. The acting seat is a defender that still has a block to make
// in declare blockers, and otherwise the priority holder. Fails on an
// empty move list, a decline by the priority holder, a refused
// dispatch, or a combat that does not end.
func driveOneCombat(t *testing.T, g *game.Game, pol aiseat.Policy) combatTally {
	t.Helper()
	return driveOneCombatWatching(t, g, pol, nil)
}

// driveOneCombatWatching is driveOneCombat with a hook run before every
// move (#1571 watches a goaded creature's target through it).
func driveOneCombatWatching(t *testing.T, g *game.Game, pol aiseat.Policy, watch func()) combatTally {
	t.Helper()
	advanceToStep(t, g, game.StepDeclareAttackers)
	tally := combatTally{perDefender: map[uuid.UUID]int{}, blockersBy: map[uuid.UUID]int{}}
	declined := map[uuid.UUID]bool{}
	for step := 0; step < 120; step++ {
		switch g.Turn.Step {
		case game.StepDeclareAttackers, game.StepDeclareBlockers,
			game.StepFirstStrikeDamage, game.StepCombatDamage, game.StepEndCombat:
		default:
			return tally
		}
		observeCombat(g, &tally)
		if watch != nil {
			watch()
		}

		seat := uuid.Nil
		if g.Turn.Step == game.StepDeclareBlockers {
			for _, p := range g.Seats {
				if declined[p.ID] || p.ID == g.Seats[g.Turn.ActiveSeat].ID {
					continue
				}
				if hasKind(legal.EnumerateFor(g, p.ID), legal.KindBlock) {
					seat = p.ID
					break
				}
			}
		}
		if seat == uuid.Nil {
			if g.Turn.PriorityHolder == game.NoPriority {
				t.Fatalf("step %d: nobody holds priority in %s", step, g.Turn.Step)
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
			if seat == g.Seats[g.Turn.PriorityHolder].ID {
				t.Fatalf("step %d: the priority holder declined in %s", step, g.Turn.Step)
			}
			declined[seat] = true
			continue
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
	t.Fatalf("the combat did not end in 120 moves (at %s)", g.Turn.Step)
	return tally
}

func observeCombat(g *game.Game, tally *combatTally) {
	attackers, blockers := 0, 0
	per, by := map[uuid.UUID]int{}, map[uuid.UUID]int{}
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.AttackingTarget != uuid.Nil {
			attackers++
			per[c.AttackingTarget]++
		}
		if c.BlockingTarget != uuid.Nil {
			blockers++
			by[c.Controller]++
		}
	}
	tally.attackers = max(tally.attackers, attackers)
	tally.blockers = max(tally.blockers, blockers)
	for d, n := range per {
		tally.perDefender[d] = max(tally.perDefender[d], n)
	}
	for s, n := range by {
		tally.blockersBy[s] = max(tally.blockersBy[s], n)
	}
}

// aggressive attacks and blocks with everything it is offered, then
// passes.
func aggressive() aiseat.Policy {
	return &scripted{prefer: []string{"Attack", "Block"}}
}

// TestBotsFinishACombatUnderSilentArbiter — one attacker and one
// blocker, whatever the policy wants, and the combat ends.
func TestBotsFinishACombatUnderSilentArbiter(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 2)
			limitPush(g, g.Seats[1], "Silent Arbiter", "Artifact Creature — Construct", botSilentArbiterOracle, 1, 5)
			tally := driveOneCombat(t, g, pol)
			if tally.attackers > 1 || tally.blockers > 1 {
				t.Errorf("%d attackers and %d blockers under Silent Arbiter", tally.attackers, tally.blockers)
			}
			if name == "aggressive" && (tally.attackers != 1 || tally.blockers != 1) {
				t.Errorf("the aggressive policy should have used the one attack and the one block: %+v", tally)
			}
		})
	}
}

// TestBotsFinishACombatUnderCrawlspace — four seats, Crawlspace on one
// of them: at most two creatures attack that seat, the aggressive
// policy spends its third creature elsewhere, and the combat ends.
func TestBotsFinishACombatUnderCrawlspace(t *testing.T) {
	for name, pol := range map[string]aiseat.Policy{"heuristic": heuristic.New(), "aggressive": aggressive()} {
		t.Run(name, func(t *testing.T) {
			g := limitTable(t, 4)
			crawl := g.Seats[(g.Turn.ActiveSeat+1)%4]
			limitPush(g, crawl, "Crawlspace", "Artifact", botCrawlspaceOracle, 0, 0)
			tally := driveOneCombat(t, g, pol)
			if n := tally.perDefender[crawl.ID]; n > 2 {
				t.Errorf("%d creatures attacked the Crawlspace seat", n)
			}
			if name == "aggressive" && tally.attackers != 3 {
				t.Errorf("the aggressive policy should still attack with all three: %+v", tally)
			}
		})
	}
}
