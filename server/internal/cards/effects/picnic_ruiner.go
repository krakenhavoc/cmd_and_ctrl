package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     picnicRuinerOracleID,
		Name:         "Picnic Ruiner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, hobbitFerociousAttack, "Picnic Ruiner — gains double strike until end of turn",
				thisGainsKeywordUntilEndOfTurn("double strike", "Picnic Ruiner — double strike until end of turn")),
		},
	})
	Register(Spec{
		OracleID:     picnicRuinerOracleID + "#1",
		Name:         "Stolen Goodies",
		Completeness: CompletenessFull,
		Targets: TargetCreature("any number of target creatures you control", YouControl()).
			WithCount(1, 3).Dividing(Divide(3)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error { return PutDividedCounters(ctx, game.CounterPlusOne) },
	})
}
