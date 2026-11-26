package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chancellor of the Forge — Creature — Phyrexian Giant {4}{R}{R}{R}, 5/5:
//
//	"You may reveal this card from your opening hand. If you do, at the
//	 beginning of the first upkeep, create a 1/1 red Phyrexian Goblin
//	 creature token with haste.
//	 When this creature enters, create X 1/1 red Phyrexian Goblin creature
//	 tokens with haste, where X is the number of creatures you control."
//
// The reveal is the Chancellor seam (RevealFromOpeningHand, CR 103.6). The
// count is taken as the entry trigger resolves, the Chancellor itself included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0aedf049-db6b-420e-80e7-591ed38d4fba",
		Name:         "Chancellor of the Forge",
		Completeness: effects.CompletenessFull,
		OpeningHand: effects.RevealFromOpeningHand(game.RevealAtTheFirstUpkeep,
			"Chancellor of the Forge — create a 1/1 red Phyrexian Goblin creature token with haste",
			chancellorOfTheForgeFirstUpkeep),
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Chancellor of the Forge — create X 1/1 red Phyrexian Goblin creature tokens with haste, where X is the number of creatures you control",
				effects.ChancellorOfTheForgeETB),
		},
	})
}

var chancellorOfTheForgeFirstUpkeep = game.SimpleDelayedBody("chancellor-of-the-forge/first-upkeep-goblin", func(g *game.Game, item *game.StackItem) error {
	return effects.CreateToken{Template: effects.TokenCard("1/1 red Phyrexian Goblin with haste"), N: 1}.Apply(effects.NewContext(g, item))
})
