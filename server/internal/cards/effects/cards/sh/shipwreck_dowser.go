package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shipwreck Dowser — Creature — Merfolk Wizard {3}{U}{U}, 3/3
// (Foundations, #4465):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 When this creature enters, return target instant or sorcery card
//	 from your graveyard to your hand."
//
// Pinnacle Monk's text on a bigger body: prowess is the printed keyword,
// and the enters trigger is Zealous Lorecaster's targeted return, which
// returns nothing if the card has left the graveyard by resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dc3073fc-dc23-426e-b867-fee5dd99fe25",
		Name:            "Shipwreck Dowser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Shipwreck Dowser — return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnLegalGraveyardTargetsToHand(effects.NewContext(g, item))
				}),
				effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
			),
		},
	})
}
