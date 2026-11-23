package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Additive Evolution — Enchantment {3}{G}{G} (Secrets of Strixhaven,
// #4456):
//
//	"When this enchantment enters, create a 0/0 green and blue Fractal
//	 creature token. Put three +1/+1 counters on it.
//	 At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature you control. It gains vigilance until end of turn."
//
// The Fractal is Deekah's create-then-put shape (socB8FractalWithCounters).
// The combat trigger targets a creature you control as it goes on the
// stack and does both halves to that creature if it is still legal as it
// resolves (CR 608.2b); "it" is the same creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ce1a6c1f-6b4e-4d19-b256-48d073b359b7",
		Name:         "Additive Evolution",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Additive Evolution — create a Fractal and put three +1/+1 counters on it",
				effects.SocB8FractalWithCounters("additive-evolution/fractal", func(*effects.Context) int { return 3 })),
			effects.Targeting(effects.On(game.EventStepBegan, effects.StepBegan(game.StepBeginCombat, true),
				"Additive Evolution — put a +1/+1 counter on target creature you control; it gains vigilance until end of turn",
				additiveEvolutionCombat),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}

// additiveEvolutionCombat is the beginning-of-combat trigger.
func additiveEvolutionCombat(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	if err := (effects.AddCounter{Target: target.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
		return err
	}
	return effects.GrantKeywordUntilEOT{
		Target:   target.ID,
		Keywords: []string{"vigilance"},
		Label:    "Additive Evolution — vigilance until end of turn",
	}.Apply(ctx)
}
