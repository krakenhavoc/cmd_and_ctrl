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
// S22 sandbox simplification — **the additional end step is not
// implemented.** `Turn.advance` walks a fixed twelve-step sequence by
// index and the engine has no notion of an inserted step; adding one
// is turn-machinery work well outside a flicker PR. So Y'shtola
// blinks once per turn instead of twice, which is strictly weaker
// than printed. Nothing else about the card is approximated.
func init() {
	Register(Spec{
		OracleID:     "a6a7bf77-0560-4572-a826-3bc9df1f78d1",
		Name:         "Y'shtola Rhul",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The extra end step is not created, so it blinks a creature only once per turn instead of twice."},
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
				// "under its owner's control" — leave
				// Controller zero rather than passing the
				// trigger's controller.
				return Flicker{Target: item.Targets[0].ID}.Apply(NewContext(g, item))
			},
		}},
	})
}
