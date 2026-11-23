package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Charge — Sorcery {2}{B} (Portal Three Kingdoms, #5622):
//
//	"Creatures you control get +2/+0 until end of turn."
//
// Warrior's Charge's body at +2/+0 (creaturesYouControlGetUntilEOT): the
// creatures you control as it resolves (CR 611.2c), pumped until
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "26074e44-cf2b-459f-9a5d-7eed231ed2b4",
		Name:         "Desperate Charge",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.CreaturesYouControlGetUntilEOT(2, 0, "Desperate Charge — creatures you control get +2/+0"),
	})
}
