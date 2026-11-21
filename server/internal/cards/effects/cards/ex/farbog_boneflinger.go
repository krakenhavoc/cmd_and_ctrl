package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Farbog Boneflinger — Creature — Zombie {4}{B}, 2/2 (Dark Ascension, #5571):
//
//	"When this creature enters, target creature gets -2/-2 until end
//	 of turn."
//
// A layer 7c -2/-2 on the still-legal target (CR 608.2b), gone at cleanup.
// The target is any creature, the Boneflinger included when it is the only
// one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "46027ca0-9d06-42b3-8012-124861c6b558",
		Name:         "Farbog Boneflinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Farbog Boneflinger — target creature gets -2/-2 until end of turn",
				effects.BoostTargetUntilEOT("Farbog Boneflinger — -2/-2", -2, -2)),
				effects.TargetCreature("target creature")), effects.ForTargets(effects.HarmUntilEOT(0, -2, -2))),
		},
	})
}
