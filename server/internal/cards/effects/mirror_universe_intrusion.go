package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mirror Universe Intrusion — Sorcery {1}{U}{B} (Star Trek, #4973):
//
//	"Kicker {4}{U}{B}
//	 Create a token that's a copy of target creature an opponent
//	 controls, except it has menace. If this spell was kicked, instead
//	 for each creature target opponent controls, create a token that's a
//	 copy of that creature, except it has menace."
//
// The kicked clause replaces the printed one, target and all (Tear
// Asunder's WhenPaid, #1716): an unkicked cast targets a creature an
// opponent controls, a kicked one targets an opponent instead, each judged
// at announce and again at resolution (CR 608.2b).
//
// Unkicked, the copy is one CreateTokenCopy with the "except it has
// menace" exception applied to the copiable values (CR 707.9b). Kicked,
// the creatures that player controls are read as the spell resolves, and
// every copy is made by one creation (CreateTokenCopiesTogether), so they
// enter at the same time (CR 608.2f). The tokens are the caster's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "007d08e8-4db2-47d1-ba39-8bb8efa491a1",
		Name:         "Mirror Universe Intrusion",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{WhenPaid(Kicker("{4}{U}{B}"),
			TargetPlayer("target opponent", Opponent()))},
		Targets:   TargetCreature("target creature an opponent controls", OpponentControls()),
		OnResolve: mirrorUniverseIntrusionResolve,
	})
}

func mirrorUniverseIntrusionResolve(item *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	if !ctx.WasKicked() {
		if t.Kind != game.TargetCard {
			return nil
		}
		return CreateTokenCopy{Controller: item.Controller, Copy: t.ID, N: 1, Except: trekB25ExceptItHasMenace}.Apply(ctx)
	}
	if t.Kind != game.TargetPlayer {
		return nil
	}
	var copies []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == t.ID && c.IsCreature() {
			copies = append(copies, c.InstanceID)
		}
	}
	if len(copies) == 0 {
		return nil
	}
	return CreateTokenCopiesTogether{Controller: item.Controller, Copies: copies, Except: trekB25ExceptItHasMenace}.Apply(ctx)
}
