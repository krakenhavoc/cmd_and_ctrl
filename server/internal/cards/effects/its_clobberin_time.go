package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "4a75ef46-c5ae-4b1a-a681-d439b01f6731",
		Name:            "It's Clobberin' Time!",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordRebound},
		Modes: ChooseOne(
			ModeDoing("Target creature you control deals damage equal to its power to target creature an opponent controls.",
				Clauses(
					TargetCreature("target creature you control", YouControl()),
					TargetCreature("target creature an opponent controls", OpponentControls()),
				),
				biteTheModesSecondTarget),
			ModeDoing("Destroy target artifact or enchantment.",
				TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
				DestroyTheModesTarget),
		),
	})
}
