package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hissing Iguanar — Creature — Lizard {2}{R}, 3/1 (Shards of Alara,
// #5588):
//
//	"Whenever another creature dies, you may have this creature deal 1
//	 damage to target player or planeswalker."
//
// Deathgreeter's trigger, any player's creature but this one, with a
// target chosen as it goes on the stack (CR 603.3d) and a "you may"
// asked as it resolves (CR 603.5). A target gone in response takes
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7c4fc524-4734-4fb8-a5ca-ac2617242c3f",
		Name:         "Hissing Iguanar",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					On(game.EventLTB, AnotherCreatureDied,
						"Hissing Iguanar — 1 damage to target player or planeswalker", DamageToFirstTarget(1)),
					"Hissing Iguanar — deal 1 damage to that player or planeswalker?"),
				targetPlayerOrPlaneswalker()),
		},
	})
}
