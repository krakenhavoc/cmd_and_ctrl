package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// War of the Last Alliance — Enchantment — Saga {3}{W}:
//
//	"(As this Saga enters and after your draw step, add a lore
//	 counter. Sacrifice after III.)
//	 I, II — Search your library for a legendary creature card, reveal
//	     it, put it into your hand, then shuffle.
//	 III — Creatures you control gain double strike until end of turn.
//	     The Ring tempts you."
//
// The double strike goes to the creatures you control as chapter III
// resolves (CR 611.2c), before the tempt.
//
// No simplification.
func init() {
	tutor := func(g *game.Game, item *game.StackItem) error {
		return b06TutorToHand("War of the Last Alliance — a legendary creature card", func(c game.Card) bool {
			return c.IsLegendary() && c.IsCreature()
		})(item, NewContext(g, item))
	}
	Register(Spec{
		OracleID:     "e376702a-f56e-4e47-aa8c-3f80384d12a9",
		Name:         "War of the Last Alliance",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			ChapterTrigger(1, "War of the Last Alliance — search for a legendary creature card", tutor),
			ChapterTrigger(2, "War of the Last Alliance — search for a legendary creature card", tutor),
			ChapterTrigger(3, "War of the Last Alliance — double strike, then the Ring tempts you", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"double strike"},
					Label:    "War of the Last Alliance — double strike",
				}).Apply(ctx); err != nil {
					return err
				}
				return TheRingTemptsYou{}.Apply(ctx)
			}),
		},
	})
}
