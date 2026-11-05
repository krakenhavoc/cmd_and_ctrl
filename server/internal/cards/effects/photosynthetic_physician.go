package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Photosynthetic Physician — {4}{G} Creature — Plant Doctor, 5/5 (Star
// Trek, #4973):
//
//	"Reach
//	 {G}, Exile this card from your graveyard: Create a 1/1 red Officer
//	 creature token. Activate only as a sorcery."
//
// Eternal Student's graveyard activation: exiling the card is part of the
// cost (CR 602.2b), so it can be activated once, and only at sorcery
// timing (CR 602.5d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "08787942-456f-4c71-a821-1c90b5639ca1",
		Name:            "Photosynthetic Physician",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []ActivatedAbility{{
			Label:        "{G}, Exile this card from your graveyard: Create a 1/1 red Officer creature token. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{G}"), ExileThis()),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Purpose:      game.Purpose{Answers: game.AnswerValue},
			Effect:       Do(CreateToken{Template: TokenCard("1/1 red Officer"), N: 1}),
		}},
	})
}
