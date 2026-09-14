package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beetleback Chief — Creature — Goblin Warrior {2}{R}{R}, 2/2 (EDHREC
// rank 3580):
//
//	"When this creature enters, create two 1/1 red Goblin creature
//	 tokens."
//
// Four power across three bodies. The tokens are Krenko's Goblins.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5c5f64c-6911-430c-a825-b32b96d39c7d",
		Name:         "Beetleback Chief",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Beetleback Chief — create two 1/1 red Goblins", b34CreateTokens(RedGoblinToken, 2))
			},
		}},
	})
}
