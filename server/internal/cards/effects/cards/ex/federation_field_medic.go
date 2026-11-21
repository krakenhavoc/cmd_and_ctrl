package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Federation Field Medic — {3}{W} Creature — Human Doctor 3/2 (Star Trek,
// #4973):
//
//	"Vigilance
//	 When this creature enters, choose one —
//	 • Put a +1/+1 counter on target creature. It gains vigilance until
//	   end of turn.
//	 • Draw a card if you control a permanent with a counter on it."
//
// A modal enters trigger, Charming Prince's shape: the mode is chosen as
// the trigger goes on the stack, and only the first bullet targets. The
// second bullet's "if" is part of the effect, checked as it resolves
// (CR 608.2c): any kind of counter on any permanent you control counts.
//
// No simplification.
func init() {
	medic := effects.WhenThisEnters("Federation Field Medic — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	medic.Modes = effects.ChooseOne(
		effects.ModeDoing("Put a +1/+1 counter on target creature. It gains vigilance until end of turn.",
			effects.TargetCreature("target creature"),
			func(_ *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok {
					return nil
				}
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"vigilance"},
					Label: "Federation Field Medic — vigilance until end of turn"}.Apply(ctx)
			}),
		effects.ModeDoing("Draw a card if you control a permanent with a counter on it.", nil,
			func(item *game.StackItem, ctx *effects.Context, _ int) error {
				if !effects.TrekB2ControlsAPermanentWithACounter(ctx.Game, item.Controller) {
					return nil
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:        "3759f1de-eda1-4f09-b12b-ea5f5fc731d7",
		Name:            "Federation Field Medic",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered:       []game.TriggeredAbility{medic},
	})
}
