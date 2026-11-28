package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tolarian Entrancer — Creature — Human Wizard {1}{U}, 1/1 (Weatherlight, S82 older sets batch 779, #5629):
//
//	"Whenever this creature becomes blocked by a creature, gain control of that creature at end of
//	 combat."
//
// One trigger for each creature that blocks it (CR 509.3d, Barrow-Blade's EventBlock pairing): each
// schedules a delayed trigger at the beginning of the end of combat step (CR 511.2) for that blocker, as the
// object it was when it blocked (CR 400.7). The control change has no duration (CR 611.2) and is the
// Entrancer's controller's, even if the Entrancer has left by then.
//
// No simplification.
func init() {
	const label = "Tolarian Entrancer — gain control of the creature that blocked it at end of combat"
	effects.Register(effects.Spec{
		OracleID:     "50c8c7e2-1b45-44b2-a50b-3b514d4c4c42",
		Name:         "Tolarian Entrancer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBlock, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBlock && ev.Target == source.InstanceID
			}, label, effects.S82OldB779GainControlOfTheBlockerAtEndOfCombat(label)),
		},
	})
}
