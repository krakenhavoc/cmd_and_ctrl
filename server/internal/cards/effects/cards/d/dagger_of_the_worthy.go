package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const daggerOfTheWorthyGrant = "dagger-of-the-worthy/afflict-1"

// Dagger of the Worthy — {2} Artifact — Equipment (Hour of Devastation):
//
//	"Equipped creature gets +2/+0 and has afflict 1. (Whenever it becomes
//	 blocked, defending player loses 1 life.)
//	 Equip {2}"
//
// Afflict is a triggered ability, given to the equipped creature as a bundle
// (Biorganic Carapace's shape); it goes with the Dagger.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "459d7385-11b9-4f81-a6c1-0a77d163a0e5",
		Name:         "Dagger of the Worthy",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       daggerOfTheWorthyGrant,
			Triggered: []game.TriggeredAbility{effects.Afflict(1, "Dagger of the Worthy")},
			Text:      "Afflict 1 (Whenever this creature becomes blocked, defending player loses 1 life.)",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			effects.GrantAbilitiesToAttached(daggerOfTheWorthyGrant),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
