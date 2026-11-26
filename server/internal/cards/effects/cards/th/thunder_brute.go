package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunder Brute — {4}{R}{R} Creature — Cyclops 5/5 (Born of the Gods):
//
//	"Trample
//	 Tribute 3 (As this creature enters, an opponent of your choice may put three
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, it gains haste until end of
//	 turn."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ee1d97e9-847d-4295-a9c1-7ce8006fb0bd",
		Name:            "Thunder Brute",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Replacements:    effects.Tribute("Thunder Brute", 3),
		Triggered: []game.TriggeredAbility{
			effects.WhenEntersIfTributeWasntPaid("Thunder Brute — it gains haste until end of turn",
				effects.ThisGainsKeywordUntilEndOfTurn("haste", "Thunder Brute — gains haste until end of turn")),
		},
	})
}
