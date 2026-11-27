package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Favor of the Mighty — Kindred Enchantment — Giant {1}{W} (Lorwyn, S82 OLD b557):
//
//	"Each creature with the greatest mana value has protection from each
//	 color."
//
// A layer 6 grant of the five protections to every creature whose mana value
// ties the greatest among creatures on the battlefield (any controller), read
// through the layered view. Tokens and face-down creatures have mana value 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2ad82e6-4669-4d87-bea3-f2ac6ebe5ba9",
		Name:         "Favor of the Mighty",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "Each creature with the greatest mana value has protection from each color.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, _ *game.Card) bool {
				if !target.IsCreature() {
					return false
				}
				mv, ok := g.ManaValueForEffect(*target)
				return ok && mv == favorOfTheMightyGreatest(g)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, col := range game.AllColors {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, game.ProtectionFromColor(col))
				}
			},
		}},
	})
}

// favorOfTheMightyGreatest is the greatest mana value among creatures on the
// battlefield, 0 with none.
func favorOfTheMightyGreatest(g *game.Game) int {
	greatest := 0
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if !c.IsCreature() {
			continue
		}
		if mv, ok := g.ManaValueForEffect(*c); ok && mv > greatest {
			greatest = mv
		}
	}
	return greatest
}
