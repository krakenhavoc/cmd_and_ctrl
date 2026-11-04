package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oroku Saki, Shredder Rising — Legendary Creature — Human Ninja {2}{B},
// 3/1 (Teenage Mutant Ninja Turtles, #4457):
//
//	"Sneak {1}{B} (You may cast this spell for {1}{B} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step. He enters tapped and attacking.)
//	 Whenever Oroku Saki deals combat damage to a player, you draw a card
//	 and lose 1 life."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The trigger
// is Snooping Page's: draw, then lose 1 life, in the printed order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "d14f620c-0fbd-485c-9fe4-34949e742147",
		Name:             "Oroku Saki, Shredder Rising",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Sneak("{1}{B}")},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Oroku Saki — you draw a card and lose 1 life",
				Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}
