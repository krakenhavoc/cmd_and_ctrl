package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Forced Fruition — Enchantment {4}{U}{U} (EDHREC rank 2510):
//
//	"Whenever an opponent casts a spell, that player draws seven
//	 cards."
//
// The mill deck's enchantment that never mills: every spell an
// opponent casts hands them seven cards, and a library empties fast.
// One trigger per spell cast by an opponent (b15OpponentCastSpell),
// the caster captured at trigger time, and seven ordinary draws — so
// a draw payoff sees each one and an empty library flags the loss
// at the next state check, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "448b27a5-7c0c-4ab6-bddf-bd62b920aacc",
		Name:         "Forced Fruition",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15OpponentCastSpell(ev, source)
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				caster := ev.Actor
				return game.NewTriggeredItem(source, "Forced Fruition — that player draws seven cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: caster, N: 7}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
