package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Swordsworn Cavalier — Creature — Human Knight {1}{W}, 3/1 (March of the
// Machine):
//
//	"This creature has first strike as long as another Knight entered the
//	 battlefield under your control this turn."
func init() {
	Register(Spec{
		OracleID:     "18aa69d5-9611-49fe-8aa4-7b91d5980dbc",
		Name:         "Swordsworn Cavalier",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Has first strike as long as another Knight entered the battlefield under your control this turn.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfCreature(target, g, source) && s82Mom09AnotherKnightEnteredThisTurn(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, "first strike")
			},
		}},
	})
}
