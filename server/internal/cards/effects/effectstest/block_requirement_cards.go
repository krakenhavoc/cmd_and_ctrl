package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reqCreature puts a creature onto the battlefield under `owner` through
// the zone-move path, so the layer pass and the catalog see it.
func ReqCreature(g *game.Game, owner uuid.UUID, name, oracle string, p, t int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Test", OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

// reqToBlockers declares `attackers` against the next seat, walks into
// declare_blockers and hands the defender priority, so the next
// PassPriority is the defender's. Returns the defender.
//
// #1501: priority is parked while a defender declares, so the hand-over
// is the defenderHoldsPriority shape rather than the active player's
// pass. The defender's pass is the same CR 509.1c checkpoint as
// finish_blocks (game.blockCheckpointLocked); the game package's
// block_requirements tests pin all three completion paths.
func ReqToBlockers(t *testing.T, g *game.Game, attackers ...uuid.UUID) *game.Player {
	t.Helper()
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%len(g.Seats)]
	AdvanceToDeclareAttackersOf(t, g, seat)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	for g.Turn.Step != game.StepDeclareBlockers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare blockers: %v", err)
		}
	}
	defenderHoldsPriority(t, g, opp.ID)
	return opp
}

// defenderHoldsPriority builds the one shape in which a defender still
// declaring blockers holds priority (#1501: a restore point written
// before priority was parked for the declaration, or a player who became
// a defending player after the step began), so the next PassPriority is
// theirs and runs the declaration's checkpoint. The step must be parked
// for the declaration, which is what entering it with a pending defender
// does now.
func defenderHoldsPriority(t *testing.T, g *game.Game, defender uuid.UUID) {
	t.Helper()
	if g.Turn.PriorityHolder != game.NoPriority {
		t.Fatalf("priority holder %d, want it parked for the declaration", g.Turn.PriorityHolder)
	}
	idx := -1
	for i, s := range g.Seats {
		if s != nil && s.ID == defender {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("no seat for the defender %s", defender)
	}
	g.WithWriteLock(func() { g.Turn.PriorityHolder = idx })
}

// reqRefusal asserts err is a block_requirement refusal and returns it.
func ReqRefusal(t *testing.T, err error) *game.BlockRefusedError {
	t.Helper()
	var br *game.BlockRefusedError
	if !errors.As(err, &br) || br.Reason != game.BlockReasonRequirement {
		t.Fatalf("err = %v, want a block_requirement refusal", err)
	}
	return br
}

func ReqBlock(t *testing.T, g *game.Game, attacker uuid.UUID, blockers ...uuid.UUID) error {
	t.Helper()
	decls := make([]game.BlockDeclaration, 0, len(blockers))
	for _, b := range blockers {
		decls = append(decls, game.BlockDeclaration{Blocker: b, Attacker: attacker})
	}
	return g.DeclareBlockers(decls)
}
