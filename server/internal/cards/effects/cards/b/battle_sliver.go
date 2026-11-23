package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battle Sliver — Creature — Sliver {4}{R}, 3/3 (Magic 2014, #5563):
//
//	"Sliver creatures you control get +2/+0."
//
// The modern wording names "Sliver creatures you control": the anthem includes
// this Sliver itself and leaves the Slivers across the table alone
// (TribeFilter.YoursOnly). It is a layer 7c modification, read live, so a
// Sliver that arrives later gets it at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d513e5df-31b0-4756-bf47-403460d0068e",
		Name:         "Battle Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 2, 0),
		},
	})
}
