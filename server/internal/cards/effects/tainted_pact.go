package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tainted Pact — Instant {1}{B} (EDHREC rank 887):
//
//	"Exile the top card of your library. You may put that card into
//	 your hand unless it has the same name as another card exiled this
//	 way. Repeat this process until you put a card into your hand or you
//	 exile two cards with the same name, whichever comes first."
//
// One card at a time, each through the CR 614 window (MillToZone to
// exile, one card), then a yes/no for that card. Saying yes moves it
// from exile to your hand and ends the process; saying no exiles the
// next card. A card whose name matches one already exiled this way
// can't be taken and ends the process there. An empty library ends it
// too, with nothing taken: the deck-out happens the next time you draw,
// which is how Tainted Pact is used to win with a Laboratory Maniac
// effect.
//
// The cards are exiled face up: the whole table sees each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1a85ba2b-ae10-4917-954a-7709b75a9740",
		Name:         "Tainted Pact",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return taintedPactStep(ctx, nil)
		},
	})
}

// taintedPactStep exiles the next card and asks about it. `seen` is the
// names of the cards already exiled this way.
func taintedPactStep(ctx *Context, seen []string) error {
	player := ctx.Controller()
	return MillToZone{
		Player: player,
		N:      1,
		To:     game.ZoneExile,
		Then: func(ctx *Context, exiled []uuid.UUID) error {
			if len(exiled) == 0 {
				// The library is empty, or the card went somewhere a
				// replacement sent it: nothing was exiled this way, and
				// there is nothing left to repeat on.
				return nil
			}
			id := exiled[0]
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			name := strings.TrimSpace(c.Name)
			for _, n := range seen {
				if strings.EqualFold(n, name) {
					return nil // two cards with the same name: the process ends
				}
			}
			next := append(append([]string(nil), seen...), name)
			item := ctx.Item
			ctx.Game.QueueConfirmForEffect(game.ConfirmPrompt{
				Chooser:      player,
				Source:       ctx.Source(),
				Question:     "Tainted Pact — put " + name + " into your hand?",
				AcceptLabel:  "Put it into my hand",
				DeclineLabel: "Exile the next card",
				OnAccept: func(g *game.Game) error {
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
						return nil
					}
					return g.BounceToHandForEffect(id)
				},
				OnDecline: func(g *game.Game) error {
					return taintedPactStep(NewContext(g, item), next)
				},
			})
			return nil
		},
	}.Apply(ctx)
}
