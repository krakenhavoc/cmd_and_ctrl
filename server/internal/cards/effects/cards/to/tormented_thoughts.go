package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tormented Thoughts — Sorcery {2}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Target player discards a number of cards equal to the sacrificed
//	 creature's power."
//
// "The power of the sacrificed creature as it last existed on the
// battlefield" (the 2014-04-26 ruling, CR 608.2h), read off the payment
// record (ADR 0113 §1). The target player chooses what to discard, as
// for Mind Rot; zero or negative power discards nothing (CR 107.1b).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "18aee20d-ab73-4fdb-a65c-101beae5fcf5",
		Name:           "Tormented Thoughts",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeCost("a creature", effects.Creature()),
		Targets:        effects.TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := ctx.SacrificedPower()
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetPlayer || n <= 0 {
				return nil
			}
			ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
				Player: item.Targets[0].ID,
				Source: item.SourceCardID,
				N:      n,
			})
			return nil
		},
	})
}
