package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Firewing Phoenix — Creature — Phoenix {3}{R}, 4/2 (Magic 2013, #5568):
//
//	"Flying
//	 {1}{R}{R}{R}: Return this card from your graveyard to your hand."
//
// Blessed Ghoul's graveyard row: activated only while the card is in its
// owner's graveyard, and nothing if it has left by resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "326d9b5c-ac04-45b5-9a8c-a9be7d8cca1b",
		Name:            "Firewing Phoenix",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}{R}{R}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{1}{R}{R}{R}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
