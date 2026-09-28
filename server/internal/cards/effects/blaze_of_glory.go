package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blaze of Glory — Instant {W}:
//
//	"Cast this spell only during combat before blockers are declared.
//	 Target creature defending player controls can block any number of
//	 creatures this turn. It blocks each attacking creature this turn if
//	 able."
//
// The timing is a CastCondition: the beginning of combat or the
// declare attackers step (Mandate of Peace's shape, narrowed to the
// two combat steps before the declaration). "Defending player" is
// every opponent of the active player during combat (CR 802.2;
// ControlledByDefendingPlayer), which is what gives it a target in the
// beginning of combat step.
//
// The second and third sentences are one record (#1715): "any number"
// (game.BlockAnyNumberMod) and game.BlockRequirementBlocksEach — one
// CR 509.1c requirement per attacker, each obeyed by blocking that
// attacker. So the target must block every attacker it legally can: a
// flyer it can't reach, or a menace attacker it can't block alone,
// asks nothing (a requirement never beats a restriction), and the
// defending player may add other blockers alongside it. A creature
// that is tapped when blockers are declared can't block, and the
// requirements ask nothing of it.
func init() {
	Register(Spec{
		OracleID:     "b330ac89-790e-4cc9-96a5-532c48252088",
		Name:         "Blaze of Glory",
		Completeness: CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepBeginCombat || g.Turn.Step == game.StepDeclareAttackers
		},
		CastConditionLabel: "Cast this spell only during combat before blockers are declared.",
		Targets:            TargetCreature("target creature defending player controls", ControlledByDefendingPlayer()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			mods := append(blockCapacityMods(0, true), game.AddBlockRequirementMod(game.BlockRequirementBlocksEach))
			return untilEndOfTurn(ctx, ts[0].ID, nil, "Blaze of Glory — blocks each attacking creature if able", mods...)
		},
	})
}
