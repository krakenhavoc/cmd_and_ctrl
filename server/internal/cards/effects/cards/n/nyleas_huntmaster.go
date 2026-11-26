package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nylea's Huntmaster — Creature — Centaur Shaman {3}{G}, 4/3 (Theros Beyond Death, #5514):
//
//	"When this creature enters, target creature you control gets +X/+0 until end
//	 of turn, where X is your devotion to green."
//
// Devotion is counted as the trigger resolves (CR 700.5, CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5ce273f2-a865-4934-afd4-cf6d5c4f1a4f",
		Name:         "Nylea's Huntmaster",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(effects.WhenThisEnters("Nylea's Huntmaster — target creature you control gets +X/+0 until end of turn",
				effects.S82OldB133DevotionBoostFirstTarget("Nylea's Huntmaster — +X/+0", "G", 1, 0)),
				effects.TargetCreature("target creature you control", effects.YouControl())),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0))),
		},
	})
}
