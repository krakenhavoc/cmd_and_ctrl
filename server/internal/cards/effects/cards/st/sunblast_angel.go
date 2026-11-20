package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunblast Angel — Creature — Angel {4}{W}{W}, 4/5:
//
//	"Flying
//	 When this creature enters, destroy all tapped creatures."
//
// Every creature tapped as the ability resolves, yours and the
// Angel itself included.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b34d1df-4024-497d-bd28-a7caa8c3fa24",
		Name:            "Sunblast Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisEnters("Sunblast Angel — destroy all tapped creatures",
					effects.Do(effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.TappedPermanent())})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		},
	})
}
