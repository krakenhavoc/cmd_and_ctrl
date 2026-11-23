package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Navigation Orb — Artifact {3} (Commander Legends: Battle for Baldur's Gate,
// #5948):
//
//	"{2}, {T}, Sacrifice this artifact: Search your library for up to two
//	 basic land cards and/or Gate cards, reveal those cards, put one onto the
//	 battlefield tapped and the other into your hand, then shuffle."
//
// Cultivate's two-land search over "basic land or Gate" (TwoLandSearch): one
// card onto the battlefield tapped, then the other into your hand, then the
// shuffle, each revealed. The Orb is sacrificed as a cost, so the search
// resolves even if it is destroyed in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba7a46f1-c2a5-42d3-93eb-4dcea7186e5e",
		Name:         "Navigation Orb",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: Search your library for up to two basic land cards and/or Gate cards, reveal those cards, put one onto the battlefield tapped and the other into your hand, then shuffle.",
			Purpose: game.Purpose{Lands: 1, Tutors: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.TwoLandSearch(effects.NewContext(g, item), "Navigation Orb", "basic land or Gate card", navigationOrbLand, navigationOrbHand)
			},
		}},
	})
}

func navigationOrbLand(c game.Card) bool {
	return effects.IsBasicLand(c) || (c.IsLand() && c.HasSubtype("Gate"))
}

// navigationOrbHand is the hand half of the search. An on-disk key: never
// renamed, never reused.
var navigationOrbHand = effects.LandToHandTail("navigation-orb/land-to-hand", "Navigation Orb", "basic land or Gate card", navigationOrbLand)
