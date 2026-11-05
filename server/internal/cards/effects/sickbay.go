package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sickbay — Artifact {5} (Star Trek, #4973):
//
//	"Whenever a creature you control enters, you gain 1 life.
//	 At the beginning of combat on your turn, put X +1/+1 counters on
//	 target creature you control, where X is one plus the amount of life
//	 you gained this turn."
//
// Two triggers. The first is Ajani's Welcome's (CreatureEnteredUnderYourControl,
// a footprinted GainLife). The second is Additive Evolution's
// beginning-of-combat trigger on a creature you control; X is read as the
// trigger resolves (CR 608.2h), from the turn's tally of the life its
// controller gained (b15LifeGainedThisTurn), so life gained in response
// counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d4163d9-ca02-4f25-9f53-01f6d7cc4b70",
		Name:         "Sickbay",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, CreatureEnteredUnderYourControl, "Sickbay — a creature you control entered: you gain 1 life", Do(GainLife{Amount: 1})),
			Targeting(On(game.EventStepBegan, StepBegan(game.StepBeginCombat, true),
				"Sickbay — put X +1/+1 counters on target creature you control, where X is one plus the life you gained this turn",
				sickbayCounters),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

// sickbayCounters is the beginning-of-combat trigger's resolution.
func sickbayCounters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	target, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return AddCounter{Target: target.ID, Kind: game.CounterPlusOne, N: trekB27OnePlusLifeGainedThisTurn(g, item.Controller)}.Apply(ctx)
}
