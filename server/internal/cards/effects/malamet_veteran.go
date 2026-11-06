package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Malamet Veteran — Creature — Cat Warrior {4}{G}, 5/4 (The Lost Caverns
// of Ixalan, #4470):
//
//	"Trample
//	 Descend 4 — Whenever this creature attacks, if there are four or more
//	 permanent cards in your graveyard, put a +1/+1 counter on target
//	 creature."
//
// Trample is the printed keyword (CR 702.19). Descend 4 is an ability
// word (CR 207.2c); the "if" is an intervening if (CR 603.4), checked as
// the creature attacks and again as the trigger resolves
// (WheneverThisAttacksIfDescend, Stinging Cave Crawler's row). The target
// is any creature, the Veteran included, and a target gone by resolution
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "25a9c864-59c6-4230-a234-bf78bf1ef24c",
		Name:            "Malamet Veteran",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacksIfDescend(4, "Malamet Veteran — put a +1/+1 counter on target creature",
				PlusOneCounterOnChosenTarget),
				TargetCreature("target creature")),
		},
	})
}
