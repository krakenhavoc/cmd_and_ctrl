package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Cacodemon — Creature — Demon {7}{B}{B}{B}, 8/8 (Commander 2011, S82
// older sets batch 453, #5574):
//
//	"When this creature enters, if you cast it from your hand, destroy all
//	 creatures your opponents control, then tap all other creatures you
//	 control."
//
// Wakening Sun's Avatar's cast-from-hand trigger, rechecked as it resolves
// (CR 603.4). The destroy is one simultaneous event; the tap then reaches
// every other creature the Cacodemon's controller still controls.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1d34e78e-4fe9-44e1-88d3-a98fd64555a9",
		Name:         "Dread Cacodemon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.B06SelfETB(ev, source, lki, g) && effects.B30CastFromHand(g, source.InstanceID)
			}, "Dread Cacodemon — destroy all creatures your opponents control, then tap all other creatures you control",
				effects.S82OldB453DestroyThenTapOthers),
		},
	})
}
