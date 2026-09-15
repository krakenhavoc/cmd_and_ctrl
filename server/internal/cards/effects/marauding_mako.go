package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marauding Mako — 1/1 Creature — Shark Pirate for {R}:
//
//	"Whenever you discard one or more cards, put that many +1/+1
//	counters on this creature."
//
// A one-drop that grows every time the deck loots. Cycling {2} is an
// alternative cast path (S29) and isn't modelled.
//
// Batching: one EventDiscardCard per card, so a two-card discard
// puts two counters on via two triggers rather than one trigger for
// two. Same board state.
func init() {
	Register(Spec{
		OracleID:     "e349be42-5f14-44a9-9608-281985c10e2d",
		Name:         "Marauding Mako",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Cycling isn't implemented — the card can only be cast, not cycled."},
		Triggered: []game.TriggeredAbility{
			On(game.EventDiscardCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return discardedByYou(ev, source)
			}, "Marauding Mako — +1/+1 counter", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return AddCounter{Target: ctx.Source(), Kind: "+1/+1", N: 1}.Apply(ctx)
			}),
		},
	})
}
