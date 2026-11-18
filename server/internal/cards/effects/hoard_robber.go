package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoard Robber — Creature — Tiefling Rogue {1}{B}, 1/3 (Adventures in the
// Forgotten Realms):
//
//	"Whenever this creature deals combat damage to a player, create a
//	 Treasure token."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a04f70c-288d-4010-a6d5-24e9778b9b45",
		Name:         "Hoard Robber",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Hoard Robber — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
