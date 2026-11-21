package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Strike Force — Creature — Dwarf Berserker {4}{R}, 4/3:
//
//	"Discard a card at random: This creature gains first strike and haste
//	 until end of turn."
//
// ADR 0109 §7, owner decision 3: "Discard a card at random" is a cost
// the activator chooses nothing for (CR 701.9b). The engine draws the
// card from the hand the rest of the cost leaves, paid after every other
// cost (CR 601.2h), and an empty hand can't pay it (CR 118.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "66dd0e68-f5ec-45e5-991f-d588aa726387",
		Name:         "Dwarven Strike Force",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card at random: This creature gains first strike and haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.DiscardAtRandom(1, "a card at random"),
			Effect:  effects.ThisCreatureUntilEOT("Dwarven Strike Force — first strike and haste", 0, 0, "first strike", "haste"),
		}},
	})
}
