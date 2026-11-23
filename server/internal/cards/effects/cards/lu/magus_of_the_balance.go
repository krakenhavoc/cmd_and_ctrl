package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magus of the Balance — Creature — Human Wizard {1}{W}, 2/2 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"{4}{W}, {T}, Sacrifice this creature: Each player chooses a number of
//	 lands they control equal to the number of lands controlled by the player
//	 who controls the fewest, then sacrifices the rest. Players discard cards
//	 and sacrifice creatures the same way."
//
// Balance, as an ability: three steps in printed order (lands, cards in
// hand, creatures), each reading every player's count as it begins. Each
// player over the fewest sacrifices or discards the difference, of their own
// choice (choosing what to keep is choosing what to lose), the players
// answering in APNAP order (CR 101.4). The Magus is sacrificed as a cost,
// before the counts are read, so it is not among the creatures.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ed450f48-e69d-48d7-82cc-79cf89b2a1fb",
		Name:         "Magus of the Balance",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{4}{W}, {T}, Sacrifice this creature: Each player chooses a number of lands they control equal to the number of lands controlled by the player who controls the fewest, then sacrifices the rest. Players discard cards and sacrifice creatures the same way.",
			Cost:  game.AbilityCost{Tap: true, Mana: "{4}{W}", SacrificeSelf: true},
			Purpose: game.Purpose{
				Answers: game.AnswerRemove,
				Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepSacrifice, Partial: true},
			},
			Effect: effects.Do(effects.MagusOfTheBalanceSteps()...),
		}},
	})
}
