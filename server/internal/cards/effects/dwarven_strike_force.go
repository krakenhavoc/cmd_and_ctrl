package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "66dd0e68-f5ec-45e5-991f-d588aa726387",
		Name:         "Dwarven Strike Force",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Discard a card at random: This creature gains first strike and haste until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    DiscardAtRandom(1, "a card at random"),
			Effect:  thisCreatureUntilEOT("Dwarven Strike Force — first strike and haste", 0, 0, "first strike", "haste"),
		}},
	})
}
