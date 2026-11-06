package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merchant of Many Hats — Creature — Human Peasant Ally {1}{B}, 2/2
// (Avatar: The Last Airbender, #4459):
//
//	"{2}{B}: Return this card from your graveyard to your hand."
//
// Blessed Ghoul's graveyard ability with its own cost: it works only
// while the card is in your graveyard (Zones), and a card that left the
// graveyard before the ability resolves stays where it is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23820b45-2f3a-4b6e-99ea-d84caccb7e37",
		Name:         "Merchant of Many Hats",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
