package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Public Execution — Instant {5}{B} (Conspiracy: Take the Crown, S82 older
// sets batch 273, #5540):
//
//	"Destroy target creature an opponent controls. Each other creature that
//	 player controls gets -2/-0 until end of turn."
//
// "That player" is the target's controller as the spell resolves, read
// before the destruction (CR 608.2h). The set of other creatures is fixed
// as the spell resolves (CR 611.2c) and the -2/-0 lasts until cleanup; the
// destroyed creature itself is not among them. An illegal target at
// resolution makes the whole spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6125f993-970c-469a-b557-23d4dc6684f4",
		Name:         "Public Execution",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.ControlledBy(victim.Controller), effects.NotSelf(id)),
				Power: -2,
				Label: "Public Execution — -2/-0",
			}.Apply(ctx)
		},
	})
}
