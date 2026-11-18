package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flood Plain — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}, Sacrifice this land: Search your library for a Plains or Island
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
		OracleID:     "9043d8d5-b38a-406f-a44c-49f579c644f0",
		Name:         "Flood Plain",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this land: Search your library for a Plains or Island card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    Plus(TapCost(), SacrificeThis()),
			Effect:  fetchDual("plains", "island"),
		}},
	})
}
