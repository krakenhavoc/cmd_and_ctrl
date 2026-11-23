package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "6588ea9a-658f-40d9-a64b-3a418ab81183",
		Name:         "Ulvenwald Mysteries",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.SifterOfSkullsCondition,
				"Ulvenwald Mysteries — investigate", effects.Do(effects.Investigate{N: 1})),
			effects.On(game.EventSacrifice, effects.LonisYouSacrificedAClue,
				"Ulvenwald Mysteries — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
	})
}
