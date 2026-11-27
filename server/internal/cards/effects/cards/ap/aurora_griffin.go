package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aurora Griffin — Creature — Griffin {3}{W}, 2/2 (Planeshift, S82 OLD #5615):
//
//	"Flying
//	 {W}: Target permanent becomes white until end of turn."
//
// The colour change is Eight-and-a-Half-Tails': a layer-5 set (CR 613.1e), so
// the permanent is white and no other colour, on a target still legal as the
// ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4d74c735-8035-4c42-bfb1-1bc97e5bdd9a",
		Name:            "Aurora Griffin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}: Target permanent becomes white until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.ManaCost("{W}"),
			Targets: effects.TargetPermanent("target permanent"),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				return effects.Gap2B05TargetBecomesColor(ctx, []string{"W"}, "Aurora Griffin — white until end of turn")
			})),
		}},
	})
}
