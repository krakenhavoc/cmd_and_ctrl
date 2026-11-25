package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thundering Ceratok — Creature — Rhino {4}{G}, 4/5 (War of the Spark, #5520):
//
//	"Trample
//	 When this creature enters, other creatures you control gain trample
//	 until end of turn."
//
// Loxodon Sergeant's grant with trample.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b7a3a8a9-96ca-48f7-8354-345d9ccfcb77",
		Name:            "Thundering Ceratok",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thundering Ceratok — other creatures you control gain trample until end of turn",
				effects.S82OldB171OtherCreaturesGainUntilEOT("Thundering Ceratok", "trample")),
		},
	})
}
