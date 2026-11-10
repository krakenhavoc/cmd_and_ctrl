package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Greed's Gambit — Enchantment {3}{B} (The Big Score, #4468):
//
//	"When this enchantment enters, you draw three cards, gain 6 life, and
//	 create three 2/1 black Bat creature tokens with flying.
//	 At the beginning of your end step, you discard a card, lose 2 life,
//	 and sacrifice a creature.
//	 When this enchantment leaves the battlefield, you discard three
//	 cards, lose 6 life, and sacrifice three creatures."
//
// Three triggers, each in printed order. The discards and sacrifices are
// the controller's own choices (CR 701.9a, 701.21a); a hand or board with
// fewer cards than owed gives what it has.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0868ebbb-2fe9-4f53-9697-acad8cda5ec1",
		Name:         "Greed's Gambit",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 3, Tokens: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Greed's Gambit — draw three, gain 6, create three 2/1 Bats", Do(
				DrawCards{N: 3}, GainLife{Amount: 6}, CreateToken{Template: b35BatToken(), N: 3})),
			AtYourEndStep("Greed's Gambit — discard a card, lose 2 life, sacrifice a creature",
				b35YouDiscardLoseSacrifice(1, 2, 1, "Greed's Gambit")),
			WhenThisLeaves("Greed's Gambit — discard three cards, lose 6 life, sacrifice three creatures",
				b35YouDiscardLoseSacrifice(3, 6, 3, "Greed's Gambit")),
		},
	})
}
