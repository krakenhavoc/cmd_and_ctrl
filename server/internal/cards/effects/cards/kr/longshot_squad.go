package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Longshot Squad — Creature — Dog Archer {3}{G}, 3/3 (Khans of Tarkir,
// #5554):
//
//	"Outlast {1}{G} ({1}{G}, {T}: Put a +1/+1 counter on this creature.
//	 Outlast only as a sorcery.)
//	 Each creature you control with a +1/+1 counter on it has reach. (A
//	 creature with reach can block creatures with flying.)"
//
// Outlast is the shared sorcery-speed tap-and-pay counter row
// (rprB04Outlast, CR 702.107a). The static is Abzan Battle Priest's
// layer-6 grant over every creature you control with a +1/+1 counter,
// this one included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c153e202-1d93-4044-b457-b8ebd5b36ec3",
		Name:         "Longshot Squad",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.RprB04Outlast("{1}{G}")},
		Static: []game.StaticAbility{
			effects.CreaturesYouControlWithCountersHave(game.CounterPlusOne, false, "reach"),
		},
	})
}
