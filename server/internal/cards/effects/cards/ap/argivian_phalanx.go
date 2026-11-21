package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Argivian Phalanx — Creature — Human Kor Soldier {5}{W}, 4/4 (Dominaria
// United, S82 DMU batch 4):
//
//	"Affinity for creatures (This spell costs {1} less to cast for each
//	 creature you control.)
//	 Vigilance"
//
// Claws Out's affinity (CR 702.41a) over every creature the caster
// controls, reducing only the generic part; vigilance is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1ca8d2cb-58ba-476b-a4f2-1bf273a54cf3",
		Name:            "Argivian Phalanx",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for creatures", effects.Creature()),
		},
	})
}
