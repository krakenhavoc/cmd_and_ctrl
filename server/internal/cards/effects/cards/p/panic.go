package cardsp

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Panic — Instant {R} (Ice Age, S82 OLD batch 837, #5636):
//
//	"Cast this spell only during combat before blockers are declared.
//	 Target creature can't block this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Blaze of Glory's timing (the beginning of combat or the declare attackers
// step), Ahn-Crop Crasher's can't-block restriction until end of turn, then the
// delayed draw (CR 603.7a, next_upkeep_draw.go). With its one target gone the
// spell does not resolve, so nothing is drawn (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e98429a7-a83f-4b69-afad-9eb68b09e38e",
		Name:         "Panic",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepBeginCombat || g.Turn.Step == game.StepDeclareAttackers
		},
		CastConditionLabel: "Cast this spell only during combat before blockers are declared.",
		Targets:            effects.TargetCreature("target creature"),
		Purpose:            effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.TargetCreatureCantBlockThisTurn("Panic — can't block this turn")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Panic — draw a card").Apply(ctx)
		},
	})
}
