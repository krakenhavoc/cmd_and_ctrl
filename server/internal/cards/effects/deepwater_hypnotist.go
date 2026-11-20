package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e30c639a-d46c-4488-bfdd-d4282046153c",
		Name:         "Deepwater Hypnotist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventUntapCard, Self,
				"Deepwater Hypnotist — target creature an opponent controls gets -3/-0 until end of turn (inspired)",
				neoB07TargetGetsUntilEOT("Deepwater Hypnotist", -3, 0)),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
