package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mist Raven — {2}{U}{U} Creature — Bird 2/2 (Avacyn Restored, #5570):
//
//	"Flying
//	 When this creature enters, return target creature to its owner's
//	 hand."
//
// Flying and Man-o'-War's enters bounce of a target creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "12654457-3aae-4046-a976-8e41c3f78927",
		Name:            "Mist Raven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Mist Raven — return target creature to its owner's hand",
				effects.BounceTheTargetToHand), effects.TargetCreature("target creature")),
		},
	})
}
