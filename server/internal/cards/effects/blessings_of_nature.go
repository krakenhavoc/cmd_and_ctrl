package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// DECLARED SIMPLIFICATION: Miracle has no engine shape yet — unlike
// Madness (Spec.Madness, used by Violent Eruption in this same batch),
// there is no draw-triggered alternative-cost offer, so casting for
// {G} on the reveal is not offered. Weaker than printed, never
// stronger: the counters half is complete, cast normally for {4}{G}.
func init() {
	Register(Spec{
		OracleID:     "1e77cf90-ac51-4ac7-b123-be04aabe1688",
		Name:         "Blessings of Nature",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Miracle isn't implemented — this can only be cast normally for {4}{G}, never for its {G} miracle cost.",
		},
		Targets: TargetCreature("any number of target creatures").WithCount(0, 0).Dividing(Divide(4)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
