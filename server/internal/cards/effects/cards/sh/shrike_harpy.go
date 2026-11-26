package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrike Harpy — {3}{B}{B} Creature — Harpy 2/2 (Born of the Gods):
//
//	"Flying
//	 Tribute 2 (As this creature enters, an opponent of your choice may put two
//	 +1/+1 counters on it.)
//	 When this creature enters, if tribute wasn't paid, target opponent sacrifices a
//	 creature of their choice."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9c38b337-487c-4cef-843b-b0a2ed3f1b27",
		Name:            "Shrike Harpy",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements:    effects.Tribute("Shrike Harpy", 2),
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenEntersIfTributeWasntPaid("Shrike Harpy — target opponent sacrifices a creature",
				effects.TargetPlayerSacrificesACreature("Shrike Harpy — sacrifice a creature")),
				effects.TargetPlayer("target opponent", effects.Opponent())),
		},
	})
}
