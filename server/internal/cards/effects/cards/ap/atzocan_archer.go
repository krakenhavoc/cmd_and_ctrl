package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atzocan Archer — Creature — Human Archer {2}{G}, 1/4 (Ixalan, #5532):
//
//	"Reach
//	 When this creature enters, you may have it fight another target
//	 creature."
//
// Somberwald Stag's trigger with the target widened to any other creature
// (yours too): the "may" is asked as the trigger resolves, and the fight
// happens only while the Archer is still on the battlefield and the
// target is still legal (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "aed10a6c-ef93-46a7-a5d7-c9cd2a4fd91a",
		Name:            "Atzocan Archer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Atzocan Archer — it fights another target creature",
					effects.B30SourceFightsFirstLegalTarget),
				effects.Another(effects.TargetCreature("another target creature"))),
				"Atzocan Archer — have it fight another target creature?"),
		},
	})
}
