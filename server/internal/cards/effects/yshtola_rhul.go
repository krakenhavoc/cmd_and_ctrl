package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Y'shtola Rhul — "At the beginning of your end step, exile target
// creature you control, then return it to the battlefield under its
// owner's control. Then if it's the first end step of the turn,
// there is an additional end step after this step."
//
// The first sentence is an immediate flicker on an end-step trigger:
// exile and return resolve together, so the creature is a new object
// with every ETB re-fired, but there is no window in which the board
// is missing it.
//
// The second is CR 500.9 through the turn plan (ADR 0059 sub-PR 2b,
// #753): "the first end step of the turn" is Turn.StepOrdinal, so the
// trigger fires again in the added end step, blinks again, and adds
// nothing more. A delayed "at the beginning of the next end step"
// trigger created in the first end step fires in the added one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6a7bf77-0560-4572-a826-3bc9df1f78d1",
		Name:         "Y'shtola Rhul",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			Key:     "Y'shtola Rhul — blink a creature you control",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				// "your end step" — not every end step.
				return ev.Actor == source.Controller
			},
			Targets: TargetCreature("target creature you control", YouControl()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if len(item.Targets) == 0 {
					return nil
				}
				ctx := NewContext(g, item)
				// "under its owner's control" — leave
				// Controller zero rather than passing the
				// trigger's controller.
				if err := (Flicker{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
					return err
				}
				if g.Turn.Step != game.StepEnd || !g.IsFirstStepOfItsKindForEffect() {
					return nil
				}
				return AddStepAfterThisStep{Step: game.StepEnd}.Apply(ctx)
			},
		}},
	})
}
