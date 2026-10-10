package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magus of the Will — Creature — Human Wizard {2}{B}, 2/2:
//
//	"{2}{B}, {T}, Exile this creature: Until end of turn, you may play
//	 lands and cast spells from your graveyard. If a card would be put
//	 into your graveyard from anywhere this turn, exile that card
//	 instead."
//
// Yawgmoth's Will's clause pair (GraveyardPlayThisTurn) on an
// activated ability. Exiling the Magus is a cost (CR 118.3), so it
// never reaches the graveyard and is not itself exiled by the effect.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f32c5530-6692-4d47-8789-c73da23fd5b7",
		Name:         "Magus of the Will",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, {T}, Exile this creature: Until end of turn, you may play lands and cast spells from your graveyard. If a card would be put into your graveyard from anywhere this turn, exile that card instead.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{B}"), TapCost(), ExileThis()),
			Effect:  Do(YawgmothsWillThisTurn()),
		}},
	})
}
