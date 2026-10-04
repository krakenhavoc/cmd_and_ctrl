package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "18aee20d-ab73-4fdb-a65c-101beae5fcf5",
		Name:           "Tormented Thoughts",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
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
