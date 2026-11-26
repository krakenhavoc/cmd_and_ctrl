package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon-Scarred Bear — Creature — Bear {2}{G}, 3/2 (Dragons of Tarkir, S82 older
// sets):
//
//	"Formidable — {1}{G}: Regenerate this creature. Activate only if creatures
//	 you control have total power 8 or greater."
//
// The shared regenerate row (CR 701.19a) with the formidable condition (CR
// 602.5b): the total power of the creatures you control, the Bear included, is
// read as the ability is announced. Once the shield is up it does not depend on
// the total.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f6f49a4-f2ea-4147-a6d3-69cf875c302e",
		Name:         "Dragon-Scarred Bear",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.OldB324FormidableRow(effects.RegenerateThisRow("{1}{G}"), "{1}{G}: Regenerate this creature."),
		},
	})
}
