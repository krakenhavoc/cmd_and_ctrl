package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spiteful Squad — Creature — Human Warlock {2}{W}{B}, 0/0 (Strixhaven:
// School of Mages, #5502):
//
//	"Deathtouch
//	 This creature enters with two +1/+1 counters on it.
//	 When this creature dies, put its counters on target creature you
//	 control."
//
// The entry counters are an enters-with-counters replacement (CR 614.1c).
// The counters moved are those the creature last had (CR 603.10a), put on
// through the ordinary counter door so Hardened Scales applies, and only if
// the target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "41ee78ce-a5ef-4f88-b479-2eb6651123f7",
		Name:            "Spiteful Squad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Replacements: []game.ReplacementEffect{
			effects.B10EntersWithCounters(game.CounterPlusOne, 2, "Spiteful Squad: enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisDies("Spiteful Squad — put its counters on target creature you control",
				effects.PutItsCountersOnTargetCreature),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
	})
}
