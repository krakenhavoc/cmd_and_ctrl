package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Verduran Emissary — Creature — Human Wizard {2}{G}, 2/3 (Invasion, S82 OLD
// #5616):
//
//	"Kicker {1}{R} (You may pay an additional {1}{R} as you cast this spell.)
//	 When this creature enters, if it was kicked, destroy target artifact. It
//	 can't be regenerated."
//
// Mold Shambler's shape: an enters trigger that exists only when the spell
// was kicked (CR 702.33d, 603.4), its target chosen as it goes on the stack
// and judged again as it resolves (CR 608.2b), destroying without
// regeneration as Annihilate does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "915bd598-c3dd-4982-bdcc-5904a6f17436",
		Name:          "Verduran Emissary",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedAtLeast(1)),
					"Verduran Emissary — kicked, destroy target artifact; it can't be regenerated",
					func(g *game.Game, item *game.StackItem) error {
						return effects.DestroyTheTargetPermanentNoRegen(item, effects.NewContext(g, item))
					}),
				effects.TargetPermanent("target artifact", effects.Artifact())),
		},
	})
}
