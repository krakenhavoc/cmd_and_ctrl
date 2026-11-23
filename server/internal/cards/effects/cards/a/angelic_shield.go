package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelic Shield — Enchantment {W}{U} (Invasion, #5616):
//
//	"Creatures you control get +0/+1.
//	 Sacrifice this enchantment: Return target creature to its owner's
//	 hand."
//
// Mirari's Wake's anthem shape (a layer-7c TribalAnthem with no tribe,
// yours only) at +0/+1, and a sacrifice ability with the shared
// bounceTheTargetToHand body. The sacrifice is the cost, so the +0/+1
// is gone while the bounce waits on the stack, as the printed ruling
// says. A creature that left in response is no longer a legal target,
// and the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05b020fd-21be-495d-ae45-7de3b1224e6d",
		Name:         "Angelic Shield",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{YoursOnly: true}, 0, 1),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Return target creature to its owner's hand.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}
