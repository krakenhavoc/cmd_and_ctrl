package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Picnic Ruiner // Stolen Goodies — Creature — Goblin Rogue {1}{R}, 2/2 //
// Sorcery — Adventure {3}{G} (Wilds of Eldraine, #4471):
//
//	Picnic Ruiner — "Whenever this creature attacks while you control a
//	 creature with power 4 or greater, this creature gains double strike
//	 until end of turn."
//	Stolen Goodies — "Distribute three +1/+1 counters among any number of
//	 target creatures you control."
//
// The attack trigger is Ruby's (hobbitFerociousAttack): "while" is read
// as the attack is declared. The Adventure is Biogenic Upgrade's divided
// counters over one to three targets (each gets at least one, CR 601.2d).
//
// No simplification.
const picnicRuinerOracleID = "334fba11-e500-40fd-a142-41ff553642b5"

func init() {
	effects.Register(effects.Spec{
		OracleID:     picnicRuinerOracleID,
		Name:         "Picnic Ruiner",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.HobbitFerociousAttack, "Picnic Ruiner — gains double strike until end of turn",
				effects.ThisGainsKeywordUntilEndOfTurn("double strike", "Picnic Ruiner — double strike until end of turn")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     picnicRuinerOracleID + "#1",
		Name:         "Stolen Goodies",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCreature("any number of target creatures you control", effects.YouControl()).
			WithCount(1, 3).Dividing(effects.Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
