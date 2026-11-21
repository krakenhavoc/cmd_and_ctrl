package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aggressive Mammoth — Creature — Elephant {3}{G}{G}{G}, 8/8
// (Foundations, #4465):
//
//	"Trample
//	 Other creatures you control have trample."
//
// Surrak Dragonclaw's last line: a layer-6 grant over the other
// creatures its controller controls. Trample rides PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "032b5fc4-5aca-41b6-9bf6-2c1ed0018968",
		Name:            "Aggressive Mammoth",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Others: true, YoursOnly: true}, "trample"),
		},
	})
}
