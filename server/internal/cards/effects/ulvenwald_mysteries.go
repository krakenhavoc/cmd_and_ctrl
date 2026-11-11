package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ulvenwald Mysteries — Enchantment {2}{G} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Whenever a nontoken creature you control dies, investigate.
//	 Whenever you sacrifice a Clue, create a 1/1 white Human Soldier
//	 creature token."
//
// The death trigger is Sifter of Skulls's condition with the "another"
// dropped: the Mysteries is not a creature, so every nontoken creature
// of yours counts, read from its last-known information (CR 603.10a).
// The Clue sacrifice is Tireless Tracker's (lonisYouSacrificedAClue),
// so a Clue the first ability just made pays out when it is cracked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6588ea9a-658f-40d9-a64b-3a418ab81183",
		Name:         "Ulvenwald Mysteries",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, sifterOfSkullsCondition,
				"Ulvenwald Mysteries — investigate", Do(Investigate{N: 1})),
			On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Ulvenwald Mysteries — create a 1/1 white Human Soldier creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
