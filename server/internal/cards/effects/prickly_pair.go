package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prickly Pair — Creature — Plant Mercenary {2}{R}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature enters, create a 1/1 red Mercenary creature token
//	 with '{T}: Target creature you control gets +1/+0 until end of
//	 turn. Activate only as a sorcery.'"
//
// An enters trigger making the set's Mercenary template (MercenaryToken).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ce8e0d0-6a3e-4c47-a7b7-c1b1ea1b9426",
		Name:         "Prickly Pair",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Prickly Pair — create a 1/1 red Mercenary", createAMercenary()),
		},
	})
}
