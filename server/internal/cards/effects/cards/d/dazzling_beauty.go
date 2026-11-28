package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dazzling Beauty — Instant {2}{W} (Mirage):
//
//	"Cast this spell only during the declare blockers step.
//	 Target unblocked attacking creature becomes blocked. (This spell works
//	 on creatures that can't be blocked.)
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Fog Patch's timing and blocked-with-no-blocker rule for one target
// (CR 509.1h, Game.BecomeBlockedForEffect). The delayed draw (CR 603.7) is
// set up as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "de96ef93-a2e0-409d-9d25-fcb2405039b5",
		Name:         "Dazzling Beauty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent, Draws: 1},
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepDeclareBlockers
		},
		CastConditionLabel: "Cast this spell only during the declare blockers step.",
		Targets:            effects.TargetCreature("target unblocked attacking creature", effects.UnblockedAttacker()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok {
				ctx.Game.BecomeBlockedForEffect([]uuid.UUID{t.ID})
			}
			return effects.DrawACardAtTheNextUpkeep("Dazzling Beauty — draw a card").Apply(ctx)
		},
	})
}
