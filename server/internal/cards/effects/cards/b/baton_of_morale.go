package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Baton of Morale — Artifact {2} (Ice Age, S82 OLD #5636):
//
//	"{2}: Target creature gains banding until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "879011c6-550e-4e33-a664-fffb655321d9",
		Name:         "Baton of Morale",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: Target creature gains banding until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "banding")),
			Cost:    effects.ManaCost("{2}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT(game.KeywordBanding, "Baton of Morale — banding"),
		}},
	})
}
