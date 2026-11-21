package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Return from Extinction — Sorcery {1}{B} (Modern Horizons, #5519):
//
//	"Choose one —
//	 • Return target creature card from your graveyard to your hand.
//	 • Return two target creature cards that share a creature type from your
//	 graveyard to your hand."
//
// Unbury's text at sorcery speed: the second bullet's second clause is
// SharesACreatureTypeWithTargetCard over the first, and a changeling card
// shares a type with any creature card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3937a96c-3be1-465d-98b6-1db67d2215e6",
		Name:         "Return from Extinction",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
				effects.EoeABCharmReturn),
			effects.ModeDoing("Return two target creature cards that share a creature type from your graveyard to your hand.",
				effects.Clauses(
					effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
					effects.Distinct(effects.SharesACreatureTypeWithTargetCard(
						effects.TargetCardInGraveyard("another target creature card from your graveyard that shares a creature type with it", effects.Creature(), effects.YouOwn()), 0)),
				),
				effects.EoeABCharmReturn),
		),
	})
}
