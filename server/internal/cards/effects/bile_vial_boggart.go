package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "5a651557-ce08-46ba-aecd-0e4fbfdef3c0",
		Name:         "Bile-Vial Boggart",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisDies("Bile-Vial Boggart — put a -1/-1 counter on up to one target creature",
					putACounterOnTheTarget(game.CounterMinusOne)),
				Harmful(UpToOneTargetCreature("up to one target creature")),
			),
		},
	})
}
