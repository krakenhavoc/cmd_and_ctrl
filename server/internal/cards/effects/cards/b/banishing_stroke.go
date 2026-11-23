package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Banishing Stroke — Instant {5}{W} (#1665):
//
//	"Put target artifact, creature, or enchantment on the bottom of its
//	 owner's library.
//	 Miracle {W} (You may cast this card for its miracle cost when you
//	 draw it if it's the first card you drew this turn.)"
//
// A tuck rather than a destroy, so indestructible and regeneration do
// not help and nothing dies. The move is the tuck route, so a commander
// is offered the command zone instead (CR 903.9).
func init() {
	effects.Register(effects.Spec{
		OracleID:         "a6898364-c29e-4b97-a500-344efa3ec24a",
		Name:             "Banishing Stroke",
		Completeness:     effects.CompletenessFull,
		Targets:          effects.TargetPermanent("target artifact, creature, or enchantment", effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
		AlternativeCosts: []game.AlternativeCost{effects.Miracle("{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			return effects.PutIntoLibrary{Card: targets[0].ID, ToBottom: true}.Apply(ctx)
		},
	})
}
