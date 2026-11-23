package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Child of Alara — Legendary Creature — Avatar {W}{U}{B}{R}{G}, 6/6
// (Conflux, #5587):
//
//	"Trample
//	 When Child of Alara dies, destroy all nonland permanents. They can't
//	 be regenerated."
//
// A dies trigger (CR 603.6c, 603.10a) whose body is Plague Boiler's
// sweep with Damnation's no-regeneration rider (CR 701.19c): every
// nonland permanent on the battlefield as it resolves, anyone's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "70dbe8a9-505d-41c2-9b5b-a991d13ab459",
		Name:            "Child of Alara",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.WhenThisDies("Child of Alara — destroy all nonland permanents",
					effects.Do(effects.DestroyAllMatching{Match: effects.Nonland(), CantBeRegenerated: true})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy}}),
		},
	})
}
