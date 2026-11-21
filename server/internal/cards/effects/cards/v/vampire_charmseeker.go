package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vampire Charmseeker — Creature — Vampire Wizard {6}{U}{B}:
//
//	"Flying
//	 Assist (Another player can pay up to {6} of this spell's cost.)
//	 When this creature enters, return target instant, sorcery, or
//	 creature card from a graveyard to its owner's hand."
//
// The pick is any graveyard, so the spec reads every graveyard zone
// (TargetCardInGraveyard) with the three types as one Or. The card goes
// to its owner's hand through returnFirstLegalGraveyardTarget, which
// leaves the Controller at its zero value ("its owner").
func init() {
	effects.Register(effects.Spec{
		OracleID:        "993bc179-4e24-4161-866f-5e146cbdd2f5",
		Name:            "Vampire Charmseeker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordAssist, "flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Vampire Charmseeker — return target instant, sorcery, or creature card from a graveyard to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						return effects.ReturnFirstLegalGraveyardTarget(g, item, game.ZoneHand)
					}),
				effects.TargetCardInGraveyard("target instant, sorcery, or creature card in a graveyard",
					effects.Or(effects.Instant(), effects.Sorcery(), effects.Creature()))),
		},
	})
}
