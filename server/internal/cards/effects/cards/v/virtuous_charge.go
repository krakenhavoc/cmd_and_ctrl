package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Virtuous Charge — Sorcery {2}{W} (Portal Three Kingdoms, #5622):
//
//	"Creatures you control get +1/+1 until end of turn."
//
// Warrior's Charge's body (creaturesYouControlGetUntilEOT): the creatures
// you control as it resolves (CR 611.2c), pumped until cleanup
// (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5315ed83-cb93-4522-b439-3784e7f31e37",
		Name:         "Virtuous Charge",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.CreaturesYouControlGetUntilEOT(1, 1, "Virtuous Charge — creatures you control get +1/+1"),
	})
}
