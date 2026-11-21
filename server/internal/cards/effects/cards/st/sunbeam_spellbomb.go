package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sunbeam Spellbomb — Artifact {1} (Mirrodin, #5606):
//
//	"{W}, Sacrifice this artifact: You gain 5 life.
//	 {1}, Sacrifice this artifact: Draw a card."
//
// The first row gains its controller 5 life as it resolves.
// The second row is Aether Spellbomb's: {1}, sacrifice it, draw a card
// (mrdB5SpellbombDraw). The sacrifice is part of the cost, so the
// Spellbomb is gone before either row resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "014c94d8-2c39-4d33-902b-fa2398406fd5",
		Name:         "Sunbeam Spellbomb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, Sacrifice this artifact: You gain 5 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.SacrificeThis()),
				Effect:  effects.Do(effects.GainLife{Amount: 5}),
			},
			effects.MrdB5SpellbombDraw(),
		},
	})
}
