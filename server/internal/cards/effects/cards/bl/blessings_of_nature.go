package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessings of Nature — Sorcery {4}{G} (#1658, unblocked by #1656):
//
//	"Distribute four +1/+1 counters among any number of target
//	 creatures.
//	 Miracle {G} (You may cast this card for its miracle cost when
//	 you draw it if it's the first card you drew this turn.)"
//
// Verdurous Gearhulk's counters shape as a plain sorcery, no "you
// control" restriction.
//
// Miracle {G} is the keyword's own constructor since #1665: revealed
// as the first card drawn in a turn, it is castable for {G} at instant
// speed while the miracle trigger's grant is live (game/miracle.go).
func init() {
	effects.Register(effects.Spec{
		OracleID:         "1e77cf90-ac51-4ac7-b123-be04aabe1688",
		Name:             "Blessings of Nature",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{G}")},
		Targets:          effects.TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(effects.Divide(4)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
