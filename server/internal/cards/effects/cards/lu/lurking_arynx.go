package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurking Arynx — Creature — Cat Beast {4}{G}, 3/5 (Dragons of Tarkir, S82 older
// sets):
//
//	"Formidable — {2}{G}: Target creature blocks this creature this turn if able.
//	 Activate only if creatures you control have total power 8 or greater."
//
// Burning-Tree Bloodscale's requirement (TargetBlocksThisCreature), which waits
// for the Arynx to attack this turn and asks nothing if it never does (CR 509.1c),
// behind the formidable activation condition (CR 602.5b). The target is read
// again as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "57303267-7e69-4bd3-8e1f-a528d41121a1",
		Name:         "Lurking Arynx",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{effects.OldB324FormidableRow(effects.ActivatedAbility{
			Cost:    effects.ManaCost("{2}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.TargetBlocksThisCreature,
		}, "{2}{G}: Target creature blocks this creature this turn if able.")},
	})
}
