package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goldmaw Champion — Creature — Dwarf Warrior {2}{W}, 2/3 (Kaldheim,
// #5505):
//
//	"Boast — {1}{W}: Tap target creature. (Activate only if this
//	 creature attacked this turn and only once each turn.)"
//
// The boast constructor (boast.go, CR 702.142) carries the "attacked
// this turn, once each turn" gate; the body is the shared "Tap target
// creature" (tapChosenPermanent, Clockwork Drawbridge's).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9cbc8765-1735-48ca-bbc2-411aa7da2ed3",
		Name:         "Goldmaw Champion",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.BoastTargeting("{1}{W}: Tap target creature.",
				effects.ManaCost("{1}{W}"), effects.TargetCreature("target creature"), effects.TapChosenPermanent),
		},
	})
}
