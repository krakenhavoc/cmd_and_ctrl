package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lurking Deadeye — {3}{B} Creature — Human Assassin 4/2 (Jumpstart 2022, #5948):
//
//	"Flash
//	 When this creature enters, destroy target creature that was dealt
//	 damage this turn."
//
// Fatal Blow's target predicate (DealtDamageThisTurn) on an enters
// trigger; the target is re-checked as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d1dd27eb-ba7b-4d5d-a36c-3f350df2ad51",
		Name:            "Lurking Deadeye",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Lurking Deadeye — destroy target creature that was dealt damage this turn", effects.DestroyFirstLegalCardTarget),
				effects.TargetCreature("target creature that was dealt damage this turn", effects.DealtDamageThisTurn())),
		},
	})
}
