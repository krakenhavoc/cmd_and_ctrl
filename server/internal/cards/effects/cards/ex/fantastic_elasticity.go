package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fantastic Elasticity — Sorcery {2}{U}:
//
//	"Choose one —
//	 • Return target nonland permanent to its owner's hand.
//	 • Return target instant or sorcery card from your graveyard to
//	   your hand.
//	 Rebound (If you cast this spell from your hand, exile it as it
//	 resolves. At the beginning of your next upkeep, you may cast this
//	 card from exile without paying its mana cost.)"
//
// The rebound cast chooses its mode afresh (CR 700.2a). Rebound is the
// engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "67af9477-82b2-459e-8c09-b9a93a1d1c94",
		Name:            "Fantastic Elasticity",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Return target nonland permanent to its owner's hand.",
				effects.TargetPermanent("target nonland permanent", effects.Nonland()),
				effects.BounceTheModesTarget),
			effects.ModeDoing("Return target instant or sorcery card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
				effects.ReturnTheModesGraveyardTargetToHand),
		),
	})
}
