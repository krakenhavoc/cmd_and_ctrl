package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lyra, Archangel of Dawn — Legendary Creature — Angel Knight {2}{W},
// 3/3:
//
//	"Flying
//	 Whenever you gain life, put a +1/+1 counter on each Angel you
//	 control."
//
// Archangel of Thune's lifegain trigger over the Angels instead of every
// creature; Lyra is an Angel herself. The set is snapshotted before any
// counter lands, and one lifegain event is one trigger however much life
// it was.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0fca3328-484f-4050-925e-91840710b5d6",
		Name:            "Lyra, Archangel of Dawn",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Lyra, Archangel of Dawn — a +1/+1 counter on each Angel you control", effects.RfCreatureCPutCounterOnEachAngel),
		},
	})
}
