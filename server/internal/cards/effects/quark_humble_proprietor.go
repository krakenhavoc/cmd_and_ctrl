package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Quark, Humble Proprietor — Legendary Creature — Ferengi Rogue {2}{R}
// (Star Trek, #4973), 1/2:
//
//	"You can't sacrifice Treasure tokens.
//	 At the beginning of each upkeep, create a Treasure token."
//
// A player can sacrifice only permanents they control (CR 701.21a), so
// "you can't sacrifice Treasure tokens" is exactly "Treasure tokens you
// control can't be sacrificed": Quark's own layer-6 static sets the
// #3275 flag on them, which bars a sacrifice cost (the Treasure's own
// mana ability included, CR 118.3), an edict's choice and the move
// itself (CR 101.2). A Treasure that is not a token, and an opponent's
// Treasures, are untouched. The token comes every upkeep, each
// player's, for Quark's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61703e94-cb57-44d8-b37b-7b14b3cfd86d",
		Name:         "Quark, Humble Proprietor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "You can't sacrifice Treasure tokens.",
			Layer:     game.Layer6Ability,
			AppliesTo: trekB17TreasureTokensYouControl,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.CantBeSacrificed = true
			},
		}},
		Triggered: []game.TriggeredAbility{
			AtEachUpkeep("Quark, Humble Proprietor — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
