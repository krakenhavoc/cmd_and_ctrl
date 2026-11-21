package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triton Cavalry — Creature — Merfolk Soldier {3}{U}, 2/4 (Journey into Nyx, #5558):
//
//	"Heroic — Whenever you cast a spell that targets this creature, you
//	 may return target enchantment to its owner's hand."
//
// Elite Skirmisher's optional targeted heroic, returning an enchantment to
// its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fd11de4-2dfd-4345-b3cb-d7d65e11c6d9",
		Name:         "Triton Cavalry",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.Heroic("Triton Cavalry — you may return target enchantment to its owner's hand (heroic)", effects.BounceChosenTarget),
					"Triton Cavalry — return target enchantment to its owner's hand?"),
				effects.TargetPermanent("target enchantment", effects.Enchantment()),
			),
		},
	})
}
