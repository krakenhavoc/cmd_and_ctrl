package cardssy

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "1a85ba2b-ae10-4917-954a-7709b75a9740",
		Name:         "Tainted Pact",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return taintedPactStep(ctx, nil)
		},
	})
	// The names already exiled this way ride the prompt as its word, one per
	// line, and the card just exiled as its carried ID: the repeat is data on
	// the prompt, so a table waiting on the question is a restore point
	// (#4298). Registered in init, because the "no" repeats the step.
	TaintedPactAnswer = effects.MayYesNo("confirm/tainted-pact-take",
		func(ctx *effects.Context, r game.ConfirmAnswer) error {
			id := effects.CarryAt(r, 0)
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
				return nil
			}
			return ctx.Game.BounceToHandForEffect(id)
		},
		func(ctx *effects.Context, r game.ConfirmAnswer) error {
			return taintedPactStep(ctx, strings.Split(r.Word, "\n"))
		})
}

// taintedPactAnswer is the answer to "put that card into your hand?".
var TaintedPactAnswer game.ConfirmThen

// taintedPactStep exiles the next card and asks about it. `seen` is the
// names of the cards already exiled this way.
func taintedPactStep(ctx *effects.Context, seen []string) error {
	player := ctx.Controller()
	return effects.MillToZone{
		Player:   player,
		N:        1,
		To:       game.ZoneExile,
		ThenTail: effects.WithWord(ctx.ItemTail(taintedPactMilled), strings.Join(seen, "\n")),
	}.Apply(ctx)
}

// taintedPactMilled is the clause after the mill (#4253 slice 9).
var taintedPactMilled = effects.CardTail("tainted-pact/after-mill", func(ctx *effects.Context, o game.TailOutcome) error {
	player := ctx.Controller()
	var seen []string
	if o.Tail.Word != "" {
		seen = strings.Split(o.Tail.Word, "\n")
	}
	exiled := o.IDs

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
	return effects.MayChoice{
		Player:   player,
		Question: "Tainted Pact — put " + name + " into your hand?",
		YesLabel: "Put it into my hand",
		NoLabel:  "Exile the next card",
		Then:     TaintedPactAnswer,
		Carry:    []uuid.UUID{id},
		Word:     strings.Join(next, "\n"),
	}.Apply(ctx)
})
