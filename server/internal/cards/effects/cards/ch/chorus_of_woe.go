package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chorus of Woe — Sorcery {B} (Portal Second Age, #5625):
//
//	"Creatures you control get +1/+0 until end of turn."
//
// Scare Tactics' body as a sorcery (creaturesYouControlGetUntilEOT): the
// creatures you control as it resolves are the ones pumped (CR 611.2c),
// and the bonus ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e175a3c1-df6c-4aa7-a4c9-893b466ad9e0",
		Name:         "Chorus of Woe",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.CreaturesYouControlGetUntilEOT(1, 0, "Chorus of Woe — creatures you control get +1/+0"),
	})
}
