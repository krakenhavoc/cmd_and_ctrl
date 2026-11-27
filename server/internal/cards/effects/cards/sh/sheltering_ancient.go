package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sheltering Ancient — Creature — Treefolk {1}{G}, 5/5 (Coldsnap, S82 OLD b558):
//
//	"Trample
//	 Cumulative upkeep—Put a +1/+1 counter on a creature an opponent controls."
//
// One counter per age counter, each on a creature of the payer's choice (the
// same one twice if they like; CR 702.24a), paid together or not at all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4554ce06-4992-4689-8a40-558a1545b968",
		Name:            "Sheltering Ancient",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeepPaying("Sheltering Ancient — cumulative upkeep: put a +1/+1 counter on a creature an opponent controls",
				effects.CounterOnOpponentCreaturePayment(1, game.CounterPlusOne, "creature", "creatures", effects.QueryTypes("creature")),
				"creature", "creatures"),
		},
	})
}
