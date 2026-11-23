package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// seats1684 is the active player and the next seat, the defender every
// test here attacks.
func Seats1684(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// typedCreature is reqCreature with a type line of its own.
func TypedCreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, t int, keywords ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner, Keywords: keywords,
	})
}

// attackAndTrigger declares `attackers` against the next seat and locks
// the declaration in, so the attack triggers are harvested. Returns the
// defender.
func AttackAndTrigger(t *testing.T, g *game.Game, attackers ...uuid.UUID) *game.Player {
	t.Helper()
	seat := g.Turn.ActiveSeat
	opp := g.Seats[(seat+1)%len(g.Seats)]
	AdvanceToDeclareAttackersOf(t, g, seat)
	for _, a := range attackers {
		if err := g.DeclareAttacker(a, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	LockInAttacks(t, g)
	return opp
}

// onToBlockers walks from a settled declare-attackers step into
// declare blockers and hands the defender (the next seat, as
// attackAndTrigger attacks it) priority, so the next PassPriority is
// theirs — defenderHoldsPriority, since #1501 parks priority while a
// defender declares.
func OnToBlockers(t *testing.T, g *game.Game) {
	t.Helper()
	for g.Turn.Step != game.StepDeclareBlockers {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep to declare blockers: %v", err)
		}
	}
	defenderHoldsPriority(t, g, g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID)
}

func BlocksAttackerRecords(g *game.Game) []game.Mod {
	var out []game.Mod
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModAddBlockRequirement && m.Text == string(game.BlockRequirementBlocksAttacker) {
				out = append(out, m)
			}
		}
	}
	return out
}

func MustBlockNow(g *game.Game) map[uuid.UUID]uuid.UUID {
	var out map[uuid.UUID]uuid.UUID
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		out = g.MustBlockForEffect()
	})
	return out
}

func HasEffectiveAbility(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	for _, a := range EffectiveAbilities(t, g, id) {
		if a == kw {
			return true
		}
	}
	return false
}
