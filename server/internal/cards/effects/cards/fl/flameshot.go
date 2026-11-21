package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flameshot — Sorcery {3}{R}:
//
//	"You may discard a Mountain card rather than pay this spell's mana
//	 cost.
//	 Flameshot deals 3 damage divided as you choose among one, two, or
//	 three target creatures."
//
// ADR 0135 §2 (#2412): Snag's discard price with Arc Lightning's
// division, restricted to creatures. Each chosen creature gets at least
// 1 of the 3 (CR 601.2d); one that leaves in response takes nothing and
// its share is lost (CR 608.2b). The Mountain card is discarded as a
// cost, so a countered Flameshot does not give it back.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "26640648-6400-4b80-a78b-f519ab9cd09e",
		Name:             "Flameshot",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.DiscardInstead("a Mountain card", effects.HasSubtype("Mountain"))},
		Targets:          effects.TargetCreature("one, two, or three target creatures").WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DealDividedDamage(ctx)
		},
	})
}
