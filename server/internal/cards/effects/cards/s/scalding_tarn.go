package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scalding Tarn — Land (EDHREC rank 49):
//
//	"{T}, Pay 1 life, Sacrifice this land: Search your library for
//	an Island or Mountain card, put it onto the battlefield, then shuffle."
//
// One of the ten Zendikar / Onslaught fetchlands. The whole cycle is
// one card printed ten times with different subtypes, so the cost,
// the predicate and the search all live in fetchland_helpers.go and
// each file is just its own name, oracle ID and pair of types.
//
// The fetched land arrives UNTAPPED — that, plus the shuffle, is
// what the life is paying for. Evolving Wilds fetches tapped and
// costs no life, which is the entire difference between the two
// families.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb027150-848c-4a66-88ad-e20222304dd8",
		Name:         "Scalding Tarn",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay 1 life, Sacrifice this land: Search your library for an Island or Mountain card, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Lands: 1, LandsUntapped: 1},
			Cost:    effects.FetchlandCost(),
			Effect:  effects.FetchDual("island", "mountain"),
		}},
	})
}
