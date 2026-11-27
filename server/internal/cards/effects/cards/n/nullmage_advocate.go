package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nullmage Advocate — Creature — Insect Druid {2}{G}, 2/3 (Planechase 2012, S82
// OLD, #5569):
//
//	"{T}: Return two target cards from an opponent's graveyard to their hand.
//	 Destroy target artifact or enchantment."
//
// Shieldmage Advocate's two target clauses: both cards come from one opponent's
// graveyard (Decompose's single-graveyard rule), exactly two, and the artifact or
// enchantment is the second clause. Each target still legal as it resolves is
// dealt with and one that has gone is skipped (CR 608.2b); a card returns to its
// owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "558568a6-5e22-482f-9096-e76682965c21",
		Name:         "Nullmage Advocate",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: Return two target cards from an opponent's graveyard to their hand. Destroy target artifact or enchantment.",
			Cost:  effects.TapCost(),
			Targets: effects.Clauses(
				effects.TargetCardInGraveyard("two target cards from an opponent's graveyard", effects.NotOwnedByCaster()).
					WithCount(2, 2).AllShare(effects.FromASingleGraveyard()),
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			),
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Effect:  effects.S82OldB432NullmageReturnAndDestroy,
		}},
	})
}
