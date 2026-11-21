package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Earthbending Student — Creature — Human Warrior Ally {2}{G}, 1/3
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"When this creature enters, earthbend 2. (Target land you control
//	 becomes a 0/0 creature with haste that's still a land. Put two +1/+1
//	 counters on it. When it dies or is exiled, return it to the
//	 battlefield tapped.)
//	 Land creatures you control have vigilance. (Attacking doesn't cause
//	 them to tap.)"
//
// The entry is the earthbend keyword action with its own target clause.
// The grant is Aang, Destined Savior's: a layer-6 keyword grant to every
// land creature you control, read through the layers.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8f8fd2ba-df0e-49bd-a916-cc4f50b9fa24",
		Name:         "Earthbending Student",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AvB14WhenEntersEarthbend("Earthbending Student — earthbend 2", 2),
		},
		Static: []game.StaticAbility{
			effects.KeywordGrant(effects.AvB38LandCreaturesYouControl, "vigilance"),
		},
	})
}
