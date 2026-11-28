package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Heat Stroke — Enchantment {2}{R} (Weatherlight, S82 older sets batch 779, #5629):
//
//	"At end of combat, destroy each creature that blocked or was blocked this turn."
//
// Triggers at the beginning of every end of combat step (an extra combat's too, CR 511.2). The set is read as the
// trigger resolves from the turn's block declarations, by objects (CR 400.7), and covers every creature, whoever
// controls it; those still on the battlefield are destroyed together (CR 701.8).
//
// No simplification.
func init() {
	const label = "Heat Stroke — destroy each creature that blocked or was blocked this turn"
	effects.Register(effects.Spec{
		OracleID:     "a8127c04-f67f-4780-86bf-bb24480ec51a",
		Name:         "Heat Stroke",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.AtEachStep(game.StepEndCombat, label, effects.S82OldB779DestroyCreaturesThatBlockedOrWereBlocked),
		},
	})
}
