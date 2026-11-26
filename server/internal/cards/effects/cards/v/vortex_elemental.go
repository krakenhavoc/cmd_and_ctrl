package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vortex Elemental — Creature — Elemental {U}, 0/1 (Born of the Gods, S82 OLD
// b378, #5559):
//
//	"{U}: Put this creature and each creature blocking or blocked by it on top
//	 of their owners' libraries, then those players shuffle.
//	 {3}{U}{U}: Target creature blocks this creature this turn if able."
//
// The first ability tucks the source and its combat partners (read from combat
// as it resolves) as one simultaneous exit, then every owner shuffles
// (S82OldB378VortexTuck; the owners ride the tail, so a commander that went to
// the command zone still has its owner shuffle). The second is a block
// requirement on the target (Monstrous Step's BlocksAttackerUntilEOT).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4acb586c-dd87-452c-b3d8-728a4396aae6",
		Name:         "Vortex Elemental",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}: Put this creature and each creature blocking or blocked by it on top of their owners' libraries, then those players shuffle.",
				Cost:    effects.ManaCost("{U}"),
				Purpose: game.Purpose{Answers: game.AnswerSurvive},
				Effect:  effects.S82OldB378VortexTuck,
			},
			{
				Label:   "{3}{U}{U}: Target creature blocks this creature this turn if able.",
				Cost:    effects.ManaCost("{3}{U}{U}"),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.S82OldB378TargetBlocksThis,
			},
		},
	})
}
