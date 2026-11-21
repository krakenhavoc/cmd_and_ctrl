package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	landfall := effects.Landfall("Elven Raft-Steerer — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	landfall.Modes = effects.ChooseOne(
		effects.ModeDoing("Tap target creature an opponent controls.",
			effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
			effects.TapTheModesTarget),
		effects.ModeDoing("Untap target creature you control.",
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.UntapTheModesTarget),
	)
	effects.Register(effects.Spec{
		OracleID:     "d0f62148-1015-49b8-b61f-a4332325a6b0",
		Name:         "Elven Raft-Steerer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{landfall},
	})
}
