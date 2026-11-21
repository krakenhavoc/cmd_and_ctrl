package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grasslands — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}, Sacrifice this land: Search your library for a Forest or Plains
//	 card, put it onto the battlefield, then shuffle."
//
// A sacrifice-to-fetch land with no life cost: it enters tapped and the
// fetched land arrives untapped. It reads "card", not "basic land card", so a nonbasic land
// with either subtype qualifies. The shared body is fetchDual, minus
// the life payment the Zendikar fetchlands make.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e80bd454-8bc5-4921-90cf-6ad28a27a88b",
		Name:         "Grasslands",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for a Forest or Plains card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Effect:  effects.FetchDual("forest", "plains"),
		}},
	})
}
