package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanguine Savior — Creature — Vampire Cleric {1}{W}{B}, 2/1 (Murders at
// Karlov Manor, #4469):
//
//	"Flying, lifelink
//	 Disguise {W/B}{W/B} (You may cast this card face down for {3} as a
//	 2/2 creature with ward {2}. Turn it face up any time for its
//	 disguise cost.)
//	 When this creature is turned face up, another target creature you
//	 control gains lifelink until end of turn."
//
// Disguise is Alley Assailant's, and the face-up trigger (CR 708.8)
// takes Flesh Burrower's "another target creature you control" clause:
// the Savior itself is not a legal pick. A target that left in response
// gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	const label = "Sanguine Savior — another target creature you control gains lifelink until end of turn"
	effects.Register(effects.Spec{
		OracleID:         "cc3a6c83-a671-4ce8-b98e-434989fb6778",
		Name:             "Sanguine Savior",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"flying", "lifelink"},
		AlternativeCosts: []game.AlternativeCost{effects.Disguise("{W/B}{W/B}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisIsTurnedFaceUp(label, effects.JmpB7FirstLegalTargetGains(label, "lifelink")),
				effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink"))),
		},
	})
}
