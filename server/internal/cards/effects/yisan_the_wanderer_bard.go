package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yisan, the Wanderer Bard — Legendary Creature — Human Rogue Bard {2}{G},
// 2/3 (Commander Masters):
//
//	"{2}{G}, {T}, Put a verse counter on Yisan: Search your library for a
//	 creature card with mana value equal to the number of verse counters on
//	 Yisan, put it onto the battlefield, then shuffle."
//
// The counter is a cost (Mazemind Tome's), so the number a search looks for
// is the count after the payment; it is read as the ability resolves, from
// Yisan if she is still there and as she last existed if she is not
// (CR 608.2h, Context.SourcePermanent). A search that finds nothing at that
// mana value still shuffles. The ability is not sorcery speed and the {T}
// needs Yisan to have been under her controller's control since the turn
// began.
//
// No simplification.
func init() {
	const verse = "verse"
	Register(Spec{
		OracleID:     "7f4b0233-a2b0-41a2-8448-2df1f845148c",
		Name:         "Yisan, the Wanderer Bard",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Activated: []ActivatedAbility{{
			Label:   "{2}{G}, {T}, Put a verse counter on Yisan: Search your library for a creature card with mana value equal to the number of verse counters on Yisan, put it onto the battlefield, then shuffle.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tutors: 1},
			Cost:    Plus(ManaCost("{2}{G}"), TapCost(), AddCounterToThis(verse, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				info, ok := ctx.SourcePermanent()
				if !ok {
					return nil
				}
				want := info.Counters[verse]
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && c.ManaValue() == want
					},
					Dest:    game.ZoneBattlefield,
					Limit:   1,
					Shuffle: true,
					Reason:  "Yisan, the Wanderer Bard — a creature card with mana value " + strconv.Itoa(want),
				}.Apply(ctx)
			},
		}},
	})
}
