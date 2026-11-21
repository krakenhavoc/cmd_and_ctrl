package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bad River — Land:
//
//	"This land enters tapped.
//	 {T}, Sacrifice this land: Search your library for an Island or Swamp
//	 card, put it onto the battlefield, then shuffle."
//
// The fetch is the shared fetchDual search without the life payment;
// the land it finds enters untapped, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "259fc423-d078-4668-9e95-8c1a0d2e49ba",
		Name:         "Bad River",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for an Island or Swamp card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.FetchDual("island", "swamp"),
		}},
	})
}
