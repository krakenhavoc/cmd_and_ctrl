package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Megantic Sliver — Creature — Sliver {5}{G}, 3/3 (Magic 2014, #5563):
//
//	"Sliver creatures you control get +3/+3."
//
// The modern wording names "Sliver creatures you control": the anthem includes
// this Sliver itself and leaves the Slivers across the table alone
// (TribeFilter.YoursOnly). It is a layer 7c modification, read live, so a
// Sliver that arrives later gets it at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "52e8ec03-23ed-43a0-bce2-ad0cd5c4d737",
		Name:         "Megantic Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, 3, 3),
		},
	})
}
