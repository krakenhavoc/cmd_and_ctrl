package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Leering Onlooker — Creature — Vampire {1}{B}, 1/3 (Murders at Karlov
// Manor, #4469):
//
//	"Flying
//	 {2}{B}{B}, Exile this card from your graveyard: Create two tapped 1/1
//	 black Bat creature tokens with flying."
//
// The graveyard activation is Spider-Slayer, Hatred Honed's: exiling the
// card is part of the cost (CR 602.2b), so it can be activated once, at
// instant speed. The Bats enter tapped (CreateTokenAdvanced,
// EntersTapped), so they block nothing the turn they arrive, and the row
// declares that it answers nothing (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "06a37f16-c929-4b89-bf18-1b3f8f472eef",
		Name:            "Leering Onlooker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}{B}, Exile this card from your graveyard: Create two tapped 1/1 black Bat creature tokens with flying.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 2},
			Cost:    Plus(ManaCost("{2}{B}{B}"), ExileThis()),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect: Do(CreateTokenAdvanced{
				Spec: Token(TokenCard("1/1 black Bat with flying")).EntersTapped(),
				N:    2,
			}),
		}},
	})
}
