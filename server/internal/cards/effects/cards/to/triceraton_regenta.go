package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triceraton Regenta — Creature — Dinosaur Noble {2}{W}, 2/3 (TMNT
// Commander, #5011):
//
//	"Vigilance
//	 Whenever this creature attacks, untap up to one other target
//	 creature."
//
// Vigilance rides PrintedKeywords. The attack trigger takes zero or one
// creature other than the Regenta (Another), re-checked as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2650d2d5-576e-4b9a-95ac-0f5307f1580f",
		Name:            "Triceraton Regenta",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Triceraton Regenta — untap up to one other target creature", effects.UntapEachStillLegalTarget),
				effects.Another(effects.TargetCreature("up to one other target creature")).WithCount(0, 1)),
		},
	})
}
