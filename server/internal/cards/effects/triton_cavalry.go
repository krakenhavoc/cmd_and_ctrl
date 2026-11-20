package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "0fd11de4-2dfd-4345-b3cb-d7d65e11c6d9",
		Name:         "Triton Cavalry",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					heroic("Triton Cavalry — you may return target enchantment to its owner's hand (heroic)", bounceChosenTarget),
					"Triton Cavalry — return target enchantment to its owner's hand?"),
				TargetPermanent("target enchantment", Enchantment()),
			),
		},
	})
}
