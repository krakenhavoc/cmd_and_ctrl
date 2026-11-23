package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bile-Vial Boggart — Creature — Goblin Assassin {B}, 1/1 (Lorwyn
// Eclipsed, #4458):
//
//	"When this creature dies, put a -1/-1 counter on up to one target
//	 creature."
//
// A dies trigger (WhenThisDies) with an "up to one" creature clause
// chosen as it goes on the stack, any controller's creature. The body is
// the shared putACounterOnTheTarget: with no target chosen, or a target
// gone by resolution, nothing happens (CR 608.2b). The clause is Harmful,
// as Plague Belcher's is, so a bot that must aim it at its own board picks
// its least valuable creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a651557-ce08-46ba-aecd-0e4fbfdef3c0",
		Name:         "Bile-Vial Boggart",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Bile-Vial Boggart — put a -1/-1 counter on up to one target creature",
					effects.PutACounterOnTheTarget(game.CounterMinusOne)),
				effects.Harmful(effects.UpToOneTargetCreature("up to one target creature")),
			),
		},
	})
}
