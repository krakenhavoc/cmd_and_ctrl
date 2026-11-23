package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predatory Sliver — Creature — Sliver {1}{G}, 1/1 (Magic 2014, #5563):
//
//	"Sliver creatures you control get +1/+1."
//
// The modern wording names "Sliver creatures you control": the anthem includes
// this Sliver itself and leaves the Slivers across the table alone
// (TribeFilter.YoursOnly). It is a layer 7c modification, read live, so a
// Sliver that arrives later gets it at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b589e70-9c1d-48dd-a597-479959e692bf",
		Name:         "Predatory Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 1, 1),
		},
	})
}
