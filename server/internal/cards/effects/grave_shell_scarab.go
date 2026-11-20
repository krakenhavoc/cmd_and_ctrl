package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Grave-Shell Scarab — Creature — Insect {2}{B}{G}{G}, 4/4 (Ravnica: City of
// Guilds, #5600):
//
//	"{1}, Sacrifice this creature: Draw a card.
//	 Dredge 1 (If you would draw a card, you may mill a card instead. If
//	 you do, return this card from your graveyard to your hand.)"
//
// Heart Warden's sacrifice-to-draw row, and dredge as the engine's draw
// replacement while the card is in its owner's graveyard (CR 702.52).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d954def-0045-4860-b154-cb64b6835fe0",
		Name:         "Grave-Shell Scarab",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Effect:  ebDrawACard,
		}},
		Replacements: []game.ReplacementEffect{Dredge(1)},
	})
}
