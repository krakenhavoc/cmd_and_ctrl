package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Floodwater Dam — Artifact {3} (Alliances, #5634):
//
//	"{X}{X}{1}, {T}: Tap X target lands."
//
// X is announced with the activation and pays two of the three cost slots
// ({X}{X}{1}, Treasure Vault's shape); the ability has exactly X targets, so
// X = 0 targets nothing and taps nothing. Each is a land at resolution
// (CR 608.2b) and is tapped (a tapped one stays tapped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a3947473-0f57-4937-bb1c-f20969144415",
		Name:         "Floodwater Dam",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{X}{X}{1}, {T}: Tap X target lands.",
			Purpose: game.Purpose{Answers: game.AnswerRestrict},
			Cost:    effects.Plus(effects.ManaCost("{X}{X}{1}"), effects.TapCost()),
			Targets: effects.TargetsCountedByX(effects.TargetPermanent("X target lands", effects.Land())),
			Effect:  effects.OldB815TapTheTargets,
		}},
	})
}
