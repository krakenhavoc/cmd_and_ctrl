package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nezumi Bone-Reader — {1}{B} Creature — Rat Shaman 1/1 (Jumpstart 2022, #5948):
//
//	"{B}, Sacrifice a creature: Target player discards a card. Activate
//	 only as a sorcery."
//
// The discarding player picks the card (CR 701.9a). The sacrificed
// creature may be the Bone-Reader itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c77f8fee-8d2d-4cec-948f-2ca3e5b64070",
		Name:         "Nezumi Bone-Reader",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{B}, Sacrifice a creature: Target player discards a card. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{B}"), effects.SacrificeACreature()),
			SorcerySpeed: true,
			Targets:      effects.TargetPlayer("target player"),
			Effect:       effects.TargetPlayerDiscards(1),
		}},
	})
}
