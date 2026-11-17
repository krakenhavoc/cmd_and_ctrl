package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Benalish Sleeper — Creature — Phyrexian Human Soldier {1}{W}, 3/1
// (Dominaria United, #5487):
//
//	"Kicker {B} (You may pay an additional {B} as you cast this spell.)
//	 When this creature enters, if it was kicked, each player sacrifices a
//	 creature of their choice."
//
// Abyssal Gatekeeper's each-player edict behind the kicked intervening
// "if" (CR 603.4). The Sleeper is on the battlefield as the trigger
// resolves, so its controller may sacrifice it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "14c3692d-16a9-4ef0-98cd-60d15e4aa31f",
		Name:          "Benalish Sleeper",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Kicker("{B}")},
		Triggered: []game.TriggeredAbility{
			TriggerEdict(On(game.EventETB, AllOf(Self, ThisKickedAtLeast(1)),
				"Benalish Sleeper — kicked: each player sacrifices a creature",
				Do(EachPlayerSacrifices{
					Match: Creature(),
					Label: "a creature",
				}))),
		},
	})
}
