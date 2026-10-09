package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Horizon Seeker — Creature — Human Warrior {2}{G}, 3/2:
//
//	"Boast — {1}{G}: Search your library for a basic land card, reveal it, put it into your hand,
//	 then shuffle. (Activate only if this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0610ab1-88d3-4502-84b7-3a08d0170167",
		Name:         "Horizon Seeker",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			Boast("{1}{G}: Search your library for a basic land card, reveal it, put it into your hand, then shuffle.",
				ManaCost("{1}{G}"),
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: IsBasicLand,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Horizon Seeker — a basic land card",
					}.Apply(ctx)
				}),
		},
	})
}
