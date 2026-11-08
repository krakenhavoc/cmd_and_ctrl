package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "dc3073fc-dc23-426e-b867-fee5dd99fe25",
		Name:            "Shipwreck Dowser",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Shipwreck Dowser — return target instant or sorcery card from your graveyard to your hand",
				func(g *game.Game, item *game.StackItem) error {
					return returnLegalGraveyardTargetsToHand(NewContext(g, item))
				}),
				TargetCardInGraveyard("target instant or sorcery card from your graveyard",
					YouOwn(), Or(Instant(), Sorcery())),
			),
		},
	})
}
