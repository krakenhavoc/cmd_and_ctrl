package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Loxodon Sergeant — Creature — Elephant Soldier {3}{W}, 3/3 (War of the Spark, #5520):
//
//	"Vigilance
//	 When this creature enters, other creatures you control gain vigilance
//	 until end of turn."
//
// Loyal Unicorn's grant over the other creatures you control as the trigger
// resolves; one that enters later does not get it (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "126038e3-3e7b-4547-83e7-a7f6ec66f782",
		Name:            "Loxodon Sergeant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Loxodon Sergeant — other creatures you control gain vigilance until end of turn",
				effects.S82OldB171OtherCreaturesGainUntilEOT("Loxodon Sergeant", "vigilance")),
		},
	})
}
