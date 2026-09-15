package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Instigator — Creature — Goblin Rogue {1}{R}, 1/1 (EDHREC
// rank 3247):
//
//	"When this creature enters, create a 1/1 red Goblin creature
//	 token."
//
// Two Goblins for two mana. The entry trigger is b06SelfETB with
// Krenko's Goblin.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8b022754-6d16-470e-b754-4df6e4f4709e",
		Name:         "Goblin Instigator",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Goblin Instigator — create a 1/1 red Goblin", Do(CreateToken{Template: RedGoblinToken(), N: 1})),
		},
	})
}
