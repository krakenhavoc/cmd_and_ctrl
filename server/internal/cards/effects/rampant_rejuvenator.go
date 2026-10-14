package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rampant Rejuvenator — Creature — Plant Hydra {3}{G}, 0/0:
//
//	"This creature enters with two +1/+1 counters on it.
//	 When this creature dies, search your library for up to X basic
//	 land cards, where X is this creature's power, put them onto the
//	 battlefield, then shuffle."
//
// The counters are an entry replacement (b10EntersWithCounters). X is
// the power the creature had as it last existed on the battlefield
// (CR 608.2h, its departure record): counters, anthems and pumps
// included. A power of zero or less finds nothing, and the library is
// still shuffled. The lands enter untapped, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b5f5d6e3-e72e-40a5-a5c7-410617b6f933",
		Name:         "Rampant Rejuvenator",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			b10EntersWithCounters(game.CounterPlusOne, 2, "Rampant Rejuvenator enters with two +1/+1 counters"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Rampant Rejuvenator — search for up to X basic lands",
				func(g *game.Game, item *game.StackItem) error {
					x := 0
					if info, ok := g.LastKnownPermanentForEffect(item.SourceCardID); ok {
						x = info.Power
					}
					match := IsBasicLand
					if x <= 0 {
						match = func(game.Card) bool { return false }
					}
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: match,
						Dest:      game.ZoneBattlefield,
						Limit:     x,
						Shuffle:   true,
						Reason:    "Rampant Rejuvenator — up to X basic lands onto the battlefield",
						Source:    item.SourceCardID,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
