package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Final-Sting Faerie — Creature — Faerie Assassin {3}{B}, 2/2
// (Morningtide, #5592):
//
//	"Flying
//	 When this creature enters, destroy target creature that was dealt
//	 damage this turn."
//
// Fatal Blow's target clause (DealtDamageThisTurn) on Shriekmaw's enters
// trigger; with no damaged creature the trigger is removed for want of a
// target (CR 603.3d). Unlike Fatal Blow it allows regeneration.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "299765d8-72b9-4ada-a607-efb36aa0f5db",
		Name:            "Final-Sting Faerie",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Final-Sting Faerie — destroy target creature that was dealt damage this turn", effects.DestroyChosenPermanent),
				effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn())),
		},
	})
}
