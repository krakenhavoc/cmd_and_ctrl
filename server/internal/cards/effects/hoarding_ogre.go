package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hoarding Ogre — "Whenever this creature attacks, roll a d20. 1–9: Create
// a Treasure token. 10–19: Create two Treasure tokens. 20: Create three
// Treasure tokens."
func init() {
	Register(Spec{
		OracleID:     "faeee3a7-c66c-48d8-a994-12377d3bb729",
		Name:         "Hoarding Ogre",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hoarding Ogre — roll a d20 and create Treasures", func(g *game.Game, item *game.StackItem) error {
				results, err := rollDice(NewContext(g, item), 20, 1)
				if err != nil || len(results) == 0 {
					return err
				}
				n := 1
				if results[0] >= 10 {
					n = 2
				}
				if results[0] == 20 {
					n = 3
				}
				return treasureTokens(g, item.Controller, n)
			}),
		},
	})
}
