package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Squandered Resources — Enchantment {B}{G} (Visions, S82 OLD, #5631):
//
//	"Sacrifice a land: Add one mana of any type the sacrificed land could
//	 produce."
//
// A mana ability (CR 605.1a) whose cost sacrifices a land. What the land could
// produce is read as it stood on the battlefield (CR 106.7, CR 608.2h), before
// the sacrifice moves it, and recorded on the payment (PaidCost.SacrificedMana):
// a Forest under Blanket of Night offers {G} or {B}. A land that could produce
// nothing adds nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "620d0572-6d22-43e2-83a9-000e605c8a2b",
		Name:         "Squandered Resources",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:            effects.ManaAbilityCost{SacrificeOther: effects.SacrificeSpec("a land", effects.Land())},
			ProducedForPaid: effects.S82OldB788AnyTypeTheSacrificedLandCouldProduce,
			Label:           "Sacrifice a land: Add one mana of any type the sacrificed land could produce",
			Answers:         game.AnswerSacOutlet,
		}},
	})
}
