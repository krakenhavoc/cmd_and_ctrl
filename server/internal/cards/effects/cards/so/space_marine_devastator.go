package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Space Marine Devastator — Creature — Astartes Warrior {3}{W}, 3/3 (Warhammer
// 40,000 Commander, S82 GAP batch 196, #5948):
//
//	"Squad {2} (As an additional cost to cast this spell, you may pay {2} any
//	 number of times. When this creature enters, create that many tokens that are
//	 copies of it.)
//	 Grav-cannon — When this creature enters, destroy up to one target artifact or
//	 enchantment."
//
// Squad is Wasteland Raider's optional cost and enters trigger; each token copy
// has its own Grav-cannon trigger, so a squad of three destroys up to three
// permanents. The target is chosen as the trigger goes on the stack, "up to one"
// meaning none is legal.
//
// No simplification.
func init() {
	grav := effects.WhenThisEnters("Space Marine Devastator — destroy up to one target artifact or enchantment", effects.GapB196GravCannon)
	grav.Targets = effects.TargetPermanent("up to one target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())).WithCount(0, 1)
	effects.Register(effects.Spec{
		OracleID:      "fa25d142-e6ba-4fc9-81e2-114c32afc030",
		Name:          "Space Marine Devastator",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.SquadMana("{2}", 20)},
		Triggered: []game.TriggeredAbility{
			effects.Squad("Space Marine Devastator"),
			grav,
		},
	})
}
