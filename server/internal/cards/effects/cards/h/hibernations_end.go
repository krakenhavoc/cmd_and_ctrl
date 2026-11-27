package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hibernation's End — Enchantment {4}{G} (Coldsnap, S82 OLD b558):
//
//	"Cumulative upkeep {1}
//	 Whenever you pay this enchantment's cumulative upkeep, you may search your
//	 library for a creature card with mana value equal to the number of age
//	 counters on this enchantment, put it onto the battlefield, then shuffle."
//
// The mana value is the Enchantment's age counters as the trigger resolves
// (live, or last-known if it has gone, CR 608.2h). The search is optional, and
// the library is shuffled once it is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cf032b80-00d5-42fc-9e1a-16072b5fed55",
		Name:         "Hibernation's End",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Hibernation's End — cumulative upkeep {1}", "{1}"),
			effects.WhenCumulativeUpkeepIsPaid("Hibernation's End — you may search your library for a creature card with mana value equal to the number of age counters on it, put it onto the battlefield, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.AgeCountersAsItLeft(g, item)
					return effects.SearchLibrary{
						Player: item.Controller,
						Predicate: func(c game.Card) bool {
							return c.IsCreature() && c.ManaValue() == n
						},
						Dest:     game.ZoneBattlefield,
						Limit:    1,
						Shuffle:  true,
						Optional: true,
						Reason:   "Hibernation's End — a creature card with mana value " + effects.NumberWord(n),
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
