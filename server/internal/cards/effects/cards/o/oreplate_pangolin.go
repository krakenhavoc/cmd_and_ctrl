package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Oreplate Pangolin — Artifact Creature — Robot Pangolin {1}{R}, 2/2
// (Edge of Eternities, #4461):
//
//	"Whenever another artifact you control enters, you may pay {1}. If
//	 you do, put a +1/+1 counter on this creature."
//
// Mentor of the Meek's "you may pay {1}" over Kappa Cannoneer's artifact
// watch, "another" so the Pangolin's own entry never counts. Each artifact
// is its own trigger and its own offer; the artifact is read through the
// layered view, so an animated or token artifact counts. The counter goes
// on the Pangolin only if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "72e8c978-f66d-42e1-ac5f-d1c307d7b93c",
		Name:         "Oreplate Pangolin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				return ok && c.IsArtifact()
			}, "Oreplate Pangolin — pay {1} to put a +1/+1 counter on it", func(g *game.Game, item *game.StackItem) error {
				return effects.MayPay{
					Chooser:  item.Controller,
					Cost:     "{1}",
					Question: "Oreplate Pangolin — pay {1} to put a +1/+1 counter on it?",
					Then:     oreplatePangolinPaid,
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}

// oreplatePangolinPaid is "If you do, put a +1/+1 counter on this
// creature", keyed.
var oreplatePangolinPaid = effects.PayUnlessBranches("pay-unless/oreplate-pangolin-counter",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}, nil)
