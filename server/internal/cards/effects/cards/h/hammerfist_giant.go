package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "ce916c92-350c-4546-882b-555dc450413d",
		Name:         "Hammerfist Giant",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}: This creature deals 4 damage to each creature without flying and each player.",
			// ADR 0126 §6: only creatures without flying.
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, Partial: true}},
			Cost:    effects.TapCost(),
			Effect:  effects.ThisDealsDamageToEachCreatureMatchingAndEachPlayer(effects.WithoutKeyword("flying"), 4),
		}},
	})
}
