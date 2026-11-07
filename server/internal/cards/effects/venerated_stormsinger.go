package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venerated Stormsinger — Creature — Orc Cleric {3}{B}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)
//	 Whenever this creature or another creature you control dies, each
//	 opponent loses 1 life and you gain 1 life."
//
// Mobilize is mobilize.go (CR 702.181). The dies trigger is Zulaport
// Cutthroat's: any creature you control dying, the Stormsinger itself and
// tokens included, drains each opponent for 1, and you gain exactly 1
// whatever the table size. The mobilized Warrior sacrificed at the end
// step is one of those creatures.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4aff2f31-619f-406e-b2c9-a2c733c04bc4",
		Name:         "Venerated Stormsinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Mobilize("Venerated Stormsinger", 1),
			TriggerWithPurpose(WheneverACreatureYouControlDies("Venerated Stormsinger — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})), game.Purpose{DeathPayoff: true}),
		},
	})
}
