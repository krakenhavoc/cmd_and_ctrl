package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Falthis, Shadowcat Familiar — Legendary Creature — Nightmare Cat {2}{B},
// 2/2 (Commander Legends, #5507):
//
//	"Commanders you control have menace and deathtouch.
//	 Partner (You can have two commanders if both have partner.)"
//
// The black member of the familiar cycle, Anara's and Guardian
// Augmenter's grant with two keywords: one layer-6 static over
// Card.IsCommander under the source's controller, so Falthis grants them
// to itself when it is one of your commanders, a commander you have stolen
// is covered, and your own commander under another player is not. Partner
// is the deck rule internal/deck reads off the oracle text (ADR 0144).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6a0b1b7e-1a93-4c00-be94-8fa56af57468",
		Name:         "Falthis, Shadowcat Familiar",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Commanders you control have menace and deathtouch.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.IsCommander && target.Controller == source.Controller
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "menace")
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "deathtouch")
			},
		}},
	})
}
