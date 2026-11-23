package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loyal Guardian — Creature — Rhino {4}{G}, 4/4 (Marvel Super Heroes
// Commander, #4455):
//
//	"Trample
//	 Lieutenant — At the beginning of combat on your turn, if you control
//	 your commander, put a +1/+1 counter on each creature you control."
//
// An intervening "if" (CR 603.4): the trigger is checked as combat begins
// and again as it resolves. The commander is one the controller owns and
// controls (b18ControlsYourCommander, Loyal Unicorn's read).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "de1cfa83-2049-48ab-b093-2f120a3ef4f2",
		Name:            "Loyal Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventStepBegan, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.StepBegan(game.StepBeginCombat, true)(ev, source, lki, g) &&
					effects.B18ControlsYourCommander(g, source.Controller)
			}, "Loyal Guardian — Lieutenant: put a +1/+1 counter on each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					if !effects.B18ControlsYourCommander(g, item.Controller) {
						return nil
					}
					return effects.B11PutCounterOnEachCreatureYouControl(g, item)
				}),
		},
	})
}
