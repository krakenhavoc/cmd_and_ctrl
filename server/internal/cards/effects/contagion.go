package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contagion — Instant {3}{B}{B} (#1664; skipped from #1663 for it):
//
//	"You may pay 1 life and exile a black card from your hand rather
//	 than pay this spell's mana cost.
//	 Distribute two -2/-1 counters among one or two target creatures."
//
// Force of Will's alternative cost in black (Pitch, life 1) and
// Splendid Agony's distribution with a different counter kind. It
// waited on #1664, not on either of those: a -2/-1 counter used to be
// stored on the creature and change nothing, because power and
// toughness read only +1/+1 and -1/-1. Every P/T counter kind now
// counts (game.PTCounterDelta, CR 122.1a), so a 3/3 with one of these
// is a 1/2, and an X/1 with one dies to the toughness check.
//
// A -2/-1 is not a -1/-1: it does not annihilate a +1/+1 counter
// (CR 704.5q names those two kinds only), so a 2/2 with a +1/+1
// counter hit by one is a 1/2 carrying both counters.
func init() {
	Register(Spec{
		OracleID:     "433e86ce-df04-4182-bb63-ad2607d67dbd",
		Name:         "Contagion",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			Pitch(
				"Pay 1 life and exile a black card from your hand",
				1,
				CardInYourHand("a black card from your hand", OfColor("B")),
				"a black card from your hand",
			),
		},
		Targets: TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return PutDividedCounters(ctx, contagionCounter)
		},
	})
}

// contagionCounter is the printed kind — a P/T counter, not a -1/-1.
const contagionCounter = "-2/-1"
