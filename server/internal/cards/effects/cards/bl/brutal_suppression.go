package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brutal Suppression — Enchantment {R} (Prophecy, S82 OLD batch 674, #5617):
//
//	"Activated abilities of nontoken Rebels cost an additional "Sacrifice a
//	 land" to activate."
//
// The imposed-sacrifice static (#6205, CR 602.2b): the engine folds one land
// sacrifice into the cost of every activated ability, mana abilities
// included, whose source is a nontoken Rebel, whoever controls it. A token
// Rebel's and a non-Rebel's abilities are untouched. The creature type is
// read through the source's text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b4971aa3-32d1-4bc7-8133-ed67763d2dc6",
		Name:         "Brutal Suppression",
		Completeness: effects.CompletenessFull,
		ImposedSacrifices: []game.ImposedSacrifice{{
			Label:       "Activated abilities of nontoken Rebels cost an additional \"Sacrifice a land\" to activate.",
			Kind:        game.SacrificeKind{Label: "a land", CardTypes: []string{"land"}},
			Activations: true,
			AppliesTo: func(q game.ImposedSacrificeQuery) bool {
				return !q.Object.IsToken() && q.Object.HasSubtype(q.Source.TextSubtype("Rebel"))
			},
		}},
	})
}
