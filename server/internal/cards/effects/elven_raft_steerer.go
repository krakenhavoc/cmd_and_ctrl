package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elven Raft-Steerer — Creature — Elf Pilot {2}{U}, 3/2 (The Hobbit, HOB):
//
//	"Landfall — Whenever a land you control enters, choose one —
//	 • Tap target creature an opponent controls.
//	 • Untap target creature you control."
//
// Felidar Retreat's shape: a landfall trigger that does nothing itself,
// each bullet a ModeDoing body with its own target (picked as the trigger
// goes on the stack, CR 603.3c). A target that left or became illegal by
// resolution is skipped (CR 608.2b).
//
// No simplification.
func init() {
	landfall := Landfall("Elven Raft-Steerer — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	landfall.Modes = ChooseOne(
		ModeDoing("Tap target creature an opponent controls.",
			TargetCreature("target creature an opponent controls", OpponentControls()),
			TapTheModesTarget),
		ModeDoing("Untap target creature you control.",
			TargetCreature("target creature you control", YouControl()),
			UntapTheModesTarget),
	)
	Register(Spec{
		OracleID:     "d0f62148-1015-49b8-b61f-a4332325a6b0",
		Name:         "Elven Raft-Steerer",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{landfall},
	})
}
