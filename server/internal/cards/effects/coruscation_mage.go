package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coruscation Mage — Creature — Otter Wizard {1}{R}, 2/2:
//
//	"Offspring {2} (You may pay an additional {2} as you cast this
//	 spell. If you do, when this creature enters, create a 1/1 token
//	 copy of it.)
//	 Whenever you cast a noncreature spell, this creature deals 1
//	 damage to each opponent."
//
// Offspring is the optional additional cost plus its gated entry
// trigger (offspring.go); the token copy carries the same cast
// trigger, as the 1/1 copy of a Mage should. The ping is Black Waltz
// No. 3's, for one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "88bb91b5-2ccd-4ce9-8cd4-e54d63c12abf",
		Name:          "Coruscation Mage",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered: []game.TriggeredAbility{
			OffspringToken("Coruscation Mage"),
			WheneverYouCast(Noncreature(), "Coruscation Mage — 1 damage to each opponent", coruscationMagePing),
		},
	})
}

func coruscationMagePing(g *game.Game, item *game.StackItem) error {
	return damageToEachOpponent(g, item, 1)
}
