package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Planar Portal — Artifact {6} (Eighth Edition, S82 older sets batch 618, #5607):
//
//	"{6}, {T}: Search your library for a card, put that card into your hand,
//	 then shuffle."
//
// Profane Tutor's search as an activated ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4b10505-2b7a-4099-aacb-ff2359b24310",
		Name:         "Planar Portal",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}, {T}: Search your library for a card, put that card into your hand, then shuffle.",
			Purpose: game.Purpose{Tutors: 1, Answers: game.AnswerValue}, // ADR 0142: a tutor answers nothing on the stack
			Cost:    effects.Plus(effects.ManaCost("{6}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player:  item.Controller,
					Dest:    game.ZoneHand,
					Limit:   1,
					Shuffle: true,
					Reason:  "Planar Portal — search your library for a card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
