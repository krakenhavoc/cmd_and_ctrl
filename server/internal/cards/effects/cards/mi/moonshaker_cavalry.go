package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonshaker Cavalry — Creature — Spirit Knight {5}{W}{W}{W}, 6/6
// (Secrets of Strixhaven Commander, #4834):
//
//	"Flying
//	 When this creature enters, creatures you control gain flying and
//	 get +X/+X until end of turn, where X is the number of creatures you
//	 control."
//
// Craterhoof Behemoth's trigger with flying for trample
// (socB9TeamKeywordAndPump). X is counted as the trigger resolves, the
// Cavalry included, and neither X nor the affected set changes later in
// the turn (the rulings, CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3f16f34-d78c-4b61-821b-05fc6bd0da37",
		Name:            "Moonshaker Cavalry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Moonshaker Cavalry — flying and +X/+X", effects.SocB9TeamKeywordAndPump("Moonshaker Cavalry", "flying")),
		},
	})
}
