package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wizened Cenn — Creature — Kithkin Cleric {W}{W}, 2/2 (Lorwyn, #5593):
//
//	"Other Kithkin creatures you control get +1/+1."
//
// A lord (tribal.go): "other" leaves the source out and "you control" is
// printed, so only your other Kithkin creatures get it. The boost is layer 7c
// (CR 613.4c); a changeling counts (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5905ef3-6565-462f-a015-2bb5b5b09eff",
		Name:         "Wizened Cenn",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Kithkin"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
