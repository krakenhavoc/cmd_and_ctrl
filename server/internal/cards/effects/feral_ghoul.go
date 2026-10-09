package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feral Ghoul — Creature — Zombie Mutant {2}{B}, 2/2:
//
//	"Menace
//	 Whenever another creature you control dies, put a +1/+1 counter on
//	 this creature.
//	 When this creature dies, each opponent gets a number of rad counters
//	 equal to its power."
//
// #2042. "Its power" is the Ghoul's power as it last existed on the
// battlefield (CR 603.10a), counters and pumps included, carried on the
// trigger (eerieDiesWithPower). A Ghoul that died with 0 or less power
// gives nothing.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "284db94d-e583-4355-89fd-6c0906f9d17e",
		Name:            "Feral Ghoul",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AllOf(AnotherCreatureDied, ACreatureYouControlDied),
				"Feral Ghoul — put a +1/+1 counter on it", putACounterOnThis(game.CounterPlusOne)),
			eerieDiesWithPower("Feral Ghoul — each opponent gets rad counters equal to its power", nil,
				func(g *game.Game, item *game.StackItem) error {
					return eachOpponentGetsRadCounters(g, item.Controller, item.Params.Amount)
				}),
		},
	})
}
