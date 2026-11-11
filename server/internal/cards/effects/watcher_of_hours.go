package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watcher of Hours — Creature — Sphinx {5}{U}, 6/6 (Murders at Karlov Manor
// Commander):
//
//	"Flying, ward {3}
//	 Whenever you remove a time counter from this card while it's exiled,
//	 surveil 1.
//	 Suspend 6—{1}{U} (...It has haste.)"
//
// Suspend is the engine's (suspend.go); the trigger works from exile and
// fires once per time counter removed, as Dinosaurs on a Spaceship's does.
// The counters come off only at its controller's upkeep or by an effect
// that player controls, so "you remove" is the card's owner here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b2354092-7a4c-4e69-9c0a-0a2e146dc00c",
		Name:            "Watcher of Hours",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Suspend(6, "{1}{U}")},
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{3}"), "Watcher of Hours — ward {3}"),
			WheneverACounterIsRemovedFromThis(game.CounterTime,
				"Watcher of Hours — surveil 1",
				Do(Surveil{N: 1}), game.ZoneExile),
		},
	})
}
