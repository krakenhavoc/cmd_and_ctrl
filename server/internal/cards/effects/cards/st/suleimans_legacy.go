package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suleiman's Legacy — Enchantment {R}{W} (Visions, S82 older sets batch 790,
// #5631):
//
//	"When this enchantment enters, destroy all Djinns and Efreets. They can't
//	 be regenerated.
//	 Whenever a Djinn or Efreet enters, destroy it. It can't be regenerated."
//
// Two triggers. The first sweeps every Djinn and Efreet on the battlefield,
// whoever controls them (a changeling is both). The second fires for any
// permanent of either type entering, under any player's control, and destroys
// it as it resolves if it is still on the battlefield. The type words are read
// through the Legacy's own text (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9888d67d-fcad-427f-8ba7-b6d421b3d88e",
		Name:         "Suleiman's Legacy",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Suleiman's Legacy — destroy all Djinns and Efreets; they can't be regenerated",
				effects.Old790DestroyAllOfTypes("Djinn", "Efreet")),
			effects.On(game.EventETB, effects.Old790CreatureEnteredOfTypes("Djinn", "Efreet"),
				"Suleiman's Legacy — destroy it; it can't be regenerated", effects.Old790DestroyTheEnteredCreature),
		},
	})
}
