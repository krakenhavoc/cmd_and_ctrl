package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spinner of Souls — Creature — Spider Spirit {2}{G}, 4/3 (Foundations,
// #4465):
//
//	"Reach
//	 Whenever another nontoken creature you control dies, you may reveal
//	 cards from the top of your library until you reveal a creature card.
//	 Put that card into your hand and the rest on the bottom of your
//	 library in a random order."
//
// Sifter of Skulls' trigger condition (another creature, not a token,
// that you controlled when it died) with Page, Loose Leaf's
// reveal-until-to-hand as its effect. "You may" is the trigger's CR 603.5
// prompt, asked as it resolves. A library with no creature card reveals
// itself and every card goes to the bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6883e79e-0b4f-4fb8-aee3-a3905246a408",
		Name:            "Spinner of Souls",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Optional(On(game.EventLTB, sifterOfSkullsCondition,
				"Spinner of Souls — reveal until a creature card and put it into your hand",
				socB30RevealUntilToHandRestBottom("Spinner of Souls", game.Card.IsCreature)),
				"Spinner of Souls — reveal cards until you reveal a creature card?"),
				game.Purpose{DeathPayoff: true}),
		},
	})
}
