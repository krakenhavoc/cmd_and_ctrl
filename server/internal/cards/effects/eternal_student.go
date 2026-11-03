package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Student — Creature — Zombie Warlock {3}{B}, 4/2 (Secrets of
// Strixhaven, #4456):
//
//	"{1}{B}, Exile this card from your graveyard: Create two 1/1 white and
//	 black Inkling creature tokens with flying."
//
// An ability that functions from the graveyard (its cost moves the card out of the graveyard, CR 113.6m), built the way
// embalm is: the exile is part of the cost, paid at the announce (CR
// 602.2b), so the card is gone before the ability resolves and a response
// cannot get it back. Unlike embalm it has no timing restriction, so it can
// be activated at instant speed whenever its owner has priority.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "140cb5cf-8aa2-457c-9d65-5acefea00195",
		Name:         "Eternal Student",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{1}{B}, Exile this card from your graveyard: Create two 1/1 white and black Inkling creature tokens with flying.",
			Cost:  Plus(ManaCost("{1}{B}"), ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			// Two tokens: value, not an answer to anything on the stack.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: Do(CreateToken{
				Template: TokenCard("1/1 white and black Inkling with flying"), N: 2,
			}),
		}},
	})
}
