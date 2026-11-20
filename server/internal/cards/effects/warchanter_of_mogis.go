package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warchanter of Mogis — Creature — Minotaur Shaman {3}{B}{B}, 3/3 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, target
//	 creature you control gains intimidate until end of turn. (A creature
//	 with intimidate can't be blocked except by artifact creatures and/or
//	 creatures that share a color with it.)"
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The target is chosen as the trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9fb5ea12-5b87-462b-b9c0-904cd0f55a9c",
		Name:         "Warchanter of Mogis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(On(game.EventUntapCard, Self,
				"Warchanter of Mogis — target creature you control gains intimidate until end of turn (inspired)",
				neoB07TargetGetsUntilEOT("Warchanter of Mogis", 0, 0, "intimidate")),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
