package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigiled Sentinel — Creature — Human Knight {2}{W}, 2/2 (March of the Machine):
//
//	"Backup 1 (When this creature enters, put a +1/+1 counter on target creature.
//	 If that's another creature, it gains the following ability until end of turn.)
//	 Vigilance"
//
// Backup (CR 702.165) is the shared Backup constructor (soc_b6_helpers.go):
// the counters go on the chosen creature, and another creature also gains
// vigilance until end of turn; the card has it itself, declared in PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79c90a68-1c31-4e16-be53-1304132da80f",
		Name:            "Sigiled Sentinel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Backup("Sigiled Sentinel", 1, "", "vigilance"),
		},
	})
}
