package cardsth

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thief of Blood — Creature — Vampire {4}{B}{B}, 1/1 (Commander 2015, S82 OLD
// #5546):
//
//	"Flying
//	 As this creature enters, remove all counters from all permanents. This
//	 creature enters with a +1/+1 counter on it for each counter removed this
//	 way."
//
// A self-replacement on entry (CR 614.1c): as the Thief enters, every counter
// of every kind comes off every permanent on the battlefield, and the number
// removed is the number of +1/+1 counters it enters with. The Thief is not on
// the battlefield yet, so it keeps what it enters with. Counters on players are
// not on permanents and are left alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "97d61346-bd53-4eb8-a920-6ae0382eb20d",
		Name:            "Thief of Blood",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Replacements: []game.ReplacementEffect{{
			RowLabel:  "As it enters, remove all counters from all permanents; it enters with a +1/+1 counter on it for each counter removed this way.",
			EntryOnly: true,
			Watches:   []game.EventKind{game.EventZoneMove},
			AppliesTo: func(ev *game.ReplacementEvent, _ *game.Game, src *game.Card) bool {
				return ev.Kind == game.RepEventMove && ev.NewZone == game.ZoneBattlefield && src != nil && ev.CardID == src.InstanceID
			},
			Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
				removed := 0
				for _, c := range g.BattlefieldCardsForEffect() {
					kinds := make([]string, 0, len(c.Counters))
					for kind, n := range c.Counters {
						if n > 0 {
							kinds = append(kinds, kind)
						}
					}
					sort.Strings(kinds)
					for _, kind := range kinds {
						n := c.Counters[kind]
						if err := g.AddCounterForEffect(c.InstanceID, kind, -n); err != nil {
							return err
						}
						removed += n
					}
				}
				if removed > 0 {
					ev.AddCounterAtETB(game.CounterPlusOne, removed)
				}
				return nil
			},
		}},
	})
}
