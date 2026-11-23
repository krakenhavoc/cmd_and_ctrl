package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning-Yard Trainer — Creature — Human Knight {4}{R}, 3/3 (Throne of Eldraine, #5516):
//
//	"Trample, haste
//	 When this creature enters, another target Knight you control gets +2/+2
//	 and gains trample and haste until end of turn."
//
// "Another" is object identity (Another), never a name.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6ea7a921-45b2-4589-833b-75b6dc6aed07",
		Name:            "Burning-Yard Trainer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Burning-Yard Trainer — another target Knight you control gets +2/+2 and gains trample and haste until end of turn",
					effects.PumpAndGrantFirstTarget(2, 2, "Burning-Yard Trainer — +2/+2, trample and haste", "trample", "haste")),
				effects.Another(effects.TargetCreature("another target Knight you control", effects.HasSubtype("Knight"), effects.YouControl()))),
		},
	})
}
