package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gem of Becoming — Artifact {3} (Magic 2013):
//
//	"{3}, {T}, Sacrifice this artifact: Search your library for an Island card,
//	 a Swamp card, and a Mountain card. Reveal those cards, put them into your
//	 hand, then shuffle."
//
// One search for up to one card of each subtype, with a set rule that each pick
// fills a different slot (Aang's Journey's shape): a land that is both an
// Island and a Mountain fills only one of them, and three picks must cover
// Island, Swamp and Mountain one apiece. Fewer may be found, since a library
// can lack one (CR 701.23d), and the shuffle happens either way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b1e4940-11ac-4d85-98f4-3f50d1060b96",
		Name:         "Gem of Becoming",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: Search your library for an Island card, a Swamp card, and a Mountain card. Reveal those cards, put them into your hand, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 3},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.HasSubtypeWord("Island") || c.HasSubtypeWord("Swamp") || c.HasSubtypeWord("Mountain")
					},
					Dest:    game.ZoneHand,
					Limit:   3,
					UpTo:    true,
					Reveal:  true,
					Shuffle: true,
					Rule:    gemOfBecomingOneOfEach,
					Reason:  "Gem of Becoming — an Island card, a Swamp card and a Mountain card",
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var gemOfBecomingTypes = []string{"Island", "Swamp", "Mountain"}

// gemOfBecomingOneOfEach is the search's set rule: the picks can be matched to
// distinct slots among Island, Swamp and Mountain, each pick having that
// subtype. An on-disk key: never renamed, never reused.
var gemOfBecomingOneOfEach = game.RegisterSearchRule("search/s82-old-b428-island-swamp-mountain", func(picked []game.Card, _ int) bool {
	if len(picked) > len(gemOfBecomingTypes) {
		return false
	}
	var fill func(i int, used int) bool
	fill = func(i int, used int) bool {
		if i == len(picked) {
			return true
		}
		for slot, t := range gemOfBecomingTypes {
			if used&(1<<slot) == 0 && picked[i].HasSubtypeWord(t) && fill(i+1, used|1<<slot) {
				return true
			}
		}
		return false
	}
	return fill(0, 0)
})
