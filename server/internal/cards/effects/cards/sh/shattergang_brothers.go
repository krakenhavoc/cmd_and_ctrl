package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shattergang Brothers — Legendary Creature — Goblin Artificer {1}{B}{R}{G},
// 3/3 (Double Masters 2022):
//
//	"{2}{B}, Sacrifice a creature: Each other player sacrifices a creature.
//	 {2}{R}, Sacrifice an artifact: Each other player sacrifices an artifact.
//	 {2}{G}, Sacrifice an enchantment: Each other player sacrifices an
//	 enchantment."
//
// Three edicts, each paid for with a sacrifice of its own kind. The cost is
// paid first (CR 602.2b), so the Brothers can sacrifice themselves to the
// first one, and each other player then chooses their own permanent of that
// kind as the ability resolves (CR 701.21).
//
// No simplification.
func init() {
	edict := func(mana, cost string, answers game.Answers, match effects.CardPredicate, sac game.AbilityCost) effects.ActivatedAbility {
		return effects.ActivatedAbility{
			Label:   mana + ", Sacrifice " + cost + ": Each other player sacrifices " + cost + ".",
			Purpose: game.Purpose{Answers: answers},
			Cost:    effects.Plus(effects.ManaCost(mana), sac),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.EachPlayerSacrifices{ExceptController: true, Match: match, Label: cost}.Apply(effects.NewContext(g, item))
			},
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "7fb63d9a-8d90-4b43-8390-924de2d7e32c",
		Name:         "Shattergang Brothers",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			edict("{2}{B}", "a creature", game.AnswerRemove|game.AnswerSacOutlet|game.AnswerProtect|game.AnswerSurvive, effects.Creature(), effects.SacrificeACreature()),
			edict("{2}{R}", "an artifact", game.AnswerRemove|game.AnswerSacOutlet|game.AnswerProtect|game.AnswerSurvive, effects.Artifact(), effects.SacrificeN(1, "an artifact", effects.Artifact())),
			edict("{2}{G}", "an enchantment", game.AnswerRemove|game.AnswerSacOutlet|game.AnswerProtect|game.AnswerSurvive, effects.Enchantment(), effects.SacrificeN(1, "an enchantment", effects.Enchantment())),
		},
	})
}
