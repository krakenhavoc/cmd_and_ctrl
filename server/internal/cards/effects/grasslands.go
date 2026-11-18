package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "e80bd454-8bc5-4921-90cf-6ad28a27a88b",
		Name:         "Grasslands",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for a Forest or Plains card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    Plus(TapCost(), SacrificeThis()),
			Effect:  fetchDual("forest", "plains"),
		}},
	})
}
