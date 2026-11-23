package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cragsmasher Yeti — Creature — Yeti {4}{R}{R}, 4/2 (March of the Machine):
//
//	"Mountaincycling {2}
//	 Backup 2 (When this creature enters, put two +1/+1 counters on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Trample"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// trample until end of turn; the card has it itself, declared in PrintedKeywords. Mountaincycling is Typecycling (CR 702.29e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "57120486-0316-4591-abca-438ab8283d3f",
		Name:            "Cragsmasher Yeti",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Cragsmasher Yeti", 2, "", "trample"),
		},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Mountaincycling", "{2}", "a Mountain card", effects.IsLandWithSubtype("Mountain")),
		},
	})
}
