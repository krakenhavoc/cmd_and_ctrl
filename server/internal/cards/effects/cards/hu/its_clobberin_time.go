package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// It's Clobberin' Time! — Sorcery {2}{G}:
//
//	"Choose one —
//	 • Target creature you control deals damage equal to its power to
//	   target creature an opponent controls.
//	 • Destroy target artifact or enchantment.
//	 Rebound"
//
// The bite is HULK SMASH!'s mode body (biteTheModesSecondTarget): the
// power is read as the spell resolves, and nothing is dealt if either
// creature has gone. The rebound cast chooses its mode afresh (CR
// 700.2a). Rebound is the engine's keyword (game/rebound.go, #1854).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a75ef46-c5ae-4b1a-a681-d439b01f6731",
		Name:            "It's Clobberin' Time!",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Modes: effects.ChooseOne(
			effects.ModeDoing("Target creature you control deals damage equal to its power to target creature an opponent controls.",
				effects.Clauses(
					effects.TargetCreature("target creature you control", effects.YouControl()),
					effects.TargetCreature("target creature an opponent controls", effects.OpponentControls()),
				),
				effects.BiteTheModesSecondTarget),
			effects.ModeDoing("Destroy target artifact or enchantment.",
				effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
				effects.DestroyTheModesTarget),
		),
	})
}
