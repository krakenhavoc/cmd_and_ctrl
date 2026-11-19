package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Escape Detection — Instant {1}{U}{U}:
//
//	"Freerunning—Return a blue creature you control to its owner's hand.
//	 (You may cast this spell for its freerunning cost if you dealt
//	 combat damage to a player this turn with an Assassin or commander.)
//	 Return target creature to its owner's hand.
//	 Draw a card."
//
// The freerunning cost is a bounce, paid at announcement as Daze's is, and
// gated like every freerunning cost on the turn's combat-damage tally. The
// spell itself is Leave in the Dust's bounce-then-draw on a creature.
func init() {
	Register(Spec{
		OracleID:     "e16c5516-f73c-40c4-9ffa-39c6b2e92ab7",
		Name:         "Escape Detection",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		AlternativeCosts: []game.AlternativeCost{
			gapB41FreerunningReturn(
				"Freerunning—Return a blue creature you control to its owner's hand",
				PermanentYouControl("a blue creature you control", Creature(), OfColor("U")),
				"a blue creature you control",
			),
		},
		Targets:   TargetCreature("target creature"),
		OnResolve: bounceTargetThenDrawACard,
	})
}
