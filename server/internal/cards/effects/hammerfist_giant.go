package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hammerfist Giant — Creature — Giant Warrior {4}{R}{R}, 5/4 (Ravnica: City
// of Guilds, #5600):
//
//	"{T}: This creature deals 4 damage to each creature without flying and
//	 each player."
//
// Warmonger's sweep at 4 on a tap row: 4 damage from the Giant to each
// creature without flying (itself included) and to each player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce916c92-350c-4546-882b-555dc450413d",
		Name:         "Hammerfist Giant",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{T}: This creature deals 4 damage to each creature without flying and each player.",
			// ADR 0126 §6: only creatures without flying.
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, Partial: true}},
			Cost:    TapCost(),
			Effect:  thisDealsDamageToEachCreatureMatchingAndEachPlayer(WithoutKeyword("flying"), 4),
		}},
	})
}
