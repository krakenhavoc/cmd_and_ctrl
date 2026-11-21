package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deepwater Hypnotist — Creature — Merfolk Wizard {1}{U}, 2/1 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, target
//	 creature an opponent controls gets -3/-0 until end of turn."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The target is chosen as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e30c639a-d46c-4488-bfdd-d4282046153c",
		Name:         "Deepwater Hypnotist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventUntapCard, effects.Self,
				"Deepwater Hypnotist — target creature an opponent controls gets -3/-0 until end of turn (inspired)",
				effects.NeoB07TargetGetsUntilEOT("Deepwater Hypnotist", -3, 0)),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
