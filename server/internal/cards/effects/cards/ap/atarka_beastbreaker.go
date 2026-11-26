package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atarka Beastbreaker — Creature — Human Warrior {1}{G}, 2/2 (Dragons of Tarkir,
// S82 older sets):
//
//	"Formidable — {4}{G}: This creature gets +4/+4 until end of turn. Activate
//	 only if creatures you control have total power 8 or greater."
//
// The shared self-pump row (ThisGetsRow) behind the formidable activation
// condition (CR 602.5b), the total read as the ability is announced.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fdd4ef61-607a-4090-9380-dabbbb448519",
		Name:         "Atarka Beastbreaker",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.OldB324FormidableRow(effects.ThisGetsRow("{4}{G}", 4, 4, "Atarka Beastbreaker"), "{4}{G}: This creature gets +4/+4 until end of turn."),
		},
	})
}
