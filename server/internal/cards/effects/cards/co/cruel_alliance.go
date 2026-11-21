package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cruel Alliance — {2}{B} Sorcery:
//
//	"Teamwork 2 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 2 or more.)
//	 Exile target creature with mana value 3 or less. If this spell was
//	 cast using teamwork, instead exile target creature and you gain 3
//	 life."
//
// #1703's builder could not register this: the teamwork guard refused
// a target-clause rewrite on the cost. #1716 allows it, so the
// teamwork cast is announced and re-checked (CR 608.2b) under "target
// creature" and an ordinary cast under the printed mana-value ceiling.
//
// The life is part of the teamwork branch's one sentence, so a spell
// whose target became illegal fizzles (CR 608.2b) and gains nothing —
// which is what an OnResolve that is never run gives for free.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "6897f9e0-f654-4c0a-9fda-2ad4e264bf9a",
		Name:          "Cruel Alliance",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.WhenPaid(effects.Teamwork(2), effects.TargetCreature("target creature"))},
		Targets:       effects.TargetCreature("target creature with mana value 3 or less", effects.ManaValueLE(3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
				return err
			}
			if ctx.UsedTeamwork() {
				return effects.GainLife{Amount: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
