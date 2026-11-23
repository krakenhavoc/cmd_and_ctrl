package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Pongify — Instant {U}:
//
//	"Destroy target creature. It can't be regenerated. Its controller
//	creates a 3/3 green Ape creature token."
//
// Blue's one-mana unconditional creature answer, and mechanically
// Beast Within's little brother: destroy, then hand the VICTIM a
// vanilla 3/3. The controller is read BEFORE the destroy for the same
// reason Beast Within reads it first — afterwards the card is in a
// graveyard and its controller field is stale for a permanent that
// had changed hands.
//
// "It can't be regenerated" is ENFORCED as of #667 (CR 701.19c).
// The
// clause is a no-op today rather than a simplification — if a
// regeneration shield ever lands, this card has to be revisited, so
// the note stays.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05849bd6-8f38-4031-be2b-e2aa03beb8cc",
		Name:         "Pongify",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		// #2679: what the target's controller is given back.
		Purpose: game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})), Answers: game.AnswerRemove}, // #4734: destroys a permanent
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: target, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{
				Controller: controller,
				Template:   effects.TokenCard("3/3 green Ape"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
