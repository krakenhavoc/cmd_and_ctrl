package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yoshimaru, Ever Faithful — Legendary Creature — Dog {W}, 1/1
// (EDHREC rank 4303):
//
//	"Whenever another legendary permanent you control enters, put a
//	 +1/+1 counter on Yoshimaru.
//	 Partner (You can have two commanders if both have partner.)"
//
// A one-mana Partner commander that grows with the thing every legend
// deck is already doing. The reason it is a commander and not a
// two-drop is the mana cost: {W} means the second partner can be any
// colour, and Yoshimaru is the cheapest body a Partner pair can open
// on.
//
// "Another legendary PERMANENT", not creature: a legendary land, a
// legendary artifact, a Saga, a planeswalker — all of them feed the
// Dog. That is what makes the card work in a list that is mostly
// noncreature legends. Legendary is read off the EFFECTIVE type line,
// so a permanent something made legendary counts and a token copy
// printed "except it isn't legendary" does not — the same answer the
// legend rule itself gets.
//
// "Another" excludes Yoshimaru's own entry, by instance rather than by
// name: the event carries the entering card's ID and the source is the
// Dog, so a token copy of Yoshimaru entering really does trigger the
// original (before the legend rule takes one of them).
//
// Partner is a deck-construction rule and needs nothing on the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "963834c8-42df-4ee6-9b45-b9de88ce2eac",
		Name:         "Yoshimaru, Ever Faithful",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b41AnotherLegendaryPermanentYouControlEntered(ev, source, g)
			}, "Yoshimaru, Ever Faithful — put a +1/+1 counter on Yoshimaru",
				func(g *game.Game, item *game.StackItem) error {
					return AddCounter{
						Target: item.SourceCardID,
						Kind:   game.CounterPlusOne,
						N:      1,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
