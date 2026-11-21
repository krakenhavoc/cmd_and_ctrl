package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lightbringer — Creature — Kor Rebel {2}{W}, 2/2 (Nemesis, #5618):
//
//	"{T}, Sacrifice this creature: Exile target black creature."
//
// Lawbringer's shape with black: a tap-and-sacrifice cost (a {T} cost,
// so not the turn it came under your control, CR 302.6) and the shared
// exileFirstLegalCardTarget body. A creature that stopped being black in
// response is no longer a legal target and is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8656d9c1-d890-4473-9da8-7ec339d7d546",
		Name:         "Lightbringer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Exile target black creature.",
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target black creature", effects.OfColor("B")),
			Effect:  effects.ExileFirstLegalCardTarget,
		}},
	})
}
