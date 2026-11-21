package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chronomancer — Artifact Creature — Necron Wizard {1}{B}, 1/1 (Warhammer
// 40,000 Commander, S82 GAP batch 187, #5948):
//
//	"Flying
//	 Atomic Transmutation — {1}, {T}, Sacrifice another artifact: Draw a
//	 card.
//	 Unearth {2}{B}"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9833e3bd-6e7f-41b7-bf7e-ffb74ff17844",
		Name:            "Chronomancer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}, {T}, Sacrifice another artifact: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeAnotherN(1, "another artifact", effects.Artifact())),
				Effect:  effects.Do(effects.DrawCards{N: 1}),
			},
			effects.Unearth("{2}{B}"),
		},
	})
}
